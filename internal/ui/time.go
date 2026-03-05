package ui

import "github.com/rivo/tview"

type Time struct {
	*tview.TextView

	app *App
}

func NewTime(app *App) *Time {
	t := Time{
		TextView: tview.NewTextView().
			SetText("-1h").
			SetScrollable(false).
			SetTextAlign(tview.AlignCenter),
		app: app,
	}

	return &t
}

func (t *Time) Name() string {
	return "time"
}
