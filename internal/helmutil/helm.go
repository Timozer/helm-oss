package helmutil

// This file contains helm helpers suitable for both v2 and v3.

import (
	"fmt"
	"strings"

	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/repo"
)

var (
	helm3Env *cli.EnvSettings

	// func that loads helm repo file.
	// Defined for testing purposes.
	helm3LoadRepoFile func(path string) (*repo.File, error)
)

func repoFilePathV3() string {
	return helm3Env.RepositoryConfig
}

func cacheDirPathV3() string {
	return helm3Env.RepositoryCache
}

// SetupHelm initializes the Helm environment settings.
func SetupHelm() {
	helm3Env = cli.New()
	helm3LoadRepoFile = repo.LoadFile
}

// IndexFileURL returns index file URL for the provided repository URL.
func IndexFileURL(repoURL string) string {
	return strings.TrimSuffix(repoURL, "/") + "/index.yaml"
}

func repoCacheFileName(name string) string {
	return fmt.Sprintf("%s-index.yaml", name)
}
