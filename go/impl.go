package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/release"
)

type Helm struct{}

func init() {
	HelmCallImpl = Helm{}
}

// install implements DemoCall.
func (d Helm) install(req *HelmChartInstallRequest) (resp HelmChartInstallResponse) {
	install := install{
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

	release, err := runInstall(context.TODO(), log.Default(), initSettings(req.env, req.ns), install)
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

// upgrade implements HelmCall.
func (d Helm) upgrade(req *HelmChartUpgradeRequest) (resp HelmChartUpgradeResponse) {
	upgrade := upgrade{
		ReleaseName:  req.release_name,
		ChartRef:     req.chart,
		ChartVersion: req.version,
		Wait:         req.wait,
		DryRunOption: req.dry_run,
		ReuseValues:  req.reuse_values,
		ResetValues:  req.reset_values,
	}

	set(req.timeout, &upgrade.Timeout)

	if len(req.values) > 0 {
		if err := json.Unmarshal(req.values, &upgrade.Values); err != nil {
			resp.err = append(resp.err, err.Error())

			return
		}
	}

	release, err := runUpgrade(context.TODO(), log.Default(), initSettings(req.env, req.ns), upgrade)
	if err != nil {
		resp.err = append(resp.err, err.Error())

		return
	}

	data, err := json.Marshal(release)
	if err != nil {
		resp.err = append(resp.err, fmt.Errorf("failed to marshal release from upgrade: %w", err).Error())

		return
	}

	resp.data = string(data)

	return
}

// list implements HelmCall.
func (d Helm) list(req *HelmChartListRequest) (resp HelmChartListResponse) {
	logger := log.Default()

	actionConfig, err := initActionConfigList(initSettings(req.env, req.ns), logger, req.all_namespaces)
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

	releases, err := runList(listClient)
	if err != nil {
		resp.err = append(resp.err, err.Error())

		return
	}

	data, err := json.Marshal(releases)
	if err != nil {
		resp.err = append(resp.err, fmt.Errorf("failed to marshal releases from list: %w", err).Error())

		return
	}

	if len(releases) > 0 {
		resp.data = string(data)
	}

	return
}

// search implements HelmCall.
func (d Helm) repo_search(req *HelmChartSearchRequest) (resp HelmChartSearchResponse) {
	settings := initSettings(req.env, "")

	search := searchRepoOptions{
		terms:        req.terms,
		version:      req.version,
		versions:     req.versions,
		regexp:       req.regexp,
		devel:        req.devel,
		repoFile:     settings.RepositoryConfig,
		repoCacheDir: settings.RepositoryCache,
	}

	searchResult, err := search.run()
	if err != nil {
		resp.err = append(resp.err, err.Error())

		return
	}

	data, err := json.Marshal(searchResult)
	if err != nil {
		resp.err = append(resp.err, fmt.Errorf("failed to marshal releases from list: %w", err).Error())

		return
	}

	if len(searchResult) > 0 {
		resp.data = string(data)
	}

	return
}

func set[T any](from []T, to *T) {
	if len(from) > 0 {
		*to = from[0]
	}
}

func runList(listClient *action.List) ([]*release.Release, error) {
	results, err := listClient.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run list action: %w", err)
	}

	return results, nil
}

var helmDriver string = os.Getenv("HELM_DRIVER")

func initSettings(env HelmEnv, namespace string) *cli.EnvSettings {
	settings := cli.New()
	set(env.kube_config, &settings.KubeConfig)
	set(env.kube_context, &settings.KubeContext)
	set(env.kube_token, &settings.KubeToken)
	set(env.kube_ca_file, &settings.KubeCaFile)
	settings.KubeInsecureSkipTLSVerify = env.kube_insecure_skip_tls_verify
	settings.SetNamespace(namespace)

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
