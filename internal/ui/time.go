package ui

import "github.com/rivo/tview"

type TimeListener interface {
	SetTimeText(text string)
}

type Time struct {
	*tview.TextView
	listeners []TimeListener

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

func (t *Time) AddListener(l TimeListener) {
	t.listeners = append(t.listeners, l)
}

func (t *Time) RemoveListener(l TimeListener) {
	for i, listener := range t.listeners {
		if listener == l {
			t.listeners = append(t.listeners[:i], t.listeners[i+1:]...)
			return
		}
	}
}

func (t *Time) SetText(timeStr string) {
	t.TextView.SetText(timeStr)

	for _, l := range t.listeners {
		l.SetTimeText(timeStr)
	}
}

func (t *Time) Name() string {
	return "time"
}
