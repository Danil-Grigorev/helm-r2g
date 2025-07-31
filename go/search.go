package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver/v3"

	"helm.sh/helm/v3/cmd/helm/search"
	"helm.sh/helm/v3/pkg/helmpath"
	"helm.sh/helm/v3/pkg/repo"
)

// searchMaxScore suggests that any score higher than this is not considered a match.
const searchMaxScore = 25

type searchRepoOptions struct {
	versions     bool
	regexp       string
	devel        bool
	version      string
	terms        []string
	repoFile     string
	repoCacheDir string
}

func (o *searchRepoOptions) run() ([]*search.Result, error) {
	o.setupSearchedVersion()

	index, err := o.buildIndex()
	if err != nil {
		return nil, err
	}

	var res []*search.Result
	if len(o.terms) == 0 {
		res = index.All()
	} else {
		if o.regexp != "" {
			res, err = index.SearchRegexp(o.regexp, searchMaxScore)
		} else {
			q := strings.Join(o.terms, " ")
			res = index.SearchLiteral(q, searchMaxScore)
		}

		if err != nil {
			return nil, err
		}
	}

	search.SortScore(res)
	data, err := o.applyConstraint(res)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (o *searchRepoOptions) setupSearchedVersion() {
	slog.Debug("original chart version", "version", o.version)

	if o.version != "" {
		return
	}

	if o.devel { // search for releases and prereleases (alpha, beta, and release candidate releases).
		slog.Debug("setting version to >0.0.0-0")
		o.version = ">0.0.0-0"
	} else { // search only for stable releases, prerelease versions will be skipped
		slog.Debug("setting version to >0.0.0")
		o.version = ">0.0.0"
	}
}

func (o *searchRepoOptions) applyConstraint(res []*search.Result) ([]*search.Result, error) {
	if o.version == "" {
		return res, nil
	}

	constraint, err := semver.NewConstraint(o.version)
	if err != nil {
		return res, fmt.Errorf("an invalid version/constraint format: %w", err)
	}

	data := res[:0]
	foundNames := map[string]bool{}
	for _, r := range res {
		// if not returning all versions and already have found a result,
		// you're done!
		if !o.versions && foundNames[r.Name] {
			continue
		}
		v, err := semver.NewVersion(r.Chart.Version)
		if err != nil {
			continue
		}
		if constraint.Check(v) {
			data = append(data, r)
			foundNames[r.Name] = true
		}
	}

	return data, nil
}

func (o *searchRepoOptions) buildIndex() (*search.Index, error) {
	// Load the repositories.yaml
	rf, err := repo.LoadFile(o.repoFile)
	if errors.Is(err, fs.ErrNotExist) || len(rf.Repositories) == 0 {
		return nil, errors.New("no repositories configured")
	}

	i := search.NewIndex()
	for _, re := range rf.Repositories {
		n := re.Name
		f := filepath.Join(o.repoCacheDir, helmpath.CacheIndexFile(n))
		ind, err := repo.LoadIndexFile(f)
		if err != nil {
			slog.Warn("repo is corrupt or missing", "repo", n, slog.Any("error", err))
			continue
		}

		i.AddRepo(n, ind, o.versions || len(o.version) > 0)
	}
	return i, nil
}
