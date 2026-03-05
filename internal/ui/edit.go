package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mawen12/ndx/internal/model"
	"github.com/rivo/tview"
)

type Edit struct {
	*tview.Button

	app *App
}

func NewEdit(app *App) *Edit {
	e := Edit{
		Button: tview.NewButton("Edit"),
		app:    app,
	}

	e.Button.SetTitleAlign(tview.AlignCenter)

	e.SetInputCapture(e.eventHandle)

	return &e
}

func (e *Edit) Name() string {
	return "edit"
}

func (e *Edit) eventHandle(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyTab:
		e.UnFocus()
		e.app.Table().SetFocus(e)
	case tcell.KeyBacktab:
		e.UnFocus()
		e.app.Query().SetFocus(e)
	default:
		switch event.Rune() {
		case ':':
			e.UnFocus()
			e.app.Cmd().SetFocus(e)
		}
	}

	return event
}

func (e *Edit) SetFocus(prev model.Focusable) {
	e.app.SetFocus(e)
}

func (e *Edit) UnFocus() {

}
