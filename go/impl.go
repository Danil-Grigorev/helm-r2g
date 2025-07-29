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
)

type Helm struct{}

func init() {
	HelmCallImpl = Helm{}
}

// install implements DemoCall.
func (d Helm) install(req *HelmChartInstallRequest) (resp HelmChartInstallResponse) {
	settings := cli.New()
	settings.SetNamespace(req.ns)
	set(req.env.kube_config, &settings.KubeConfig)
	set(req.env.kube_context, &settings.KubeContext)
	set(req.env.kube_token, &settings.KubeToken)
	set(req.env.kube_ca_file, &settings.KubeCaFile)
	settings.KubeInsecureSkipTLSVerify = req.env.kube_insecure_skip_tls_verify

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

	if err := runInstall(context.TODO(), log.Default(), settings, install); err != nil {
		resp.err = append(resp.err, err.Error())

		return
	}

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

func runInstall(ctx context.Context, logger *log.Logger, settings *cli.EnvSettings, install Install) error {

	actionConfig, err := initActionConfig(settings, logger)
	if err != nil {
		return fmt.Errorf("failed to init action config: %w", err)
	}

	installClient := action.NewInstall(actionConfig)

	installClient.DryRunOption = "none"
	set(install.DryRunOption, &installClient.DryRunOption)

	installClient.ReleaseName = install.ReleaseName
	chartRef := install.ChartRef
	installClient.Wait = install.Wait
	installClient.Timeout = time.Duration(install.Timeout) * time.Second
	if install.Timeout == 0 {
		installClient.Timeout = 5 * time.Minute
	}
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
		return fmt.Errorf("failed to created registry client: %w", err)
	}
	installClient.SetRegistryClient(registryClient)

	chartPath, err := installClient.ChartPathOptions.LocateChart(chartRef, settings)
	if err != nil {
		return err
	}

	providers := getter.All(settings)

	chart, err := loader.Load(chartPath)
	if err != nil {
		return err
	}

	// Check chart dependencies to make sure all are present in /charts
	if chartDependencies := chart.Metadata.Dependencies; chartDependencies != nil {
		if err := action.CheckDependencies(chart, chartDependencies); err != nil {
			err = fmt.Errorf("failed to check chart dependencies: %w", err)
			if !installClient.DependencyUpdate {
				return err
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
				return err
			}
			// Reload the chart with the updated Chart.lock file.
			if chart, err = loader.Load(chartPath); err != nil {
				return fmt.Errorf("failed to reload chart after repo update: %w", err)
			}
		}
	}

	release, err := installClient.RunWithContext(ctx, chart, install.Values)
	if err != nil {
		return fmt.Errorf("failed to run install: %w", err)
	}

	logger.Printf("release created:\n%+v", *release)

	return nil
}

var helmDriver string = os.Getenv("HELM_DRIVER")

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
