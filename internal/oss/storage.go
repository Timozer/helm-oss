package oss

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	"helm-oss/internal/helmutil"
	"sigs.k8s.io/yaml"
)

var (
	ErrBucketNotFound = errors.New("bucket not found")
	ErrObjectNotFound = errors.New("object not found")
)

const (
	// ossMetadataSoftLimitBytes is application-specific soft limit
	// for the number of bytes in OSS object metadata.
	ossMetadataSoftLimitBytes = 1900

	// metaChartMetadata is a oss object metadata key that represents chart metadata.
	metaChartMetadata = "chart-metadata"

	// metaChartDigest is a oss object metadata key that represents chart digest.
	metaChartDigest = "chart-digest"
)

type Config struct {
	Endpoint        string `json:"endpoint"`
	Region          string `json:"region"`
	AccessKeyID     string `json:"accessKeyID"`
	AccessKeySecret string `json:"accessKeySecret"`
	SessionToken    string `json:"sessionToken"`
}

type Storage struct {
	cfg    oss.Config
	client *oss.Client
}

type ChartInfo struct {
	Meta     helmutil.ChartMetadata
	Filename string
	Hash     string
}

// New returns a new Storage.
// It loads configuration from ~/.config/helm_plugin_oss.yaml if exists,
// and overrides it with environment variables.
func New() *Storage {
	// 1. Load config from file
	conf := &Config{}
	home, err := os.UserHomeDir()
	if err == nil {
		configPath := filepath.Join(home, ".config", "helm_plugin_oss.yaml")
		if data, err := os.ReadFile(configPath); err == nil {
			_ = yaml.Unmarshal(data, conf)
		}
	}

	// 2. Override with Env Vars
	if v := os.Getenv("HELM_OSS_ENDPOINT"); v != "" {
		conf.Endpoint = v
	}
	if v := os.Getenv("HELM_OSS_REGION"); v != "" {
		conf.Region = v
	}
	if v := os.Getenv("HELM_OSS_ACCESS_KEY_ID"); v != "" {
		conf.AccessKeyID = v
	}
	if v := os.Getenv("HELM_OSS_ACCESS_KEY_SECRET"); v != "" {
		conf.AccessKeySecret = v
	}
	if v := os.Getenv("HELM_OSS_SESSION_TOKEN"); v != "" {
		conf.SessionToken = v
	}

	// 3. Create Credentials Provider
	provider := credentials.NewStaticCredentialsProvider(
		conf.AccessKeyID,
		conf.AccessKeySecret,
		conf.SessionToken,
	)

	// 4. Create OSS Config
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(provider).
		WithRegion(conf.Region).
		WithEndpoint(conf.Endpoint)

	// 5. Create Client
	client := oss.NewClient(cfg)

	return &Storage{
		cfg:    *cfg,
		client: client,
	}
}

// Traverse traverses all charts in the repository.
func (s *Storage) Traverse(ctx context.Context, repoURI string) (<-chan ChartInfo, <-chan error) {
	charts := make(chan ChartInfo, 1)
	errs := make(chan error, 1)
	go s.traverse(ctx, repoURI, charts, errs)
	return charts, errs
}

