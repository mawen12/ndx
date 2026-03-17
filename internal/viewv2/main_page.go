package viewv2

import (
	"context"

	"github.com/mawen12/ndx/internal"
	"github.com/rivo/tview"
)

type MainPage struct {
	*tview.Flex
	topFlex *tview.Flex
	app     *App
}

func NewMainPage() *MainPage {
	return &MainPage{
		Flex:    tview.NewFlex().SetDirection(tview.FlexRow),
		topFlex: tview.NewFlex().SetDirection(tview.FlexColumn),
	}
}

func (m *MainPage) Name() internal.PageKey {
	return internal.KeyMainPage
}

func (m *MainPage) Init(ctx context.Context) {
	m.app = extractApp(ctx)

	m.topFlex.
		AddItem(m.app.components.MustGet(internal.QueryLabelComponent), 13, 0, false).
		AddItem(nil, 1, 0, false).
		AddItem(m.app.components.MustGet(internal.QueryComponent), 0, 1, true).
		AddItem(nil, 1, 0, false).
		AddItem(m.app.components.MustGet(internal.TimeLabelComponent), tview.TaggedStringWidth(m.app.Config.TimeRange.String())+2, 0, false).
		AddItem(nil, 1, 0, false).
		AddItem(m.app.components.MustGet(internal.EditBtnComponent), 6, 0, false)

	m.AddItem(m.topFlex, 1, 0, true).
		AddItem(m.app.components.MustGet(internal.HistogramComponent), 6, 0, false).
		AddItem(m.app.components.MustGet(internal.TableComponent), 0, 1, false).
		AddItem(m.app.components.MustGet(internal.StatusLineComponent), 1, 0, false).
		AddItem(m.app.components.MustGet(internal.CmdComponent), 1, 0, false)
}

func (m *MainPage) Start() {
	if m.app.Render {
		panic("main page should render app result")
	}
}

func (m *MainPage) Stop() {

}

func (m *MainPage) IsModal() bool {
	return false
}
