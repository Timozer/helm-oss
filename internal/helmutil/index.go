package helmutil

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/repo"
	"k8s.io/helm/pkg/urlutil" // Note that this is from Helm v2 SDK because in Helm v3 this package is internal.
	"sigs.k8s.io/yaml"
)

// Index represents a Helm chart repository index.
type Index struct {
	index *repo.IndexFile
}

// NewIndex creates a new empty index.
func NewIndex() *Index {
	return &Index{index: repo.NewIndexFile()}
}

// LoadIndex loads index from the file.
func LoadIndex(fpath string) (*Index, error) {
	idx, err := repo.LoadIndexFile(fpath)
	if err != nil {
		return nil, err
	}
	return &Index{index: idx}, nil
}

// Add adds a chart to the index.
func (idx *Index) Add(metadata interface{}, filename, baseURL, digest string) error {
	md, ok := metadata.(*chart.Metadata)
	if !ok {
		return errors.New("metadata is not *chart.Metadata")
	}

	if err := idx.index.MustAdd(md, filename, baseURL, digest); err != nil {
		return fmt.Errorf("add file to the index: %v", err)
	}
	return nil
}

// AddOrReplace adds a chart to the index or replaces it if it already exists.
func (idx *Index) AddOrReplace(metadata interface{}, filename, baseURL, digest string) error {
	// TODO: this looks like a workaround.
	// Think how we can rework this in the future.
	// Ref: https://github.com/kubernetes/helm/issues/3230

	// TODO: this code is the same as for Helm v2, only chart.Medata struct is from Helm v3 SDK.
	// We probably should reduce duplicate code .

	md, ok := metadata.(*chart.Metadata)
	if !ok {
		return errors.New("metadata is not *chart.Metadata")
	}

	u := filename
	if baseURL != "" {
		var err error
		_, file := filepath.Split(filename)
		u, err = urlutil.URLJoin(baseURL, file)
		if err != nil {
			u = filepath.Join(baseURL, file)
		}
	}
	cr := &repo.ChartVersion{
		URLs:     []string{u},
		Metadata: md,
		Digest:   digest,
		Created:  time.Now(),
	}

	// If no chart with such name exists in the index, just create a new
	// list of versions.
	entry, ok := idx.index.Entries[md.Name]
	if !ok {
		idx.index.Entries[md.Name] = repo.ChartVersions{cr}
		return nil
	}

	chartSemVer, err := semver.NewVersion(md.Version)
	if err != nil {
		return err
	}

	// If such version exists, replace it.
	for i, v := range entry {
		itemSemVer, err := semver.NewVersion(v.Version)
		if err != nil {
			return err
		}

		if chartSemVer.Equal(itemSemVer) {
			idx.index.Entries[md.Name][i] = cr
			return nil
		}
	}

	// Otherwise just add to the list of versions
	idx.index.Entries[md.Name] = append(entry, cr)
	return nil
}

// Delete removes a chart version from the index and returns its URL.
func (idx *Index) Delete(name, version string) (url string, err error) {
	for chartName, chartVersions := range idx.index.Entries {
		if chartName != name {
			continue
		}

		for i, chartVersion := range chartVersions {
			if chartVersion.Version == version {
				idx.index.Entries[chartName] = append(
					idx.index.Entries[chartName][:i],
					idx.index.Entries[chartName][i+1:]...,
				)
				if len(chartVersion.URLs) > 0 {
					return chartVersion.URLs[0], nil
				}
				return "", nil
			}
		}
	}

	return "", fmt.Errorf("chart %s version %s not found in index", name, version)
}

// Has checks if a chart version exists in the index.
func (idx *Index) Has(name, version string) bool {
	return idx.index.Has(name, version)
}

// SortEntries sorts the chart entries in the index.
func (idx *Index) SortEntries() {
	idx.index.SortEntries()
}

// UpdateGeneratedTime updates the generated timestamp of the index.
func (idx *Index) UpdateGeneratedTime() {
	idx.index.Generated = time.Now().UTC()
}

// MarshalBinary implements encoding.BinaryMarshaler.
func (idx *Index) MarshalBinary() (data []byte, err error) {
	return yaml.Marshal(idx.index)
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler.
func (idx *Index) UnmarshalBinary(data []byte) error {
	i := &repo.IndexFile{}
	if err := yaml.Unmarshal(data, i); err != nil {
		return err
	}
	i.SortEntries()

	*idx = Index{index: i}
	return nil
}

// Reader returns a reader for the index data.
func (idx *Index) Reader() (io.Reader, error) {
	b, err := idx.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(b), nil
}

// WriteFile writes the index to a file.
func (idx *Index) WriteFile(dest string, mode os.FileMode) error {
	return idx.index.WriteFile(dest, mode)
}
