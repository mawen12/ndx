package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type QueryListener interface {
	Enter(old, new string)
}

type Query struct {
	*tview.InputField
	listeners []QueryListener
	last      string

	app *App
}

func NewQuery(app *App) *Query {
	q := Query{
		InputField: tview.NewInputField(),
		app:        app,
	}

	q.SetInputCapture(q.keyboard)

	q.SetFocusFunc(q.activate)

	q.SetBlurFunc(q.inactivate)

	q.SetChangedFunc(func(text string) {
		q.notifyListener(text)
	})

	return &q
}

func (q *Query) Name() string {
	return "query"
}

func (q *Query) AddListener(l QueryListener) {
	q.listeners = append(q.listeners, l)
}

func (q *Query) RemoveListener(l QueryListener) {
	for i, listener := range q.listeners {
		if listener == l {
			q.listeners = append(q.listeners[:i], q.listeners[i+1:]...)
			break
		}
	}
}

func (q *Query) notifyListener(text string) {
	for _, listener := range q.listeners {
		listener.Enter(q.last, text)
	}
}

func (q *Query) keyboard(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEsc:
		q.app.activateTable()
	case tcell.KeyTab:

		q.app.activateEdit()
	case tcell.KeyBacktab:

		q.app.activateTable()
	case tcell.KeyEnter:
		q.app.model.SetPattern(q.GetText())
		q.app.model.DoQuery()
	}

	return event
}

func (q *Query) activate() {
	q.SetFieldStyle(tcell.Style{}.Background(tcell.ColorWhite).Foreground(tcell.ColorBlue).Bold(true))
}

func (q *Query) inactivate() {
	q.SetFieldStyle(tcell.Style{}.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite).Bold(true))
}
