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

	t.SetInputCapture(t.keyboard)

	t.SetFocusFunc(t.activate)

	t.SetBlurFunc(t.inactivate)

	return &t
}

func (t *Table) Name() string {
	return "table"
}

func (t *Table) ShowLogs(lines []model.LogLine) {
	t.Clear()

	newTableCellFunc := func(line model.LogLine) *tview.TableCell {
		tc := tview.NewTableCell(tview.Escape(line.OriginalLine())).
			SetSelectable(true).
			SetAttributes(tcell.AttrBold).
			SetAttributes(tview.AlignLeft).
			SetSelectedStyle(tcell.Style{}.Background(tcell.ColorWhite).Foreground(tcell.ColorBlue)).
			SetStyle(tcell.Style{}.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite))

		tc.SetReference(line)

		return tc
	}

	for i, line := range lines {
		t.SetCell(i, 0, newTableCellFunc(line))
	}

	t.Select(len(lines)-1, 0)
	t.ScrollToEnd()
}

func (t *Table) keyboard(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyTab:
		t.app.activateQuery()
	case tcell.KeyBacktab:
		t.app.activateHistogram()
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
