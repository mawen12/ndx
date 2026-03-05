package ui

import (
	"log/slog"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type App struct {
	*tview.Application

	Main    *Pages
	actions *KeyActions
	views   map[string]tview.Primitive
	running bool
	mx      sync.RWMutex
}

func NewApp() *App {
	a := App{
		Application: tview.NewApplication(),
		Main:        NewPages(),
	}

	a.views = map[string]tview.Primitive{
		"queryLabel": NewQueryLabel(&a),
		"query":      NewQuery(&a),
		"time":       NewTime(&a),
		"edit":       NewEdit(&a),
		"table":      NewTable(&a),
		"statusLine": NewStatusLine(&a),
		"cmd":        NewCommand(&a),
	}

	return &a
}

func (a *App) Init() {
	a.bindKeys()

	a.Query().AddListener(a.QueryLabel())

	a.SetRoot(a.Main, true)
}

func (a *App) IsRunning() bool {
	a.mx.RLock()
	defer a.mx.RUnlock()
	return a.running
}

func (a *App) SetRunning(r bool) {
	a.mx.Lock()
	defer a.mx.Unlock()
	a.running = r
}

func (a *App) QueueUpdate(f func()) {
	if a.Application == nil {
		return
	}
	go func() {
		a.Application.QueueUpdate(f)
	}()
}

func (a *App) QueueUpdateDraw(f func()) {
	if a.Application == nil {
		return
	}
	go func() {
		a.Application.QueueUpdateDraw(f)
	}()
}

func (a *App) bindKeys() {
	slog.Error("bindKeys not implemented")
	a.actions = NewKeyActionsFromMap(KeyMap{
		//KeyColon: NewKeyAction("Cmd", , false),
	})
}

// View Accessors
func (a *App) QueryLabel() *QueryLabel {
	return a.views["queryLabel"].(*QueryLabel)
}

func (a *App) Query() *Query {
	return a.views["query"].(*Query)
}

func (a *App) Edit() *Edit {
	return a.views["edit"].(*Edit)
}

func (a *App) Time() *Time {
	return a.views["time"].(*Time)
}

func (a *App) Table() *Table {
	return a.views["table"].(*Table)
}

func (a *App) StatusLine() *StatusLine {
	return a.views["statusLine"].(*StatusLine)
}

func (a *App) Cmd() *Command {
	return a.views["cmd"].(*Command)
}

func AsKey(event *tcell.EventKey) tcell.Key {
	if event.Key() != tcell.KeyRune {
		return event.Key()
	}

	key := tcell.Key(event.Rune())
	if event.Modifiers() == tcell.ModAlt {
		key = tcell.Key(int16(event.Rune()) * int16(event.Modifiers()))
	}
	return key
}
