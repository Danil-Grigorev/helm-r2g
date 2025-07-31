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

type upgrade struct {
	ReleaseName  string
	ChartRef     string
	ChartVersion string
	Wait         bool
	Timeout      int64
	DryRunOption []string
	ReuseValues  bool
	ResetValues  bool
	Values       map[string]interface{}
}

func runUpgrade(ctx context.Context, logger *log.Logger, settings *cli.EnvSettings, upgrade upgrade) (*release.Release, error) {
	actionConfig, err := initActionConfig(settings, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to init action config: %w", err)
	}

	upgradeClient := action.NewUpgrade(actionConfig)

	upgradeClient.Namespace = settings.Namespace()
	upgradeClient.DryRunOption = "none"
	upgradeClient.Version = upgrade.ChartVersion
	upgradeClient.Wait = upgrade.Wait
	upgradeClient.ReuseValues = upgrade.ReuseValues
	upgradeClient.ResetValues = upgrade.ResetValues
	upgradeClient.Timeout = time.Duration(upgrade.Timeout) * time.Second
	set(upgrade.DryRunOption, &upgradeClient.DryRunOption)

	registryClient, err := newRegistryClientTLS(
		settings,
		logger,
		upgradeClient.CertFile,
		upgradeClient.KeyFile,
		upgradeClient.CaFile,
		upgradeClient.InsecureSkipTLSverify,
		upgradeClient.PlainHTTP)
	if err != nil {
		return nil, fmt.Errorf("missing registry client: %w", err)
	}
	upgradeClient.SetRegistryClient(registryClient)

	chartPath, err := upgradeClient.ChartPathOptions.LocateChart(upgrade.ChartRef, settings)
	if err != nil {
		return nil, err
	}

	providers := getter.All(settings)

	// Check chart dependencies to make sure all are present in /charts
	chart, err := loader.Load(chartPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load chart: %w", err)
	}
	if req := chart.Metadata.Dependencies; req != nil {
		if err := action.CheckDependencies(chart, req); err != nil {
			err = fmt.Errorf("failed to check chart dependencies: %w", err)
			if !upgradeClient.DependencyUpdate {
				return nil, err
			}

			man := &downloader.Manager{
				Out:              logger.Writer(),
				ChartPath:        chartPath,
				Keyring:          upgradeClient.ChartPathOptions.Keyring,
				SkipUpdate:       false,
				Getters:          providers,
				RepositoryConfig: settings.RepositoryConfig,
				RepositoryCache:  settings.RepositoryCache,
				Debug:            settings.Debug,
			}
			if err := man.Update(); err != nil {
				return nil, err
			}
			// Reload the chart with the updated Chart.lock file.
			if chart, err = loader.Load(chartPath); err != nil {
				return nil, fmt.Errorf("failed to reload chart after repo update: %w", err)
			}
		}
	}

	release, err := upgradeClient.RunWithContext(ctx, upgrade.ReleaseName, chart, upgrade.Values)
	if err != nil {
		return nil, fmt.Errorf("failed to run upgrade action: %w", err)
	}

	return release, nil
}
