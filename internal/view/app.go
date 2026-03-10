package view

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gdamore/tcell/v2"
	"github.com/mawen12/ndx/internal"
	"github.com/mawen12/ndx/internal/config"
	"github.com/mawen12/ndx/internal/pool"
	"github.com/mawen12/ndx/internal/ui"
)

type App struct {
	*ui.App
	Content *PageStack

	Config      *config.Query
	pool        *pool.Pool
	queryCancel context.CancelFunc
}

func NewApp(config *config.Query) *App {

	a := App{
		App:     ui.NewApp(&ui.ViewModel{Query: config}),
		Content: NewPageStack(),
		Config:  config,
	}

	a.App.Model.QueryFunc = a.doQuery
	a.App.Model.RefreshFunc = a.refresh

	return &a
}

func (a *App) Init() error {
	ctx := context.WithValue(context.Background(), internal.KeyApp, a)

	if err := a.Content.Init(ctx); err != nil {
		return err
	}

	a.App.Init()

	a.SetInputCapture(a.keyboard)

	a.layout(ctx)

	a.initSignals()

	p, err := pool.Connect(a.Config.Conns)
	if err != nil {
		return err
	}

	p.AddListener(a)

	a.pool = p

	return nil
}

func (a *App) Run() error {
	go func() {
		a.QueueUpdateDraw(func() {
			a.Main.SwitchToPage("main")
		})
	}()

	a.SetRunning(true)
	slog.Info("App is running")

	if err := a.Application.Run(); err != nil {
		return err
	}

	return nil
}

func (a *App) keyboard(event *tcell.EventKey) *tcell.EventKey {
	//if k, ok := a.HasAction(ui.AsKey(event)); ok {
	//	return k.Action(event)
	//}
	//slog.Error("keyboard not implemented")
	return event
}

func (a *App) layout(ctx context.Context) {
	main := NewMainPage(a)

	a.Main.AddPage("main", main, true, false)
}

func (a *App) initSignals() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP)

	go func(sig chan os.Signal) {
		<-sig
		slog.Info("Receive SIGHUP, exiting...")
		os.Exit(0)
	}(sig)
}

func (a *App) OnStat(stat pool.Stat) {

	slog.Info("receive stat update", "stat", stat)

	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("[%s:-:%s]%s %.2d[-:-:-]", "green", "-", "🖳", stat.Idle))
	sb.WriteString(fmt.Sprintf("[%s:-:%s]%s %.2d[-:-:-]", "orange", "-", "🖳", stat.Busy))
	sb.WriteString(fmt.Sprintf("[%s:-:%s]%s %.2d[-:-:-]", "red", "-", "🖳", stat.Closed))

	a.QueueUpdateDraw(func() {
		a.StatusLine().ShowLeft(sb.String())
	})
}

func (a *App) doQuery() error {
	result, err := a.pool.Query(context.Background(), *a.Config)
	if err != nil {
		return err
	}

	a.Histogram().SetRange(int(a.Config.TimeRange.ActualFrom.Unix()), int(a.Config.TimeRange.ActualTo.Unix()))
	a.Histogram().SetData(result.Stat)

	a.Table().ShowLogs(result.Lines)

	a.Cmd().ShowQueryDuration(result.Duration)

	return nil
}

func (a *App) refresh() error {
	a.pool.Close()

	p, err := pool.Connect(a.Model.Conns)
	if err != nil {
		return err
	}

	a.pool = p

	return nil
}
