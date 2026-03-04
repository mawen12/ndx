package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func versionCmd() *cobra.Command {
	command := cobra.Command{
		Use:   "version",
		Short: "Print version/build info",
		Long:  "Print version/build information",
		Run:   printVersion,
	}

	return &command
}

func printVersion(*cobra.Command, []string) {
	const fmat = "%-20s %s\n"

	_, _ = fmt.Fprintf(os.Stdout, fmat, "Version", version)
	_, _ = fmt.Fprintf(os.Stdout, fmat, "Commit", commit)
	_, _ = fmt.Fprintf(os.Stdout, fmat, "Date", date)
}
