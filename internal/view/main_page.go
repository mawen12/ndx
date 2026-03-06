package view

import (
	"github.com/rivo/tview"
)

type MainPage struct {
	*tview.Flex

	app *App
}

func NewMainPage(app *App) *MainPage {
	mp := MainPage{
		Flex: tview.NewFlex().SetDirection(tview.FlexRow),
		app:  app,
	}

	top := tview.NewFlex().SetDirection(tview.FlexColumn)
	top.AddItem(app.QueryLabel(), 13, 0, false)
	top.AddItem(nil, 1, 0, false)
	top.AddItem(app.Query(), 0, 1, true)
	top.AddItem(nil, 1, 0, false)
	top.AddItem(app.Time(), 6, 0, false)
	top.AddItem(nil, 1, 0, false)
	top.AddItem(app.Edit(), 6, 0, false)

	mp.AddItem(top, 1, 0, true)
	mp.AddItem(app.Histogram(), 6, 0, false)
	mp.AddItem(app.Table(), 0, 1, false)
	mp.AddItem(app.StatusLine(), 1, 0, false)
	mp.AddItem(app.Cmd(), 1, 0, false)

	return &mp
}
