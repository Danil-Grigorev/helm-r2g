package main

import (
	"fmt"
	"log"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
)

type login struct {
	hostname  string
	username  string
	password  string
	certFile  string
	keyFile   string
	caFile    string
	insecure  bool
	plainHTTP bool
}

func registryLogin(logger *log.Logger, settings *cli.EnvSettings, o login) error {
	actionConfig, err := initActionConfig(settings, logger)
	if err != nil {
		return fmt.Errorf("failed to init action config: %w", err)
	}

	actionConfig.RegistryClient, err = newRegistryClientTLS(
		settings,
		logger,
		o.certFile,
		o.keyFile,
		o.caFile,
		o.insecure,
		o.plainHTTP)
	if err != nil {
		return fmt.Errorf("failed to created registry client: %w", err)
	}

	return action.NewRegistryLogin(actionConfig).Run(nil, o.hostname, o.username, o.password,
		action.WithCertFile(o.certFile),
		action.WithKeyFile(o.keyFile),
		action.WithCAFile(o.caFile),
		action.WithInsecure(o.insecure),
		action.WithPlainHTTPLogin(o.plainHTTP))
}
