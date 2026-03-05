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
	"github.com/mawen12/ndx/internal/pool"
	"github.com/mawen12/ndx/internal/ui"
)

type App struct {
	*ui.App
	Content *PageStack

	pool *pool.Pool
}

func NewApp() *App {
	a := App{
		App:     ui.NewApp(),
		Content: NewPageStack(),
	}

	return &a
}

func (a *App) Init(conns string) error {
	ctx := context.WithValue(context.Background(), internal.KeyApp, a)

	if err := a.Content.Init(ctx); err != nil {
		return err
	}

	a.App.Init()

	a.SetInputCapture(a.keyboard)

	a.layout(ctx)

	a.initSignals()

	p, err := pool.Connect(conns)
	if err != nil {
		return err
	}

	p.AddListener(a)

	slog.Info("Connection establish success", "conn", "cmd://mawen@localhost/home/mawen/logs/app.log")

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
