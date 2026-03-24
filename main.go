package main

import (
	"flag"
	"log/slog"
	"os"
	"path"
	"runtime/debug"
	"time"

	"github.com/lmittmann/tint"
	"github.com/mawen12/ndx/internal/config"
	"github.com/mawen12/ndx/internal/viewv2"
	"github.com/mawen12/ndx/pkg/times"
)

func main() {
	conns := flag.String("conns", "", "comma separated list of ndx connection strings")
	logLevel := flag.String("log-level", "", "slog log level")
	flag.Parse()

	if *conns == "" {
		slog.Error("invalid `conns` string")
		os.Exit(1)
	}

	var slogLevel slog.Leveler
	if *logLevel != "" {
		switch *logLevel {
		case "ERROR", "error", "err":
			slogLevel = slog.LevelError
		case "WARN", "warn":
			slogLevel = slog.LevelWarn
		case "INFO", "info", "inf":
			slogLevel = slog.LevelInfo
		case "DEBUG", "debug":
			slogLevel = slog.LevelDebug
		default:
			slog.Error("invalid `logLevel`, fallback to Info", "logLevel", logLevel)
			slogLevel = slog.LevelInfo
		}
	}

	// init log
	log := path.Join(os.TempDir(), "ndx.log")
	_ = os.Remove(log)
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

	logger := slog.New(tint.NewHandler(logfile, &tint.Options{
		Level:      slogLevel,
		TimeFormat: time.RFC3339,
	}))
	slog.SetDefault(logger)

	var app *viewv2.App
	defer func() {
		if err := recover(); err != nil {
			if app != nil {
				app.Close()
			}

			slog.Error("Boom! ndx init failed", "err", err, slog.String("stack", string(debug.Stack())))
		}
	}()

	// build query
	q := &config.Query{
		Origin:    *conns,
		TimeRange: times.NewDefaultTimeRange(),
	}

	// new app
	app = viewv2.NewApp(q)

	app.Init()

	// start app
	if err := app.Run(); err != nil {
		slog.Error("app run failed", "error", err)
		os.Exit(1)
	}
}
