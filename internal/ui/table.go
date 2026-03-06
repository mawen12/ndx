package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Table struct {
	*tview.Table

	app *App
}

func NewTable(app *App) *Table {
	t := Table{
		Table: tview.NewTable(),
		app:   app,
	}

	t.SetInputCapture(t.keyboard)

	t.SetFocusFunc(t.activate)

	t.SetBlurFunc(t.inactivate)

	return &t
}

func (t *Table) Name() string {
	return "table"
}

func (t *Table) keyboard(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyTab:
		t.app.activateQuery()
	case tcell.KeyBacktab:
		t.app.activateEdit()
	default:
		switch event.Rune() {
		case 'i':
			t.app.activateQuery()
		case ':':
			t.app.activateCmd(t)
		}
	}

	return event
}

func (t *Table) activate() {
	t.SetSelectable(true, false)
}

func (t *Table) inactivate() {
	t.SetSelectable(false, false)
}
