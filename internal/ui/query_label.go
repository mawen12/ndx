package ui

import (
	"github.com/rivo/tview"
)

const (
	queryLabelMatch    = "awk pattern: "
	queryLabelMismatch = "awk pattern[yellow::b]*[-::-]:"
)

type QueryLabel struct {
	*tview.TextView

	app *App
}

func NewQueryLabel(app *App) *QueryLabel {
	q := &QueryLabel{
		TextView: tview.NewTextView().
			SetText(queryLabelMatch).
			SetDynamicColors(true).
			SetScrollable(false).
			SetTextAlign(tview.AlignLeft),
		app: app,
	}

	return q
}

func (q *QueryLabel) Name() string {
	return "queryLabel"
}

func (q *QueryLabel) Enter(old, new string) {
	if old != new {
		q.SetText(queryLabelMismatch)
	} else {
		q.SetText(queryLabelMatch)
	}
}

func (q *QueryLabel) Done(text string) {

}
