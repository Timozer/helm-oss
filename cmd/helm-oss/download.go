package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"helm-oss/internal/oss"
)

const downloadDesc = `This command downloads a chart from Alibaba Cloud OSS.

Note that this command basically implements downloader plugin for Helm
and not intended to be run explicitly. For more information, see:
https://helm.sh/docs/topics/plugins/#downloader-plugins

'helm oss download' takes four arguments:
- CERT - certificate file,
- KEY - key file,
- CA - certificate authority file,
- URL - full url.
`

func newDownloadCommand() *cobra.Command {
	act := &downloadAction{
		certFile: "",
		keyFile:  "",
		caFile:   "",
		url:      "",
	}

	cmd := &cobra.Command{
		Use:     "download CERT KEY CA URL",
		Short:   "Download chart from Alibaba Cloud OSS.",
		Long:    downloadDesc,
		Example: "",
		Args:    wrapPositionalArgsBadUsage(cobra.ExactArgs(4)),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			// No completions for the arguments.
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			act.printer = cmd
			act.certFile = args[0]
			act.keyFile = args[1]
			act.caFile = args[2]
			act.url = args[3]
			return act.run(cmd.Context())
		},
		Hidden: true,
	}

	return cmd
}

type downloadAction struct {
	printer printer

	// args

	certFile string
	keyFile  string
	caFile   string
	url      string
}

func (act *downloadAction) run(ctx context.Context) error {
	const indexYaml = "index.yaml"

	storage := oss.New()

	b, err := storage.FetchRaw(ctx, act.url)
	if err != nil {
		if strings.HasSuffix(act.url, indexYaml) && err == oss.ErrObjectNotFound {
			act.printer.PrintErrf(
				"The index file does not exist by the path %s. "+
					"If you haven't initialized the repository yet, try running `helm oss init %s`",
				act.url,
				strings.TrimSuffix(strings.TrimSuffix(act.url, indexYaml), "/"),
			)
			return newSilentError()
		}

		return errors.WithMessage(err, fmt.Sprintf("fetch from oss url=%s", act.url))
	}

	// Do not use printer, use os.Stdout directly, as required by Helm.
	fmt.Print(string(b))
	return nil
}
