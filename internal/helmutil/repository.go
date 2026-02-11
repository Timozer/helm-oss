package helmutil

import (
	"strings"
)

// Repository represents a Helm chart repository.
// It supports two modes:
// 1. Local repository mode: repositories added via helm repo add
// 2. Remote repository mode: direct oss:// URI access.
type Repository interface {
	// URL returns the repository URL.
	URL() string

	// IndexURL returns the URL of index.yaml.
	IndexURL() string

	// CacheFile returns the local cache file path.
	// Returns empty string for remote mode.
	CacheFile() string

	// ShouldUpdateCache returns whether the local cache should be updated.
	ShouldUpdateCache() bool
}

// LocalRepository implements Repository for local repositories.
// Uses repository configuration added via helm repo add.
type LocalRepository struct {
	entry RepoEntry
}

func (r *LocalRepository) URL() string {
	return r.entry.URL()
}

func (r *LocalRepository) IndexURL() string {
	return r.entry.IndexURL()
}

func (r *LocalRepository) CacheFile() string {
	return r.entry.CacheFile()
}

func (r *LocalRepository) ShouldUpdateCache() bool {
	return true // Local repository needs cache updates
}

// RemoteRepository implements Repository for remote repositories.
// Uses oss:// URI directly without relying on local configuration.
type RemoteRepository struct {
	uri string
}

func (r *RemoteRepository) URL() string {
	return r.uri
}

func (r *RemoteRepository) IndexURL() string {
	return IndexFileURL(r.uri)
}

func (r *RemoteRepository) CacheFile() string {
	return "" // Remote mode does not use cache
}

func (r *RemoteRepository) ShouldUpdateCache() bool {
	return false // Remote mode does not update cache
}

// NewRepository creates an appropriate Repository implementation based on the input.
// If repoOrURI starts with "oss://", returns RemoteRepository.
// Otherwise, looks up local repository configuration and returns LocalRepository.
func NewRepository(repoOrURI string) (Repository, error) {
	if strings.HasPrefix(repoOrURI, "oss://") {
		return &RemoteRepository{uri: repoOrURI}, nil
	}

	entry, err := LookupRepoEntry(repoOrURI)
	if err != nil {
		return nil, err
	}

	return &LocalRepository{entry: entry}, nil
}
