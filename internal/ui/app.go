package ui

import (
	"sync"

	"github.com/rivo/tview"
)

type App struct {
	*tview.Application

	Main    *Pages
	Model   *ViewModel
	views   map[string]tview.Primitive
	running bool
	mx      sync.RWMutex
}

func NewApp(model *ViewModel) *App {
	a := App{
		Application: tview.NewApplication(),
		Main:        NewPages(),
		Model:       model,
	}

	a.views = map[string]tview.Primitive{
		"topFlex":      tview.NewFlex(),
		"queryLabel":   NewQueryLabel(&a),
		"query":        NewQuery(&a),
		"time":         NewTime(&a),
		"edit":         NewEdit(&a),
		"table":        NewTable(&a),
		"statusLine":   NewStatusLine(&a),
		"cmd":          NewCommand(&a),
		"histogram":    NewHistogram(&a),
		"edit_view":    NewEditView(&a),
		"message_view": NewMessageView(&a),
	}

	return &a
}

func (a *App) Init() {
	a.Query().AddListener(a.QueryLabel())

	a.EditView().Init()
	a.MessageView().Init()

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

func (a *App) ShowModal(name string, p tview.Primitive) {
	modalGrid := tview.NewGrid().
		SetColumns(0, 105, 0).
		SetRows(0, 20, 0).
		AddItem(p, 1, 1, 1, 1, 0, 0, true)

	a.Main.AddPage(name, modalGrid, true, true)

	a.SetFocus(p)
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

func (a *App) EditView() *EditView {
	return a.views["edit_view"].(*EditView)
}

func (a *App) MessageView() *MessageView {
	return a.views["message_view"].(*MessageView)
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
