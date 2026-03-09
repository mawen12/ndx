package tviews

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var ActiveStyle = tcell.Style{}.Background(tcell.ColorWhite).Foreground(tcell.ColorBlue).Bold(true)

var InactiveStyle = tcell.Style{}.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite).Bold(false)

func NewFlexRow() *tview.Flex {
	return tview.NewFlex().SetDirection(tview.FlexRow)
}

func NewFlexColumn() *tview.Flex {
	return tview.NewFlex().SetDirection(tview.FlexColumn)
}

func NewTextView(text string) *tview.TextView {
	return tview.NewTextView().SetText(text)
}

func NewButton(label string) *tview.Button {
	return tview.NewButton(label)
}

func NewInputField() *tview.InputField {
	input := tview.NewInputField()

	input.SetFocusFunc(func() {
		input.SetFieldStyle(ActiveStyle)
	})

	input.SetBlurFunc(func() {
		input.SetFieldStyle(InactiveStyle)
	})

	return input
}
