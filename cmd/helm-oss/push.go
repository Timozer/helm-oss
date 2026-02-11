package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"helm-oss/internal/helmutil"
	"helm-oss/internal/oss"
)

const pushDesc = `This command uploads a chart to the repository.

'helm oss push' takes two arguments:
- PATH - path to the chart file,
- REPO_OR_URI - target repository name or OSS URI.

[Provenance]

If the chart is signed, the provenance file is uploaded to the repository as well.
`

const pushExample = `  helm oss push ./epicservice-0.5.1.tgz my-repo              - uploads to repository 'my-repo' (configured via helm repo add)
  helm oss push ./epicservice-0.5.1.tgz oss://bucket/charts - uploads directly to OSS URI`

func newPushCommand() *cobra.Command {
	act := &pushAction{
		printer:   nil,
		chartPath: "",
		repoOrURI: "",
		dryRun:    false,
		force:     false,
	}

	cmd := &cobra.Command{
		Use:     "push PATH REPO_OR_URI",
		Short:   "Push chart to the repository.",
		Long:    pushDesc,
		Example: pushExample,
		Args:    wrapPositionalArgsBadUsage(cobra.ExactArgs(2)),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				// Allow file completion for the PATH argument.
				return nil, cobra.ShellCompDirectiveDefault
			}
			// No completions for the REPO_OR_URI argument.
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			act.printer = cmd
			act.chartPath = args[0]
			act.repoOrURI = args[1]
			return act.run(cmd.Context())
		},
	}

	flags := cmd.Flags()
	flags.BoolVar(&act.dryRun, "dry-run", act.dryRun, "Simulate push operation, but don't actually touch anything.")
	flags.BoolVar(&act.force, "force", act.force, "Replace the chart if it already exists. This can cause the repository to lose existing chart; use it with care.")

	return cmd
}

type pushAction struct {
	printer printer

	// args

	chartPath string
	repoOrURI string

	// flags

	dryRun bool
	force  bool
}

func (act *pushAction) run(ctx context.Context) error {
	chart, err := helmutil.LoadChart(act.chartPath)
	if err != nil {
		return err
	}

	repo, err := helmutil.NewRepository(act.repoOrURI)
	if err != nil {
		return err
	}

	if repo.ShouldUpdateCache() {
		if cachedIndex, err := helmutil.LoadIndex(repo.CacheFile()); err == nil {
			// if cached index exists, check if the same chart version exists in it.
			if cachedIndex.Has(chart.Name(), chart.Version()) {
				if !act.force {
					return act.chartExistsError()
				}
				// fallthrough on --force
			}
		}
	}

	hasProv := false
	provFile, err := os.Open(act.chartPath + ".prov")
	switch {
	case err == nil:
		hasProv = true
	case errors.Is(err, os.ErrNotExist):
		// No provenance file, ignore it.
	default:
		return fmt.Errorf("open prov file: %w", err)
	}

	fname := filepath.Base(act.chartPath)

	storage := oss.New()
	exists, err := storage.Exists(ctx, repo.URL()+"/"+fname)
	if err != nil {
		return errors.WithMessage(err, "check if chart already exists in the repository")
	}

	if exists && !act.force {
		return act.chartExistsError()
	}

	chartFile, err := os.Open(act.chartPath)
	if err != nil {
		return errors.Wrap(err, "open chart file")
	}

	hash, err := helmutil.DigestFile(act.chartPath)
	if err != nil {
		return errors.WithMessage(err, "get chart digest")
	}

	if !act.dryRun {
		chartMetaJSON, err := chart.Metadata().MarshalJSON()
		if err != nil {
			return err
		}
		if _, err := storage.PutChart(
			ctx,
			repo.URL()+"/"+fname,
			chartFile,
			string(chartMetaJSON),
			hash,
			"application/gzip",
			hasProv,
			provFile,
		); err != nil {
			return errors.WithMessage(err, "upload chart to oss")
		}
	}

	// The gap between index fetching and uploading should be as small as
	// possible to make the best effort to avoid race conditions.
	// See https://github.com/hypnoglow/helm-s3/issues/18 for more info.

	// Fetch current index, update it and upload it back.

	b, err := storage.FetchRaw(ctx, repo.IndexURL())
	if err != nil {
		return errors.WithMessage(err, "fetch current repo index")
	}

	idx := helmutil.NewIndex()
	if err := idx.UnmarshalBinary(b); err != nil {
		return errors.WithMessage(err, "load index from downloaded file")
	}

	// Use relative URLs to support both OSS plugin and HTTP access
	baseURL := ""

	if err := idx.AddOrReplace(chart.Metadata().Value(), fname, baseURL, hash); err != nil {
		return errors.WithMessage(err, "add/replace chart in the index")
	}
	idx.SortEntries()
	idx.UpdateGeneratedTime()

	idxReader, err := idx.Reader()
	if err != nil {
		return errors.WithMessage(err, "get index reader")
	}

	if !act.dryRun {
		if err := storage.PutIndex(ctx, repo.URL(), idxReader); err != nil {
			return errors.WithMessage(err, "upload index to oss")
		}

		if repo.ShouldUpdateCache() {
			if err := idx.WriteFile(repo.CacheFile(), helmutil.DefaultIndexFilePerm); err != nil {
				return errors.WithMessage(err, "update local index")
			}
		}
	}

	act.printer.Printf("Successfully uploaded the chart to the repository.\n")
	return nil
}

func (act *pushAction) chartExistsError() error {
	act.printer.PrintErrf(
		"The chart already exists in the repository and cannot be overwritten without an explicit intent.\n\n"+
			"If you want to replace existing chart, use --force flag:\n\n"+
			"  helm oss push --force %[1]s %[2]s\n\n",
		act.chartPath,
		act.repoOrURI,
	)
	return newSilentError()
}
