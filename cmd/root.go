package cmd

import (
	"os"

	"github.com/mawen12/ndx/internal/config"
	"github.com/mawen12/ndx/internal/view"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "dev"
	date    = "n/a"

	rootCmd = &cobra.Command{
		Use:   config.AppName,
		Short: "Ndx is a CLI to view and manage your Multi-Host Logs.",
		Long:  "Ndx is a CLI to view and manage your Multi-Host Logs.",
		RunE:  run,
	}
)

func init() {
	rootCmd.AddCommand(versionCmd())
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(*cobra.Command, []string) error {
	app := view.NewApp()

	if err := app.Init(); err != nil {
		return err
	}

	if err := app.Run(); err != nil {
		return err
	}

	return nil
}
