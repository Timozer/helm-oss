package main

import (
	"github.com/spf13/cobra"
)

const versionDesc = `This command prints plugin version.`

func newVersionCommand() *cobra.Command {
	act := &versionAction{}

	cmd := &cobra.Command{
		Use:     "version",
		Short:   "Print plugin version.",
		Long:    versionDesc,
		Example: "",
		Args:    wrapPositionalArgsBadUsage(cobra.NoArgs),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			// No completions.
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			act.printer = cmd
			return act.run()
		},
	}

	return cmd
}

type versionAction struct {
	printer printer
}

func (act *versionAction) run() error {
	act.printer.Printf("%s\n", version)
	return nil
}
