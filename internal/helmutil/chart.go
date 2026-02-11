// Package helmutil provides utilities for working with Helm charts and repositories.
package helmutil

import (
	"encoding/json"
	"fmt"
	"io"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
)

// Chart describes a helm chart.
type Chart interface {
	// Name returns chart name.
	// Example: "foo".
	Name() string

	// Version returns chart version.
	// Example: "0.1.0".
	Version() string

	// Metadata returns chart metadata.
	Metadata() ChartMetadata
}

// ChartV3 implements Chart in Helm v3.
type ChartV3 struct {
	chart *chart.Chart
}

// Name returns the chart name.
func (c ChartV3) Name() string {
	return c.chart.Name()
}

// Version returns the chart version.
func (c ChartV3) Version() string {
	if c.chart.Metadata == nil {
		return ""
	}
	return c.chart.Metadata.Version
}

// Metadata returns the chart metadata.
func (c ChartV3) Metadata() ChartMetadata {
	return &chartMetadataV3{meta: c.chart.Metadata}
}

// LoadChart returns chart loaded from the file system by path.
func LoadChart(fpath string) (Chart, error) {
	ch, err := loader.LoadFile(fpath)
	if err != nil {
		return ChartV3{}, fmt.Errorf("failed to load chart file: %s", err.Error())
	}
	return ChartV3{chart: ch}, nil
}

// LoadArchive returns chart loaded from the archive file reader.
func LoadArchive(r io.Reader) (Chart, error) {
	ch, err := loader.LoadArchive(r)
	if err != nil {
		return ChartV3{}, fmt.Errorf("failed to load chart archive: %s", err.Error())
	}
	return ChartV3{chart: ch}, nil
}

// ChartMetadata describes helm chart metadata.
type ChartMetadata interface {
	// MarshalJSON marshals chart metadata to JSON.
	MarshalJSON() ([]byte, error)

	// UnmarshalJSON unmarshals chart metadata from JSON.
	UnmarshalJSON([]byte) error

	// Value returns underlying chart metadata value.
	Value() interface{}
}

type chartMetadataV3 struct {
	meta *chart.Metadata
}

func (c *chartMetadataV3) MarshalJSON() ([]byte, error) {
	if c.meta == nil {
		return nil, nil
	}
	return json.Marshal(c.meta)
}

func (c *chartMetadataV3) UnmarshalJSON(b []byte) error {
	if c.meta == nil {
		c.meta = &chart.Metadata{}
	}
	return json.Unmarshal(b, c.meta)
}

func (c *chartMetadataV3) Value() interface{} {
	return c.meta
}

// NewChartMetadata creates a new ChartMetadata instance.
func NewChartMetadata() ChartMetadata {
	return &chartMetadataV3{meta: &chart.Metadata{}}
}
