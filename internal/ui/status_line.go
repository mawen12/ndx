package ui

import (
	"github.com/rivo/tview"
)

type StatusLine struct {
	*tview.Flex
	left, right *tview.TextView

	app *App
}

func NewStatusLine(app *App) *StatusLine {
	sl := StatusLine{
		Flex:  tview.NewFlex(),
		left:  tview.NewTextView(),
		right: tview.NewTextView(),
		app:   app,
	}

	sl.
		AddItem(sl.left, 0, 1, false).
		AddItem(nil, 1, 0, false).
		AddItem(sl.right, 30, 0, false)

	//sl.SetBackgroundColor(tcell.ColorGreen)

	sl.left.
		SetScrollable(false).
		SetDynamicColors(true)
	//SetBackgroundColor(tcell.ColorGreen)

	sl.right.
		SetScrollable(false).
		SetDynamicColors(true).
		SetTextAlign(tview.AlignRight)
	//SetBackgroundColor(tcell.ColorGreen)

	return &sl
}

func (sl *StatusLine) Name() string {
	return "statusLine"
}

func (sl *StatusLine) ShowLeft(text string) {
	sl.left.SetText(text)
}

func (sl *StatusLine) ShowRight(text string) {
	sl.right.SetText(text)
}
