package ui

import "github.com/rivo/tview"

type Edit struct {
	*tview.Button
}

func NewEdit() *Edit {
	e := Edit{
		Button: tview.NewButton("Edit"),
	}

	return &e
}
