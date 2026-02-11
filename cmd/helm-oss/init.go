package main

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"helm-oss/internal/helmutil"
	"helm-oss/internal/oss"
)

const initDesc = `This command initializes an empty repository on Alibaba Cloud OSS.

'helm oss init' takes one argument:
- URI - URI of the repository.
`

const initExample = `  helm oss init oss://bucket/charts - inits chart repository in 'bucket' bucket under 'charts' path.`

func newInitCommand() *cobra.Command {
	act := &initAction{
		printer: nil,
		uri:     "",
	}

	cmd := &cobra.Command{
		Use:     "init URI",
		Short:   "Initialize empty repository on Alibaba Cloud OSS.",
		Long:    initDesc,
		Example: initExample,
		Args:    wrapPositionalArgsBadUsage(cobra.ExactArgs(1)),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			// No completions for the URI argument.
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			act.printer = cmd
			act.uri = args[0]
			return act.run(cmd.Context())
		},
	}

	return cmd
}

type initAction struct {
	printer printer

	uri string
}

func (act *initAction) run(ctx context.Context) error {
	storage := oss.New()

	exists, err := storage.IndexExists(ctx, act.uri)
	if err != nil {
		return fmt.Errorf("check if index exists in the storage: %v", err)
	}
	if exists {
		return act.alreadyExistsInStorageError()
	}

	r, err := helmutil.NewIndex().Reader()
	if err != nil {
		return errors.WithMessage(err, "get index reader")
	}

	if err := storage.PutIndex(ctx, act.uri, r); err != nil {
		return errors.WithMessage(err, "upload index to oss")
	}

	act.printer.Printf("Initialized empty repository at %s\n\n", act.uri)
	act.printer.Printf("To add this repository to your local Helm configuration, run:\n\n")
	act.printer.Printf("  helm repo add <name> %s\n\n", act.uri)
	act.printer.Printf("Replace <name> with your preferred repository name.\n")
	return nil
}

func (act *initAction) alreadyExistsInStorageError() error {
	act.printer.PrintErrf(
		"The index file already exists in the remote storage at the provided URI.\n",
	)
	return newSilentError()
}
