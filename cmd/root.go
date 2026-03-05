package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"runtime/debug"
	"time"

	"github.com/lmittmann/tint"
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
	log := path.Join(os.TempDir(), "ndx.log")
	logfile, err := os.OpenFile(log, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("log file %q init failed: %w", log, err)
	}

	defer func() {
		if logfile != nil {
			_ = logfile.Close()
		}
	}()
	defer func() {
		if err := recover(); err != nil {
			slog.Error("Boom! ndx init failed", "error", err)
			slog.Error("", "stack", string(debug.Stack()))
			fmt.Printf("%s", "Boom!\n")
			fmt.Printf("%v.\n", err)
		}
	}()

	slog.SetDefault(slog.New(tint.NewHandler(logfile, &tint.Options{
		Level:      slog.LevelDebug,
		TimeFormat: time.RFC3339,
	})))

	app := view.NewApp()

	if err := app.Init(); err != nil {
		return err
	}

	if err := app.Run(); err != nil {
		return err
	}

	return nil
}
