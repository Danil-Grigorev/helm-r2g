package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"
)

type install struct {
	ReleaseName     string
	ChartRef        string
	ChartVersion    string
	Wait            bool
	Timeout         int64
	CreateNamespace bool
	DryRunOption    []string
	Values          map[string]interface{}
}

func runInstall(ctx context.Context, logger *log.Logger, settings *cli.EnvSettings, install install) (*release.Release, error) {
	actionConfig, err := initActionConfig(settings, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to init action config: %w", err)
	}

	installClient := action.NewInstall(actionConfig)

	installClient.DryRunOption = "none"
	installClient.DryRunOption = get(install.DryRunOption)

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
