package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofrs/flock"
	"sigs.k8s.io/yaml"

	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
)

type repoAddOptions struct {
	name                 string
	url                  string
	username             string
	password             string
	passCredentialsAll   bool
	forceUpdate          bool
	allowDeprecatedRepos bool

	certFile              string
	keyFile               string
	caFile                string
	insecureSkipTLSverify bool

	repoFile  string
	repoCache string
}

func (o *repoAddOptions) run(logger *log.Logger, settings *cli.EnvSettings) error {
	// Ensure the file directory exists as it is required for file locking
	err := os.MkdirAll(filepath.Dir(o.repoFile), os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return err
	}

	// Acquire a file lock for process synchronization
	repoFileExt := filepath.Ext(o.repoFile)
	var lockPath string
	if len(repoFileExt) > 0 && len(repoFileExt) < len(o.repoFile) {
		lockPath = strings.TrimSuffix(o.repoFile, repoFileExt) + ".lock"
	} else {
		lockPath = o.repoFile + ".lock"
	}
	fileLock := flock.New(lockPath)
	lockCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	locked, err := fileLock.TryLockContext(lockCtx, time.Second)
	if err == nil && locked {
		defer fileLock.Unlock()
	}
	if err != nil {
		return err
	}

	b, err := os.ReadFile(o.repoFile)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	var f repo.File
	if err := yaml.Unmarshal(b, &f); err != nil {
		return err
	}

	c := repo.Entry{
		Name:                  o.name,
		URL:                   o.url,
		Username:              o.username,
		Password:              o.password,
		PassCredentialsAll:    o.passCredentialsAll,
		CertFile:              o.certFile,
		KeyFile:               o.keyFile,
		CAFile:                o.caFile,
		InsecureSkipTLSverify: o.insecureSkipTLSverify,
	}

	// Check if the repo name is legal
	if strings.Contains(o.name, "/") {
		return fmt.Errorf("repository name (%s) contains '/', please specify a different name without '/'", o.name)
	}

	// If the repo exists do one of two things:
	// 1. If the configuration for the name is the same continue without error
	// 2. When the config is different require --force-update
	if !o.forceUpdate && f.Has(o.name) {
		existing := f.Get(o.name)
		if c != *existing {
			// The input coming in for the name is different from what is already
			// configured. Return an error.
			return fmt.Errorf("repository name (%s) already exists, please specify a different name", o.name)
		}

		// The add is idempotent so do nothing
		logger.Printf("%q already exists with the same configuration, skipping\n", o.name)
		return nil
	}

	r, err := repo.NewChartRepository(&c, getter.All(settings))
	if err != nil {
		return err
	}

	if o.repoCache != "" {
		r.CachePath = o.repoCache
	}
	if _, err := r.DownloadIndexFile(); err != nil {
		return fmt.Errorf("looks like %q is not a valid chart repository or cannot be reached: %w", o.url, err)
	}

	f.Update(&c)

	if err := f.WriteFile(o.repoFile, 0o600); err != nil {
		return err
	}

	return nil
}
