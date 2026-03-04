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
}

func NewQuery() *Query {
	li := Query{
		InputField: tview.NewInputField(),
	}

	li.SetFieldStyle(tcell.Style{}.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite).Bold(true))

	li.SetChangedFunc(func(text string) {

	})

	return &li
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
		listener.Enter(q.GetText(), text)
	}
}
