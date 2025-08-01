package main

import (
	"cmp"
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

// registry_login implements HelmCall.
func (d Helm) registry_login(req *LoginRequest) (resp LoginResponse) {
	login := login{
		hostname:  req.hostname,
		username:  req.username,
		password:  req.password,
		certFile:  req.cert_file,
		keyFile:   req.key_file,
		caFile:    req.ca_file,
		insecure:  req.insecure,
		plainHTTP: req.plain_http,
	}

	if err := registryLogin(log.Default(), initSettings(req.env, ""), login); err != nil {
		resp.err = append(resp.err, err.Error())

		return
	}

	return
}

// install implements DemoCall.
func (d Helm) install(req *InstallRequest) (resp InstallResponse) {
	install := install{
		ReleaseName:     req.release_name,
		ChartRef:        req.chart,
		ChartVersion:    req.version,
		Wait:            req.wait,
		CreateNamespace: req.create_namespace,
		DryRunOption:    req.dry_run,
	}

	install.Timeout = get(req.timeout)

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
func (d Helm) upgrade(req *UpgradeRequest) (resp UpgradeResponse) {
	upgrade := upgrade{
		ReleaseName:  req.release_name,
		ChartRef:     req.chart,
		ChartVersion: req.version,
		Wait:         req.wait,
		DryRunOption: req.dry_run,
		ReuseValues:  req.reuse_values,
		ResetValues:  req.reset_values,
	}

	upgrade.Timeout = get(req.timeout)

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
func (d Helm) list(req *ListRequest) (resp ListResponse) {
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

// uninstall implements HelmCall.
func (d Helm) uninstall(req *UninstallRequest) (resp UninstallResponse) {
	settings := initSettings(req.env, req.ns)

	uninstall := uninstall{
		ReleaseName:         req.release_name,
		KeepHistory:         req.keep_history,
		DryRun:              req.dry_run,
		DisableHooks:        req.disable_hooks,
		IgnoreNotFound:      req.ignore_not_found,
		Wait:                req.wait,
		Description:         req.description,
		DeletionPropagation: req.deletion_propagation,
	}

	uninstall.Timeout = get(req.timeout)

	release, err := runUninstall(log.Default(), settings, uninstall)
	if err != nil {
		resp.err = append(resp.err, err.Error())

		return
	}

	data, err := json.Marshal(release)
	if err != nil {
		resp.err = append(resp.err, fmt.Errorf("failed to marshal release from uninstall: %w", err).Error())

		return
	}

	resp.data = string(data)

	return
}

// search implements HelmCall.
func (d Helm) repo_search(req *SearchRequest) (resp SearchResponse) {
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

// repo_add implements HelmCall.
func (d Helm) repo_add(req *AddRequest) (resp AddResponse) {
	settings := initSettings(req.env, "")

	add := repoAddOptions{
		name:                  req.name,
		url:                   req.url,
		username:              req.username,
		password:              req.password,
		passCredentialsAll:    req.pass_credentials_all,
		forceUpdate:           req.force_update,
		allowDeprecatedRepos:  req.allow_deprecated_repos,
		certFile:              req.cert_file,
		keyFile:               req.key_file,
		caFile:                req.ca_file,
		insecureSkipTLSverify: req.insecure_skip_tls_sverify,
		repoFile:              settings.RepositoryConfig,
		repoCache:             settings.RepositoryCache,
	}

	err := add.run(log.Default(), settings)
	if err != nil {
		resp.err = append(resp.err, err.Error())

		return
	}

	return
}

func get[T any](from []T) T {
	if len(from) > 0 {
		return from[0]
	}

	return *new(T)
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
	settings.KubeConfig = cmp.Or(get(env.kube_config), settings.KubeConfig)
	settings.KubeContext = cmp.Or(get(env.kube_context), settings.KubeContext)
	settings.KubeToken = cmp.Or(get(env.kube_token), settings.KubeToken)
	settings.KubeCaFile = cmp.Or(get(env.kube_ca_file), settings.KubeCaFile)
	settings.KubeInsecureSkipTLSVerify = settings.KubeInsecureSkipTLSVerify || env.kube_insecure_skip_tls_verify
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
