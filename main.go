package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path"
	"runtime/debug"
	"time"

	"github.com/lmittmann/tint"
	"github.com/mawen12/ndx/internal/view"
)

func main() {
	conn := flag.String("conns", "", "comma separated list of ndx connection strings")
	flag.Parse()

	if *conn == "" {
		slog.Error("invalid connection string")
		os.Exit(1)
	}

	log := path.Join(os.TempDir(), "ndx.log")
	logfile, err := os.OpenFile(log, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		slog.Error("log file init failed", "log", log, "err", err)
		os.Exit(1)
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

	if err := app.Init(*conn); err != nil {
		slog.Error("app init failed", "error", err)
		os.Exit(1)
	}

	if err := app.Run(); err != nil {
		slog.Error("app run failed", "error", err)
		os.Exit(1)
	}
}
