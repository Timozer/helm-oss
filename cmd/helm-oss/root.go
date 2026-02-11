package main

import (
	"context"
	"os"

	"github.com/spf13/cobra"
)

const rootDesc = `Manage chart repositories on Alibaba Cloud OSS.

This plugin provides OSS integration for Helm.

Basic usage:

  $ helm oss init oss://bucket-name/charts

  $ helm repo add mynewrepo oss://bucket-name/charts

  $ helm oss push ./epicservice-0.7.2.tgz mynewrepo

  $ helm search repo mynewrepo

  $ helm fetch mynewrepo/epicservice --version 0.7.2

  $ helm oss delete epicservice --version 0.7.2 mynewrepo

For detailed documentation, see README at https://github.com/Timozer/helm-oss

[Verbose output]

You can enable verbose output with '--verbose' flag.
`

func newRootCmd() *cobra.Command {
	ctx, cancel := context.WithCancel(context.Background())

	opts := newDefaultOptions()

	cmd := &cobra.Command{
		Use:   "oss",
		Short: "Manage chart repositories on Alibaba Cloud OSS",
		Long:  rootDesc,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			ctx, cancel = context.WithTimeout(cmd.Context(), opts.timeout)
			cmd.SetContext(ctx)
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			cancel()
		},
		// Completion is disabled for now.
		// Also, see: https://helm.sh/docs/topics/plugins/#static-auto-completion
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		// The command may produce system error, even if the usage is correct.
		SilenceUsage: true,
		// We handle errors by ourselves.
		SilenceErrors: true,
	}

	flags := cmd.PersistentFlags()
	flags.BoolVar(&opts.verbose, "verbose", opts.verbose, "Enable verbose output.")

	cmd.SetFlagErrorFunc(func(command *cobra.Command, err error) error {
		return newBadUsageError(err)
	})

	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)

	cmd.AddCommand(
		newDownloadCommand(),
		newInitCommand(),
		newPushCommand(),
		newReindexCommand(opts),
		newDeleteCommand(),
		newVersionCommand(),
	)

	return cmd
}
