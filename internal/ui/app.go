package ui

import (
	"sync"

	"github.com/mawen12/ndx/internal/model"
	"github.com/rivo/tview"
)

type App struct {
	*tview.Application

	Main    *Pages
	model   *model.QueryContext
	views   map[string]tview.Primitive
	running bool
	mx      sync.RWMutex
}

func NewApp(m *model.QueryContext) *App {
	a := App{
		Application: tview.NewApplication(),
		Main:        NewPages(),
		model:       m,
	}

	a.views = map[string]tview.Primitive{
		"queryLabel": NewQueryLabel(&a),
		"query":      NewQuery(&a),
		"time":       NewTime(&a),
		"edit":       NewEdit(&a),
		"table":      NewTable(&a),
		"statusLine": NewStatusLine(&a),
		"cmd":        NewCommand(&a),
		"histogram":  NewHistogram(&a),
	}

	return &a
}

func (a *App) Init() {
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

// View Accessors
func (a *App) QueryLabel() *QueryLabel {
	return a.views["queryLabel"].(*QueryLabel)
}

func (a *App) Query() *Query {
	return a.views["query"].(*Query)
}

func (a *App) Time() *Time {
	return a.views["time"].(*Time)
}

func (a *App) Edit() *Edit {
	return a.views["edit"].(*Edit)
}

func (a *App) Histogram() *Histogram {
	return a.views["histogram"].(*Histogram)
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

func (a *App) activateQuery() {
	a.SetFocus(a.Query())
}

func (a *App) activateEdit() {
	a.SetFocus(a.Edit())
}

func (a *App) activateHistogram() {
	a.SetFocus(a.Histogram())
}

func (a *App) activateTable() {
	a.SetFocus(a.Table())
}

func (a *App) activateCmd(prev tview.Primitive) {
	a.Cmd().SetFocus(prev)
}
