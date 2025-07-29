package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/release"
)

type Helm struct{}

func init() {
	HelmCallImpl = Helm{}
}

// install implements DemoCall.
func (d Helm) install(req *HelmChartInstallRequest) (resp HelmChartInstallResponse) {
	install := Install{
		ReleaseName:     req.release_name,
		ChartRef:        req.chart,
		ChartVersion:    req.version,
		Wait:            req.wait,
		CreateNamespace: req.create_namespace,
		DryRunOption:    req.dry_run,
	}

	set(req.timeout, &install.Timeout)

	if len(req.values) > 0 {
		if err := json.Unmarshal(req.values, &install.Values); err != nil {
			resp.err = append(resp.err, err.Error())

			return
		}
	}

	release, err := runInstall(context.TODO(), log.Default(), initSettings(req.env), install)
	if err != nil {
		resp.err = append(resp.err, err.Error())

		return
	}

	data, err := json.Marshal(release)
	if err != nil {
		resp.err = append(resp.err, fmt.Errorf("failed to marshal release from install: %w", err).Error())

		return
	}

	resp.data = string(data)

	return
}

// list implements HelmCall.
func (d Helm) list(req *HelmChartListRequest) (resp HelmChartListResponse) {
	logger := log.Default()

	actionConfig, err := initActionConfigList(initSettings(req.env), logger, req.all_namespaces)
	if err != nil {
		resp.err = append(resp.err, fmt.Errorf("failed to init action config: %w", err).Error())

		return
	}

	listClient := action.NewList(actionConfig)
	listClient.AllNamespaces = req.all_namespaces

	listClient.Sort = action.Sorter(req.sort)
	listClient.StateMask = action.ListStates(req.state_mask)
	listClient.SetStateMask()

	listClient.ByDate = req.by_date
	listClient.SortReverse = req.sort_reverse
	listClient.Limit = int(req.limit)
	listClient.Offset = int(req.offset)
	listClient.Filter = req.filter
	listClient.NoHeaders = req.no_headers
	listClient.TimeFormat = req.time_format
	listClient.Uninstalled = req.uninstalled
	listClient.Superseded = req.superseded
	listClient.Uninstalling = req.uninstalling
	listClient.Deployed = req.deployed
	listClient.Failed = req.failed
	listClient.Pending = req.pending
	listClient.Selector = req.selector

	releases, err := runList(logger, listClient)
	if err != nil {
		resp.err = append(resp.err, err.Error())

		return
	}

	data, err := json.Marshal(releases)
	if err != nil {
		resp.err = append(resp.err, fmt.Errorf("failed to marshal releases from list: %w", err).Error())

		return
	}

	resp.data = string(data)

	return
}

func set[T any](from []T, to *T) {
	if len(from) > 0 {
		*to = from[0]
	}
}

type Install struct {
	ReleaseName     string
	ChartRef        string
	ChartVersion    string
	Wait            bool
	Timeout         int64
	CreateNamespace bool
	DryRunOption    []string
	Values          map[string]interface{}
}

