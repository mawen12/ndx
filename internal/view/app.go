package view

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/gdamore/tcell/v2"
	"github.com/mawen12/ndx/internal"
	"github.com/mawen12/ndx/internal/config"
	"github.com/mawen12/ndx/internal/model"
	"github.com/mawen12/ndx/internal/pool"
	"github.com/mawen12/ndx/internal/ui"
)

type App struct {
	*ui.App

	Config      *config.Query
	pool        *pool.Pool
	queryCancel context.CancelFunc
}

func NewApp(config *config.Query) *App {

	a := App{
		App:    ui.NewApp(&ui.ViewModel{Query: config}),
		Config: config,
	}

	a.App.Model.RefreshFunc = a.refresh
	a.App.Model.QueryFunc = a.query

	return &a
}

func (a *App) Init() error {
	ctx := context.WithValue(context.Background(), internal.KeyApp, a)

	a.App.Init(ctx)

	a.SetInputCapture(a.keyboard)

	a.layout(ctx)

	a.initSignals()

	a.Main.AddListener(a)

	return nil
}

func (a *App) Run() error {
	a.SetRunning(true)
	slog.Info("App is running")

	go a.EditView().Show()

	if err := a.Application.Run(); err != nil {
		return err
	}

	return nil
}

func (a *App) Close() error {
	a.pool.Close()

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
	//main := NewMainPage(a)

	//a.Main.AddPage("main", main, true, true)
	a.Main.Push(NewMainPage(a))
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

func (a *App) refresh() error {
	if a.pool != nil {
		a.pool.Close()
	}

	a.pool = pool.NewPool(a.Config.Conns)

	ctx, cancel := context.WithCancel(context.Background())

	noticeCh := make(chan model.Notice, 10)

	callback := func(conn, message string, finished bool) {
		noticeCh <- model.Notice{Conn: conn, Message: message, Finished: finished}
	}

	var wg sync.WaitGroup
	wg.Add(1)

	onOk := func() {
		defer wg.Done()
	}
	onCancel := func() {
		defer wg.Done()
		cancel()
	}
	a.MessageConnect().ShowError(noticeCh, onOk, onCancel)

	err := a.pool.Connect(ctx, callback)
	close(noticeCh)
	wg.Wait()
	if err != nil {
		return err
	}

	a.pool.AddListener(a)
	return nil
}

func (a *App) query() error {
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

func (a *App) StackPushed(c model.Component) {
	c.Start()
	a.QueueUpdateDraw(func() {
		a.SetFocus(c)
	})
}

func (a *App) StackPopped(old, new model.Component) {
	old.Stop()
	a.StackTop(new)
}

func (a *App) StackTop(top model.Component) {
	if top == nil {
		return
	}
	top.Start()
	a.SetFocus(top)
}
