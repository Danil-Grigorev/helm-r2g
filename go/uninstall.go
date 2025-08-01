package main

import (
	"fmt"
	"log"
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
)

type uninstall struct {
	ReleaseName         string
	DisableHooks        bool
	DryRun              bool
	IgnoreNotFound      bool
	KeepHistory         bool
	Wait                bool
	DeletionPropagation string
	Timeout             int64
	Description         string
}

func runUninstall(logger *log.Logger, settings *cli.EnvSettings, uninstall uninstall) (*release.UninstallReleaseResponse, error) {
	actionConfig, err := initActionConfig(settings, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to init action config: %w", err)
	}

	uninstallClient := action.NewUninstall(actionConfig)

	uninstallClient.DisableHooks = uninstall.DisableHooks
	uninstallClient.DryRun = uninstall.DryRun
	uninstallClient.IgnoreNotFound = uninstall.IgnoreNotFound
	uninstallClient.KeepHistory = uninstall.KeepHistory
	uninstallClient.Wait = uninstall.Wait
	uninstallClient.Timeout = time.Duration(uninstall.Timeout) * time.Second
	uninstallClient.Description = uninstall.Description
	uninstallClient.DeletionPropagation = uninstall.DeletionPropagation

	result, err := uninstallClient.Run(uninstall.ReleaseName)
	if err != nil {
		return result, fmt.Errorf("failed to run uninstall action: %w", err)
	}

	return result, nil
}
