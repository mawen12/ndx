package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path"
	"runtime/debug"
	"strings"
	"time"

	"github.com/lmittmann/tint"
	"github.com/mawen12/ndx/internal/config"
	"github.com/mawen12/ndx/internal/viewv2"
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

	var app *viewv2.App
	defer func() {
		if err := recover(); err != nil {
			if app != nil {
				app.Close()
			}

			slog.Error("Boom! ndx init failed", "error", err)
			stack := string(debug.Stack())
			lines := strings.Split(stack, "\n")
			for _, line := range lines {
				slog.Error("", "stack", line)
			}
			fmt.Printf("%s", "Boom!\n")
			fmt.Printf("%v.\n", err)
		}
	}()

	// build query
	q := &config.Query{
		Origin:    *conns,
		TimeRange: times.NewDefaultTimeRange(),
	}

	// new app
	app = viewv2.NewApp(q)

	//if err := app.Init(); err != nil {
	//	slog.Error("app init failed", "error", err)
	//	os.Exit(1)
	//}

	// start app
	if err := app.Run(); err != nil {
		slog.Error("app run failed", "error", err)
		os.Exit(1)
	}
}
