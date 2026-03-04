package ui

import "github.com/rivo/tview"

type Time struct {
	*tview.TextView
}

func NewTime() *Time {
	t := Time{
		TextView: tview.NewTextView(),
	}

	t.SetText("-1h")
	t.SetScrollable(false)

	return &t
}
