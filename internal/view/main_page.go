package view

import (
	"context"

	"github.com/rivo/tview"
)

type MainPage struct {
	*tview.Flex
	Top *tview.Flex

	app *App
}

func NewMainPage(app *App) *MainPage {
	mp := MainPage{
		Flex: tview.NewFlex().SetDirection(tview.FlexRow),
		Top:  tview.NewFlex().SetDirection(tview.FlexColumn),
		app:  app,
	}

	mp.Top.AddItem(app.QueryLabel(), 13, 0, false)
	mp.Top.AddItem(nil, 1, 0, false)
	mp.Top.AddItem(app.Query(), 0, 1, true)
	mp.Top.AddItem(nil, 1, 0, false)
	mp.Top.AddItem(app.Time(), 6, 0, false)
	mp.Top.AddItem(nil, 1, 0, false)
	mp.Top.AddItem(app.Edit(), 6, 0, false)

	mp.AddItem(mp.Top, 1, 0, true)
	mp.AddItem(app.Histogram(), 6, 0, false)
	mp.AddItem(app.Table(), 0, 1, false)
	mp.AddItem(app.StatusLine(), 1, 0, false)
	mp.AddItem(app.Cmd(), 1, 0, false)

	app.Time().AddListener(&mp)

	return &mp
}

func (mp *MainPage) SetTimeText(timeStr string) {
	mp.Top.ResizeItem(mp.app.Time(), len(timeStr), 0)
}

func (mp *MainPage) Name() string {
	return "main"
}

func (mp *MainPage) Init(ctx context.Context) error {

	return nil
}

func (mp *MainPage) Start() {
	mp.app.Query().SetText(mp.app.Model.Pattern)
	mp.app.Time().SetText(mp.app.Model.TimeRange.String())
}

func (mp *MainPage) Stop() {

}