// traverse traverses all charts in the repository.
// It writes an info item about every chart to items, and errors to errs.
// It always closes both channels when returns.
//
//nolint:funcorder // Keep traverse near Traverse for better code organization
func (s *Storage) traverse(ctx context.Context, repoURI string, items chan<- ChartInfo, errs chan<- error) {
	defer close(items)
	defer close(errs)

	bucket, prefixKey, err := parseURI(repoURI)
	if err != nil {
		errs <- err
		return
	}

	var continuationToken *string
	for {
		listOut, err := s.client.ListObjectsV2(ctx, &oss.ListObjectsV2Request{
			Bucket:            oss.Ptr(bucket),
			Prefix:            oss.Ptr(prefixKey),
			ContinuationToken: continuationToken,
		})
		if err != nil {
			errs <- fmt.Errorf("list oss bucket objects: %w", err)
			return
		}

		for _, obj := range listOut.Contents {
			// We need to make object key relative to repo root.
			key := strings.TrimPrefix(*obj.Key, prefixKey)
			// Additionally trim prefix slash if exists
			key = strings.TrimPrefix(key, "/")

			if strings.Contains(key, "/") {
				// This is a subfolder. Ignore it.
				continue
			}

			if !strings.HasSuffix(key, ".tgz") {
				// Ignore any file that isn't a chart
				continue
			}

			metaOut, err := s.client.HeadObject(ctx, &oss.HeadObjectRequest{
				Bucket: oss.Ptr(bucket),
				Key:    obj.Key,
			})
			if err != nil {
				errs <- fmt.Errorf("head oss object %q: %w", key, err)
				return
			}

			reindexItem := ChartInfo{Filename: key}

			// Try to get metadata with case-insensitivity handling
			// OSS might capitalize keys.
			serializedChartMeta := getMetadataValue(metaOut.Metadata, metaChartMetadata)
			chartDigest := getMetadataValue(metaOut.Metadata, metaChartDigest)

			if serializedChartMeta == "" || chartDigest == "" {
				// Metadata missing, fallback to downloading
				objectOut, err := s.client.GetObject(ctx, &oss.GetObjectRequest{
					Bucket: oss.Ptr(bucket),
					Key:    obj.Key,
				})
				if err != nil {
					errs <- fmt.Errorf("get oss object %q: %w", key, err)
					return
				}

				buf := &bytes.Buffer{}
				tr := io.TeeReader(objectOut.Body, buf)

				ch, err := helmutil.LoadArchive(tr)
				objectOut.Body.Close()
				if err != nil {
					errs <- fmt.Errorf("load archive from oss object %q: %w", key, err)
					return
				}

				digest, err := helmutil.Digest(buf)
				if err != nil {
					errs <- fmt.Errorf("get chart hash for %q: %w", key, err)
					return
				}

				reindexItem.Meta = ch.Metadata()
				reindexItem.Hash = digest
			} else {
				meta := helmutil.NewChartMetadata()
				if err := meta.UnmarshalJSON([]byte(serializedChartMeta)); err != nil {
					errs <- fmt.Errorf("unserialize chart meta for %q: %w", key, err)
					return
				}

				reindexItem.Meta = meta
				reindexItem.Hash = chartDigest
			}

			items <- reindexItem
		}

		if !listOut.IsTruncated {
			break
		}
		if listOut.NextContinuationToken == nil || *listOut.NextContinuationToken == "" {
			break
		}
		continuationToken = listOut.NextContinuationToken
	}
}

// FetchRaw downloads the object from URI and returns it in the form of byte slice.
// uri must be in the form of oss protocol: oss://bucket-name/key[...].
func (s *Storage) FetchRaw(ctx context.Context, uri string) ([]byte, error) {
	bucket, key, err := parseURI(uri)
	if err != nil {
		return nil, err
	}

	result, err := s.client.GetObject(ctx, &oss.GetObjectRequest{
		Bucket: oss.Ptr(bucket),
		Key:    oss.Ptr(key),
	})
	if err != nil {
		var serviceErr *oss.ServiceError
		if errors.As(err, &serviceErr) {
			if serviceErr.StatusCode == http.StatusNotFound || serviceErr.Code == "NoSuchKey" {
				return nil, ErrObjectNotFound
			}
			if serviceErr.Code == "NoSuchBucket" {
				return nil, ErrBucketNotFound
			}
		}
		return nil, fmt.Errorf("fetch object from oss: %w", err)
	}
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("read object body: %w", err)
	}

	return data, nil
}

// Exists returns true if an object exists in the storage.
func (s *Storage) Exists(ctx context.Context, uri string) (bool, error) {
	bucket, key, err := parseURI(uri)
	if err != nil {
		return false, err
	}

	_, err = s.client.HeadObject(ctx, &oss.HeadObjectRequest{
		Bucket: oss.Ptr(bucket),
		Key:    oss.Ptr(key),
	})
	if err != nil {
		var serviceErr *oss.ServiceError
		if errors.As(err, &serviceErr) {
			if serviceErr.StatusCode == http.StatusNotFound || serviceErr.Code == "NoSuchKey" || serviceErr.Code == "NotFound" {
				return false, nil
			}
		}
		// If custom error mapping is needed:
		if IsNotFound(err) {
			return false, nil
		}

		return false, fmt.Errorf("head object from oss: %w", err)
	}

	return true, nil
}

// IndexExists returns true if index file exists in the storage for repository
// with the provided uri.
// uri must be in the form of oss protocol: oss://bucket-name/key[...].
func (s *Storage) IndexExists(ctx context.Context, uri string) (bool, error) {
	if strings.HasPrefix(uri, "index.yaml") {
		return false, errors.New("uri must not contain \"index.yaml\" suffix, it appends automatically")
	}
	uri = helmutil.IndexFileURL(uri)

	return s.Exists(ctx, uri)
}

