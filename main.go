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
	"github.com/mawen12/ndx/internal/config"
	"github.com/mawen12/ndx/internal/view"
	"github.com/mawen12/ndx/pkg/times"
)

func main() {
	conns := flag.String("conns", "", "comma separated list of ndx connection strings")
	flag.Parse()

	if *conns == "" {
		slog.Error("invalid connection string")
		os.Exit(1)
	}

	// init log
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

	slog.SetDefault(slog.New(tint.NewHandler(logfile, &tint.Options{
		Level:      slog.LevelDebug,
		TimeFormat: time.RFC3339,
	})))

	var app *view.App

	defer func() {
		if err := recover(); err != nil {
			if app != nil {
				app.Close()
			}

			slog.Error("Boom! ndx init failed", "error", err)
			slog.Error("", "stack", string(debug.Stack()))
			fmt.Printf("%s", "Boom!\n")
			fmt.Printf("%v.\n", err)
		}
	}()

	// load config
	queryConns, err := config.ParseConns(*conns)
	if err != nil {
		slog.Error("parse conns failed", "error", err)
		os.Exit(1)
	}

	// build query
	q := &config.Query{
		Conns:     queryConns,
		TimeRange: times.NewDefaultTimeRange(),
	}

	// new app
	app = view.NewApp(q)

	if err := app.Init(); err != nil {
		slog.Error("app init failed", "error", err)
		os.Exit(1)
	}

	// start app
	if err := app.Run(); err != nil {
		slog.Error("app run failed", "error", err)
		os.Exit(1)
	}
}
