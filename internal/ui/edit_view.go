package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mawen12/ndx/internal/model"
	"github.com/mawen12/ndx/pkg/tviews"
	"github.com/rivo/tview"
)

const (
	timeLabelText = `Time range in the format "[yellow]<time>[ to <time>][-]", where [yellow]<time>[-] is either absolute like "[yellow]Mar27 12:00[-]"
or relative like "[yellow]-2h30m[-]" (relative to current time). If the "to" part is omitted,
current time is used.`

	queryLabelText = `awk pattern. Examples: "[yellow]/foo bar/[-]", or "[yellow]( /foo bar/ || /other stuff/ ) && !/baz/[-]"`

	logStreamLabelText = `Logstreams. Comma-separated strings in the format "[yellow][user@]myserver.com[:port[:/path/to/logfile]][-]"
Examples: "[yellow]user@myserver.com[-]", or "[yellow]user@myserver.com:22:/var/log/syslog[-]"`

	selectQueryLabelText = `Select field expression. Example: "[yellow]time STICKY, message, lstream, level_name AS level, *[-]".`
)

type EditView struct {
	*tview.Grid
	*model.CycleList
	Model *ViewModel

	frame *tview.Frame
	flex  *tview.Flex

	timeFlex      *tview.Flex
	timeLabel     *tview.TextView
	timeInput     *tview.InputField
	timezoneLabel *tview.TextView

	logStreamLabel *tview.TextView
	logStreamInput *tview.InputField

	queryLabel *tview.TextView
	queryInput *tview.InputField

	selectQueryFlex    *tview.Flex
	selectQueryLabel   *tview.TextView
	selectQueryInput   *tview.InputField
	selectQueryEditBtn *tview.Button

	app *App
}

func NewEditView(app *App) *EditView {
	ev := EditView{
		Grid:      tview.NewGrid().SetColumns(0, 105, 0).SetRows(0, 20, 0),
		CycleList: model.NewCycleList(),
		Model:     app.Model,

		flex: tview.NewFlex().SetDirection(tview.FlexRow),

		timeFlex:      tview.NewFlex().SetDirection(tview.FlexColumn),
		timeLabel:     tview.NewTextView().SetText(timeLabelText).SetDynamicColors(true),
		timeInput:     tviews.NewInputField(),
		timezoneLabel: tview.NewTextView(),

		logStreamLabel: tview.NewTextView().SetText(logStreamLabelText).SetDynamicColors(true),
		logStreamInput: tviews.NewInputField(),

		queryLabel: tview.NewTextView().SetText(queryLabelText).SetDynamicColors(true),
		queryInput: tviews.NewInputField(),

		selectQueryFlex:    tview.NewFlex().SetDirection(tview.FlexColumn),
		selectQueryLabel:   tview.NewTextView().SetText(selectQueryLabelText).SetDynamicColors(true),
		selectQueryInput:   tviews.NewInputField(),
		selectQueryEditBtn: tview.NewButton("Edit"),

		app: app,
	}

	return &ev
}

func (ev *EditView) Init() {
	ev.handle()

	ev.layout()

}

func (ev *EditView) handle() {
	ev.PushBack(ev.timeInput)
	ev.PushBack(ev.logStreamInput)
	ev.PushBack(ev.queryInput)
	ev.PushBack(ev.selectQueryInput)

	ev.timeInput.SetInputCapture(ev.keyboard)
	ev.logStreamInput.SetInputCapture(ev.keyboard)
	ev.queryInput.SetInputCapture(ev.keyboard)
	ev.selectQueryInput.SetInputCapture(ev.keyboard)
}

func (ev *EditView) layout() {
	ev.timeFlex.
		AddItem(ev.timeInput, 0, 1, true).
		AddItem(nil, 1, 0, false).
		AddItem(ev.timezoneLabel, 0, 0, false)

	ev.selectQueryFlex.
		AddItem(ev.selectQueryInput, 0, 1, false).
		AddItem(nil, 1, 0, false).
		AddItem(ev.selectQueryEditBtn, 6, 0, false)

	ev.flex.
		AddItem(ev.timeLabel, 3, 0, false).
		AddItem(ev.timeFlex, 1, 0, true).
		AddItem(nil, 1, 0, false).
		AddItem(ev.logStreamLabel, 2, 0, false).
		AddItem(ev.logStreamInput, 1, 1, false).
		AddItem(nil, 1, 0, false).
		AddItem(ev.queryLabel, 1, 0, false).
		AddItem(ev.queryInput, 1, 0, false).
		AddItem(nil, 1, 0, false).
		AddItem(ev.selectQueryLabel, 1, 0, false).
		AddItem(ev.selectQueryFlex, 1, 0, false)

	ev.frame = tview.NewFrame(ev.flex).SetBorders(0, 0, 0, 0, 0, 0)
	ev.frame.SetBorder(true).SetBorderPadding(1, 1, 1, 1)
	ev.frame.SetTitle("Edit query params")

	ev.Grid.AddItem(ev.frame, 1, 1, 1, 1, 0, 0, true)
}

func (ev *EditView) keyboard(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEsc:
		ev.Hide()
		ev.app.activateEdit()
	case tcell.KeyTab:
		p := ev.Next().(tview.Primitive)
		ev.app.SetFocus(p)
	case tcell.KeyBacktab:
		p := ev.Prev().(tview.Primitive)
		ev.app.SetFocus(p)
	case tcell.KeyEnter:
		qv := model.QueryView{
			Conns:       ev.logStreamInput.GetText(),
			Pattern:     ev.queryInput.GetText(),
			TimeRange:   ev.timeInput.GetText(),
			SelectQuery: ev.selectQueryInput.GetText(),
		}
		if err := ev.app.Model.Save(qv); err != nil {

			modal := tview.NewModal().
				SetText(err.Error()).
				AddButtons([]string{"OK", "Cancel"}).
				SetDoneFunc(func(buttonIndex int, buttonLabel string) {

				})

			modal.SetTitle("Log query error")
			modal.SetBackgroundColor(tcell.ColorDarkRed)
			modal.SetBorder(true)
			modal.SetBorderColor(tcell.ColorDarkRed)
			modal.Box.SetBackgroundColor(tcell.ColorDarkRed)
			modal.Box.SetBorderColor(tcell.ColorWhite)
			modal.SetButtonStyle(tcell.Style{}.Foreground(tcell.ColorWhite).Background(tcell.ColorBlue))
			modal.SetButtonActivatedStyle(tcell.Style{}.Foreground(tcell.ColorBlue).Background(tcell.ColorWhite))
			ev.app.Main.AddPage("modal", modal, true, true)
			ev.app.SetFocus(modal)

		}
	}

	return event
}

func (ev *EditView) Show() {
	ev.render()

	ev.Reset()

	ev.app.Main.AddPage("edit_view", ev, true, true)

	ev.app.SetFocus(ev.frame)
}

func (ev *EditView) Hide() {
	ev.app.Main.RemovePage("edit_view")
}

func (ev *EditView) render() {
	ev.timeInput.SetText(ev.Model.TimeRange.Spec())
	ev.logStreamInput.SetText(ev.Model.Conns.String())
	ev.queryInput.SetText(ev.Model.Pattern)
	ev.selectQueryInput.SetText(ev.Model.SelectQuery)
}
