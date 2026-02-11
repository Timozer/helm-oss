package main

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"helm-oss/internal/helmutil"
	"helm-oss/internal/oss"
)

const deleteDesc = `This command removes a chart from the repository.

'helm oss delete' takes two arguments:
- NAME - name of the chart to delete,
- REPO_OR_URI - target repository name or OSS URI.

[Provenance]

If the chart is signed, the provenance file is removed from the repository as well.
`

const deleteExample = `  helm oss delete epicservice --version 0.5.1 my-repo              - deletes from repository 'my-repo'
  helm oss delete epicservice --version 0.5.1 oss://bucket/charts - deletes directly from OSS URI`

func newDeleteCommand() *cobra.Command {
	act := &deleteAction{
		printer:   nil,
		chartName: "",
		repoOrURI: "",
		version:   "",
	}

	cmd := &cobra.Command{
		Use:     "delete NAME REPO_OR_URI",
		Aliases: []string{"del"},
		Short:   "Delete chart from the repository.",
		Long:    deleteDesc,
		Example: deleteExample,
		Args:    wrapPositionalArgsBadUsage(cobra.ExactArgs(2)),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			// No completions for the NAME and REPO_OR_URI arguments.
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			act.printer = cmd
			act.chartName = args[0]
			act.repoOrURI = args[1]
			return act.run(cmd.Context())
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&act.version, "version", act.version, "Version of the chart to delete.")
	_ = cobra.MarkFlagRequired(flags, "version")

	return cmd
}

type deleteAction struct {
	printer printer

	chartName string
	repoOrURI string

	version string
}

func (act *deleteAction) run(ctx context.Context) error {
	repo, err := helmutil.NewRepository(act.repoOrURI)
	if err != nil {
		return err
	}

	storage := oss.New()

	b, err := storage.FetchRaw(ctx, repo.IndexURL())
	if err != nil {
		return errors.WithMessage(err, "fetch current repo index")
	}

	idx := helmutil.NewIndex()
	if err := idx.UnmarshalBinary(b); err != nil {
		return errors.WithMessage(err, "load index from downloaded file")
	}

	url, err := idx.Delete(act.chartName, act.version)
	if err != nil {
		return err
	}
	idx.UpdateGeneratedTime()

	idxReader, err := idx.Reader()
	if err != nil {
		return errors.Wrap(err, "get index reader")
	}

	if url != "" {
		if !strings.HasPrefix(url, repo.URL()) {
			url = strings.TrimSuffix(repo.URL(), "/") + "/" + url
		}

		if err := storage.DeleteChart(ctx, url); err != nil {
			return errors.WithMessage(err, "delete chart file from oss")
		}
	}

	if err := storage.PutIndex(ctx, repo.URL(), idxReader); err != nil {
		return errors.WithMessage(err, "upload new index to oss")
	}

	if repo.ShouldUpdateCache() {
		if err := idx.WriteFile(repo.CacheFile(), helmutil.DefaultIndexFilePerm); err != nil {
			return errors.WithMessage(err, "update local index")
		}
	}

	act.printer.Printf("Successfully deleted the chart from the repository.\n")
	return nil
}
