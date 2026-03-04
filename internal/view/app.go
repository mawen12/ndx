package view

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/gdamore/tcell/v2"
	"github.com/mawen12/ndx/internal"
	"github.com/mawen12/ndx/internal/ui"
	"github.com/rivo/tview"
)

type App struct {
	*ui.App
	Content *PageStack
}

func NewApp() *App {
	a := App{
		App:     ui.NewApp(),
		Content: NewPageStack(),
	}

	return &a
}

func (a *App) Init() error {
	ctx := context.WithValue(context.Background(), internal.KeyApp, a)

	if err := a.Content.Init(ctx); err != nil {
		return err
	}

	a.App.Init()

	a.SetInputCapture(a.keyboard)

	//a.bindKeys()

	a.layout(ctx)

	a.initSignals()

	return nil
}

func (a *App) Run() error {
	//go func() {
	//	a.QueueUpdateDraw(func() {
	a.Main.SwitchToPage("main")
	//})
	//}()

	a.SetRunning(true)

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
	top := tview.NewFlex().SetDirection(tview.FlexColumn)
	top.AddItem(a.QueryLabel(), 12, 0, false)
	top.AddItem(nil, 1, 0, false)
	top.AddItem(a.Query(), 0, 1, true)
	top.AddItem(nil, 1, 0, false)
	top.AddItem(a.Time(), 6, 0, false)
	top.AddItem(nil, 1, 0, false)
	top.AddItem(a.Edit(), 6, 0, false)

	main := tview.NewFlex().SetDirection(tview.FlexRow)
	main.AddItem(top, 1, 0, true)
	main.AddItem(a.Content, 0, 10, true)

	a.Main.AddPage("main", main, true, false)
}

func (a *App) initSignals() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP)

	go func(sig chan os.Signal) {
		<-sig
		os.Exit(0)
	}(sig)
}