func runInstall(ctx context.Context, logger *log.Logger, settings *cli.EnvSettings, install Install) (*release.Release, error) {
	actionConfig, err := initActionConfig(settings, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to init action config: %w", err)
	}

	installClient := action.NewInstall(actionConfig)

	installClient.DryRunOption = "none"
	set(install.DryRunOption, &installClient.DryRunOption)

	installClient.ReleaseName = install.ReleaseName
	chartRef := install.ChartRef
	installClient.Wait = install.Wait
	installClient.Timeout = time.Duration(install.Timeout) * time.Second
	installClient.CreateNamespace = install.CreateNamespace
	installClient.Namespace = settings.Namespace()
	installClient.Version = install.ChartVersion

	registryClient, err := newRegistryClientTLS(
		settings,
		logger,
		installClient.CertFile,
		installClient.KeyFile,
		installClient.CaFile,
		installClient.InsecureSkipTLSverify,
		installClient.PlainHTTP)
	if err != nil {
		return nil, fmt.Errorf("failed to created registry client: %w", err)
	}
	installClient.SetRegistryClient(registryClient)

	chartPath, err := installClient.ChartPathOptions.LocateChart(chartRef, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to locate chart: %w", err)
	}

	providers := getter.All(settings)

	chart, err := loader.Load(chartPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load chart: %w", err)
	}

	// Check chart dependencies to make sure all are present in /charts
	if chartDependencies := chart.Metadata.Dependencies; chartDependencies != nil {
		if err := action.CheckDependencies(chart, chartDependencies); err != nil {
			err = fmt.Errorf("failed to check chart dependencies: %w", err)
			if !installClient.DependencyUpdate {
				return nil, err
			}

			manager := &downloader.Manager{
				Out:              logger.Writer(),
				ChartPath:        chartPath,
				Keyring:          installClient.ChartPathOptions.Keyring,
				SkipUpdate:       false,
				Getters:          providers,
				RepositoryConfig: settings.RepositoryConfig,
				RepositoryCache:  settings.RepositoryCache,
				Debug:            settings.Debug,
				RegistryClient:   installClient.GetRegistryClient(),
			}
			if err := manager.Update(); err != nil {
				return nil, fmt.Errorf("failed to update chart dependencies: %w", err)
			}
			// Reload the chart with the updated Chart.lock file.
			if chart, err = loader.Load(chartPath); err != nil {
				return nil, fmt.Errorf("failed to reload chart after repo update: %w", err)
			}
		}
	}

	release, err := installClient.RunWithContext(ctx, chart, install.Values)
	if err != nil {
		return nil, fmt.Errorf("failed to run install: %w", err)
	}

	return release, nil
}

func runList(logger *log.Logger, listClient *action.List) ([]*release.Release, error) {
	results, err := listClient.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run list action: %w", err)
	}

	return results, nil
}

var helmDriver string = os.Getenv("HELM_DRIVER")

func initSettings(env HelmEnv) *cli.EnvSettings {
	settings := cli.New()
	set(env.kube_config, &settings.KubeConfig)
	set(env.kube_context, &settings.KubeContext)
	set(env.kube_token, &settings.KubeToken)
	set(env.kube_ca_file, &settings.KubeCaFile)
	settings.KubeInsecureSkipTLSVerify = env.kube_insecure_skip_tls_verify

	return settings
}

func initActionConfig(settings *cli.EnvSettings, logger *log.Logger) (*action.Configuration, error) {
	return initActionConfigList(settings, logger, false)
}

func initActionConfigList(settings *cli.EnvSettings, logger *log.Logger, allNamespaces bool) (*action.Configuration, error) {

	actionConfig := new(action.Configuration)

	namespace := func() string {
		// For list action, you can pass an empty string instead of settings.Namespace() to list
		// all namespaces
		if allNamespaces {
			return ""
		}
		return settings.Namespace()
	}()

	if err := actionConfig.Init(
		settings.RESTClientGetter(),
		namespace,
		helmDriver,
		logger.Printf); err != nil {
		return nil, err
	}

	return actionConfig, nil
}

func newRegistryClient(settings *cli.EnvSettings, plainHTTP bool) (*registry.Client, error) {
	opts := []registry.ClientOption{
		registry.ClientOptDebug(settings.Debug),
		registry.ClientOptEnableCache(true),
		registry.ClientOptWriter(os.Stderr),
		registry.ClientOptCredentialsFile(settings.RegistryConfig),
	}
	if plainHTTP {
		opts = append(opts, registry.ClientOptPlainHTTP())
	}

	// Create a new registry client
	registryClient, err := registry.NewClient(opts...)
	if err != nil {
		return nil, err
	}
	return registryClient, nil
}

func newRegistryClientTLS(settings *cli.EnvSettings, logger *log.Logger, certFile, keyFile, caFile string, insecureSkipTLSverify, plainHTTP bool) (*registry.Client, error) {
	if certFile != "" && keyFile != "" || caFile != "" || insecureSkipTLSverify {
		registryClient, err := registry.NewRegistryClientWithTLS(
			logger.Writer(),
			certFile,
			keyFile,
			caFile,
			insecureSkipTLSverify,
			settings.RegistryConfig,
			settings.Debug)

		if err != nil {
			return nil, err
		}
		return registryClient, nil
	}
	registryClient, err := newRegistryClient(settings, plainHTTP)
	if err != nil {
		return nil, err
	}
	return registryClient, nil
}
