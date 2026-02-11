package helmutil

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/repo"
)

// RepoEntry represents a Helm repository entry.
type RepoEntry interface {
	// Name returns repo name.
	// Example: "my-charts".
	Name() string

	// URL returns repo URL.
	// Examples:
	// - https://kubernetes-charts.storage.googleapis.com/
	// - s3://my-charts
	URL() string

	// IndexURL returns repo index file URL.
	// Examples:
	// - https://kubernetes-charts.storage.googleapis.com/index.yaml
	// - s3://my-charts/index.yaml
	IndexURL() string

	// CacheFile returns repo local cache file path.
	// Examples:
	// - /Users/foo/Library/Caches/helm/repository/my-charts-index.yaml (on macOS)
	// - /home/foo/.cache/helm/repository/my-charts-index.yaml (on Linux)
	CacheFile() string
}

// RepoEntryV3 implements RepoEntry in Helm v3.
type RepoEntryV3 struct {
	entry *repo.Entry
}

// Name returns the repository name.
func (r RepoEntryV3) Name() string {
	return r.entry.Name
}

// URL returns the repository URL.
func (r RepoEntryV3) URL() string {
	return r.entry.URL
}

// IndexURL returns the repository index file URL.
func (r RepoEntryV3) IndexURL() string {
	return IndexFileURL(r.entry.URL)
}

// CacheFile returns the local cache file path for this repository.
func (r RepoEntryV3) CacheFile() string {
	return filepath.Join(cacheDirPathV3(), repoCacheFileName(r.entry.Name))
}

// LookupRepoEntry returns an entry from helm's repositories.yaml file by name.
// If repositories.yaml file is not found, errors.Is(err, fs.ErrNotExist) will
// return true.
func LookupRepoEntry(name string) (RepoEntry, error) {
	repoFile, err := helm3LoadRepoFile(repoFilePathV3())
	if err != nil {
		return RepoEntryV3{}, fmt.Errorf("load repo file: %w", err)
	}

	entry := repoFile.Get(name)
	if entry == nil {
		return RepoEntryV3{}, errors.Errorf("repo with name %s not found, try `helm repo add %s <uri>`", name, name)
	}

	return RepoEntryV3{entry: entry}, nil
}

// LookupRepoEntryByURL returns an entry from helm's repositories.yaml file by
// repo URL. If not found, returns false and <nil> error.
// If repositories.yaml file is not found, errors.Is(err, fs.ErrNotExist) will
// return true.
func LookupRepoEntryByURL(url string) (RepoEntry, bool, error) {
	repoFile, err := helm3LoadRepoFile(repoFilePathV3())
	if err != nil {
		return RepoEntryV3{}, false, fmt.Errorf("load repo file: %w", err)
	}

	url = strings.TrimSuffix(url, "/")
	for _, entry := range repoFile.Repositories {
		entryURL := strings.TrimSuffix(entry.URL, "/")
		if url == entryURL {
			return RepoEntryV3{entry: entry}, true, nil
		}
	}

	return RepoEntryV3{}, false, nil
}
