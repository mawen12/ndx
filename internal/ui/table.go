package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mawen12/ndx/internal/model"
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

	t.SetInputCapture(t.eventHandle)

	return &t
}

func (t *Table) Name() string {
	return "table"
}

func (t *Table) eventHandle(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyTab:
		t.UnFocus()
		t.app.Query().SetFocus(t)
	case tcell.KeyBacktab:
		t.UnFocus()
		t.app.Edit().SetFocus(t)
	default:
		switch event.Rune() {
		case 'i':
			t.UnFocus()
			t.app.Query().SetFocus(t)
		case ':':
			t.UnFocus()
			t.app.Cmd().SetFocus(t)
		}
	}

	return event
}

func (t *Table) SetFocus(prev model.Focusable) {
	t.app.SetFocus(t)
	t.SetSelectable(true, false)
}

func (t *Table) UnFocus() {
	t.SetSelectable(false, false)
}
