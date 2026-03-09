package ui

import (
	"github.com/gdamore/tcell/v2"
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

	e.SetInputCapture(e.keyboard)

	e.SetSelectedFunc(e.selected)

	return &e
}

func (e *Edit) Name() string {
	return "edit"
}

func (e *Edit) keyboard(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyTab:
		e.app.activateHistogram()
	case tcell.KeyBacktab:
		e.app.activateQuery()
	default:
		switch event.Rune() {
		case ':':
			e.app.activateCmd(e)
		}
	}

	return event
}

func (e *Edit) selected() {
	//e.app.ShowModal("edit_view", e.app.EditView())
	e.app.EditView().Show()
}
