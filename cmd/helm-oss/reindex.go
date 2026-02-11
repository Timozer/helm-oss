package main

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"helm-oss/internal/helmutil"
	"helm-oss/internal/oss"
)

const reindexDesc = `This command performs a reindex of the repository.

'helm oss reindex' takes one argument:
- REPO_OR_URI - target repository name or OSS URI.
`

const reindexExample = `  helm oss reindex my-repo              - reindexes repository 'my-repo'
  helm oss reindex oss://bucket/charts - reindexes OSS URI directly`

func newReindexCommand(opts *options) *cobra.Command {
	act := &reindexAction{
		printer:   nil,
		verbose:   false,
		repoOrURI: "",
	}

	cmd := &cobra.Command{
		Use:     "reindex REPO_OR_URI",
		Short:   "Reindex the repository.",
		Long:    reindexDesc,
		Example: reindexExample,
		Args:    wrapPositionalArgsBadUsage(cobra.ExactArgs(1)),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			// No completions for the REPO_OR_URI argument.
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			act.printer = cmd
			act.verbose = opts.verbose
			act.repoOrURI = args[0]
			return act.run(cmd.Context())
		},
	}

	return cmd
}

type reindexAction struct {
	printer   printer
	verbose   bool
	repoOrURI string
}

func (act *reindexAction) run(ctx context.Context) error {
	repo, err := helmutil.NewRepository(act.repoOrURI)
	if err != nil {
		return err
	}

	storage := oss.New()

	items, errs := storage.Traverse(ctx, repo.URL())

	builtIndex := make(chan *helmutil.Index, 1)
	go func() {
		idx := helmutil.NewIndex()
		for item := range items {
			baseURL := ""

			if act.verbose {
				act.printer.Printf("[DEBUG] Adding %s to index.\n", item.Filename)
			}

			filename := item.Filename

			if err := idx.Add(item.Meta.Value(), filename, baseURL, item.Hash); err != nil {
				act.printer.PrintErrf("[ERROR] failed to add chart to the index: %s", err)
			}
		}
		idx.SortEntries()
		idx.UpdateGeneratedTime()

		builtIndex <- idx
	}()

	for err = range errs {
		return fmt.Errorf("traverse the chart repository: %v", err)
	}

	idx := <-builtIndex

	r, err := idx.Reader()
	if err != nil {
		return errors.Wrap(err, "get index reader")
	}

	if err := storage.PutIndex(ctx, repo.URL(), r); err != nil {
		return errors.Wrap(err, "upload index to the repository")
	}

	if repo.ShouldUpdateCache() {
		if err := idx.WriteFile(repo.CacheFile(), helmutil.DefaultIndexFilePerm); err != nil {
			return errors.WithMessage(err, "update local index")
		}
	}

	act.printer.Printf("Repository %s was successfully reindexed.\n", act.repoOrURI)
	return nil
}