// PutIndex puts the index file to the storage.
// uri must be in the form of oss protocol: oss://bucket-name/key[...].
func (s *Storage) PutIndex(ctx context.Context, uri string, r io.Reader) error {
	if strings.HasPrefix(uri, "index.yaml") {
		return errors.New("uri must not contain \"index.yaml\" suffix, it appends automatically")
	}
	uri = helmutil.IndexFileURL(uri)

	bucket, key, err := parseURI(uri)
	if err != nil {
		return err
	}

	_, err = s.client.PutObject(ctx, &oss.PutObjectRequest{
		Bucket: oss.Ptr(bucket),
		Key:    oss.Ptr(key),
		Body:   r,
	})
	if err != nil {
		return fmt.Errorf("upload index to OSS bucket: %w", err)
	}

	return nil
}

// PutChart puts the chart file to the storage.
// uri must be in the form of oss protocol: oss://bucket-name/key[...].
func (s *Storage) PutChart(
	ctx context.Context,
	uri string,
	r io.Reader,
	chartMeta string,
	chartDigest string,
	contentType string,
	prov bool,
	provReader io.Reader,
) (string, error) {
	bucket, key, err := parseURI(uri)
	if err != nil {
		return "", err
	}

	_, err = s.client.PutObject(ctx, &oss.PutObjectRequest{
		Bucket:      oss.Ptr(bucket),
		Key:         oss.Ptr(key),
		Body:        r,
		ContentType: oss.Ptr(contentType),
		Metadata:    assembleObjectMetadata(chartMeta, chartDigest),
	})
	if err != nil {
		return "", fmt.Errorf("upload chart object to oss: %w", err)
	}

	if prov {
		_, err := s.client.PutObject(ctx, &oss.PutObjectRequest{
			Bucket: oss.Ptr(bucket),
			Key:    oss.Ptr(key + ".prov"),
			Body:   provReader,
		})
		if err != nil {
			return "", fmt.Errorf("upload prov object to oss: %w", err)
		}
	}

	return "", nil
}

// DeleteChart deletes the chart object by uri. Also deletes .prov file if exists.
// uri must be in the form of oss protocol: oss://bucket-name/key[...].
func (s *Storage) DeleteChart(ctx context.Context, uri string) error {
	bucket, key, err := parseURI(uri)
	if err != nil {
		return err
	}

	// Delete both chart and .prov file
	objects := []oss.DeleteObject{
		{Key: oss.Ptr(key)},
		{Key: oss.Ptr(key + ".prov")},
	}

	_, err = s.client.DeleteMultipleObjects(ctx, &oss.DeleteMultipleObjectsRequest{
		Bucket:  oss.Ptr(bucket),
		Objects: objects,
		Quiet:   true,
	})
	if err != nil {
		return fmt.Errorf("delete chart objects from OSS: %w", err)
	}

	return nil
}

func parseURI(uri string) (bucket, key string, err error) {
	if !strings.HasPrefix(uri, "oss://") {
		return "", "", fmt.Errorf("uri %s protocol is not oss", uri)
	}

	u, err := url.Parse(uri)
	if err != nil {
		return "", "", fmt.Errorf("parse uri %s: %w", uri, err)
	}

	bucket, key = u.Host, strings.TrimPrefix(u.Path, "/")
	return bucket, key, nil
}

// Helper to check if error is generic NotFound if SDK error type check fails.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrObjectNotFound) || errors.Is(err, ErrBucketNotFound)
}

// assembleObjectMetadata assembles and returns OSS object metadata.
// May return empty metadata if chart metadata is too big.
func assembleObjectMetadata(chartMeta, chartDigest string) map[string]string {
	meta := map[string]string{
		metaChartMetadata: chartMeta,
		metaChartDigest:   chartDigest,
	}
	if objectMetadataSize(meta) > ossMetadataSoftLimitBytes {
		return nil
	}

	return meta
}

// objectMetadataSize calculates object metadata size.
func objectMetadataSize(m map[string]string) int {
	var sum int
	for k, v := range m {
		sum += len([]byte(k))
		sum += len([]byte(v))
	}
	return sum
}

// getMetadataValue retrieves a value from metadata map case-insensitively.
func getMetadataValue(meta map[string]string, key string) string {
	if v, ok := meta[key]; ok {
		return v
	}
	// Try Title case (e.g. chart-metadata -> Chart-Metadata)
	// Simple assumption: only first letter capitalized or dashed word capitalized.
	// We'll just loop to find unique match if map is small.
	keyLower := strings.ToLower(key)
	for k, v := range meta {
		if strings.ToLower(k) == keyLower {
			return v
		}
	}
	return ""
}
