package main

import (
	"os"

	"helm-oss/internal/helmutil"
)

var (
	version = "dev"
)

func main() {
	helmutil.SetupHelm()

	cmd := newRootCmd()

	if err := cmd.Execute(); err != nil {
		if errorTypeSilent.Is(err) {
			os.Exit(1)
		}

		cmd.PrintErrln("Error:", err.Error())

		if errorTypeBadUsage.Is(err) {
			cmd.PrintErrf("Run '%v --help' for usage.\n", cmd.CommandPath())
		}

		os.Exit(1)
	}
}
