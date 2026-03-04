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
		"queryLabel": NewQueryLabel(),
		"query":      NewQuery(),
		"time":       NewTime(),
		"edit":       NewEdit(),
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
