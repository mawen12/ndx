package ui

import (
	"log/slog"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	queryLabelMatch    = "awk pattern: "
	queryLabelMismatch = "awk pattern[yellow::b]*[-::-]"
)

type QueryLabel struct {
	*tview.TextView
}

func NewQueryLabel() *QueryLabel {
	q := &QueryLabel{
		TextView: tview.NewTextView().
			SetText(queryLabelMatch).
			SetDynamicColors(true).
			SetScrollable(false),
	}

	q.SetBackgroundColor(tcell.ColorWhite)

	return q
}

func (q *QueryLabel) Enter(old, new string) {
	slog.Info("old: %s, new: %s", old, new)

	if old != new {
		q.SetText(queryLabelMismatch)
	} else {
		q.SetText(queryLabelMatch)
	}
}
