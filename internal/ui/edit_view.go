package ui

import (
	"context"
	"log/slog"

	"github.com/gdamore/tcell/v2"
	"github.com/mawen12/ndx/internal/model"
	"github.com/mawen12/ndx/pkg/tviews"
	"github.com/rivo/tview"
)

const (
	timeLabelText = `Time range in the format "[yellow]<time>[ to <time>][-]", where [yellow]<time>[-] is either absolute like "[yellow]Mar27 12:00[-]" or relative like "[yellow]-2h30m[-]" (relative to current time). If the "to" part is omitted, current time is used.`

	queryLabelText = `awk pattern. Examples: "[yellow]/foo bar/[-]", or "[yellow]( /foo bar/ || /other stuff/ ) && !/baz/[-]"`

	logStreamLabelText = `Logstreams. Comma-separated strings in the format "[yellow][user@]myserver.com[:port[:/path/to/logfile]][-]" Examples: "[yellow]user@myserver.com[-]", or "[yellow]user@myserver.com:22:/var/log/syslog[-]"`

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
	logStreamInput *tview.TextArea

	queryLabel *tview.TextView
	queryInput *tview.InputField

	selectQueryFlex    *tview.Flex
	selectQueryLabel   *tview.TextView
	selectQueryInput   *tview.InputField
	selectQueryEditBtn *tview.Button

	btnFlex          *tview.Flex
	okBtn, cancelBtn *tview.Button

	app *App
}

func NewEditView(app *App) *EditView {
	ev := EditView{
		Grid:      tview.NewGrid().SetColumns(0, 105, 0).SetRows(0, 23, 0),
		CycleList: model.NewCycleList(),
		Model:     app.Model,

		flex: tview.NewFlex().SetDirection(tview.FlexRow),

		timeFlex:      tview.NewFlex().SetDirection(tview.FlexColumn),
		timeLabel:     tview.NewTextView().SetText(timeLabelText).SetDynamicColors(true),
		timeInput:     tviews.NewInputField(),
		timezoneLabel: tview.NewTextView(),

		logStreamLabel: tview.NewTextView().SetText(logStreamLabelText).SetDynamicColors(true),
		logStreamInput: tviews.NewTextArea().SetSize(0, 0),

		queryLabel: tview.NewTextView().SetText(queryLabelText).SetDynamicColors(true),
		queryInput: tviews.NewInputField(),

		selectQueryFlex:    tview.NewFlex().SetDirection(tview.FlexColumn),
		selectQueryLabel:   tview.NewTextView().SetText(selectQueryLabelText).SetDynamicColors(true),
		selectQueryInput:   tviews.NewInputField(),
		selectQueryEditBtn: tview.NewButton("Edit"),

		btnFlex:   tview.NewFlex().SetDirection(tview.FlexColumn),
		okBtn:     tview.NewButton("OK"),
		cancelBtn: tview.NewButton("Cancel"),

		app: app,
	}

	return &ev
}

func (ev *EditView) Name() string {
	return "edit_view"
}

func (ev *EditView) Init(ctx context.Context) error {
	ev.handle()

	ev.layout()

	return nil
}

func (ev *EditView) Start() {
	ev.Reset()
}

func (ev *EditView) Stop() {
	slog.Info("stop edit view")
}

func (ev *EditView) handle() {
	ev.PushBack(model.NewCyclePrimitive(ev.timeInput, "timeInput"))
	ev.PushBack(model.NewCyclePrimitive(ev.logStreamInput, "logsreamInput"))
	ev.PushBack(model.NewCyclePrimitive(ev.queryInput, "queryInput"))
	ev.PushBack(model.NewCyclePrimitive(ev.selectQueryInput, "selectQueryInput"))
	ev.PushBack(model.NewCyclePrimitive(ev.okBtn, "ok"))
	ev.PushBack(model.NewCyclePrimitive(ev.cancelBtn, "cancel"))

	ev.timeInput.SetInputCapture(ev.keyboard)
	ev.logStreamInput.SetInputCapture(ev.keyboard)
	ev.queryInput.SetInputCapture(ev.keyboard)
	ev.selectQueryInput.SetInputCapture(ev.keyboard)
	ev.okBtn.SetInputCapture(ev.keyboard)
	ev.cancelBtn.SetInputCapture(ev.keyboard)
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

	ev.btnFlex.
		AddItem(nil, 0, 1, false).
		AddItem(ev.okBtn, 10, 0, false).
		AddItem(nil, 0, 1, false).
		AddItem(ev.cancelBtn, 10, 0, false).
		AddItem(nil, 0, 1, false)

	ev.flex.
		AddItem(ev.timeLabel, 3, 0, false).
		AddItem(ev.timeFlex, 1, 0, true).
		AddItem(nil, 1, 0, false).
		AddItem(ev.logStreamLabel, 2, 0, false).
		AddItem(ev.logStreamInput, 0, 1, false).
		AddItem(nil, 1, 0, false).
		AddItem(ev.queryLabel, 1, 0, false).
		AddItem(ev.queryInput, 1, 0, false).
		AddItem(nil, 1, 0, false).
		AddItem(ev.selectQueryLabel, 1, 0, false).
		AddItem(ev.selectQueryFlex, 1, 0, false).
		AddItem(nil, 1, 0, false).
		AddItem(ev.btnFlex, 1, 0, false)

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
		return nil
	case tcell.KeyTab:
		p := ev.Next().(tview.Primitive)
		ev.app.SetFocus(p)
		return nil
	case tcell.KeyBacktab:
		p := ev.Prev().(tview.Primitive)
		ev.app.SetFocus(p)
		return nil
	case tcell.KeyEnter:
		switch ev.Current().(model.CyclePrimitive).Name {
		case "ok":
			qv := model.QueryView{
				Conns:       ev.logStreamInput.GetText(),
				Pattern:     ev.queryInput.GetText(),
				TimeRange:   ev.timeInput.GetText(),
				SelectQuery: ev.selectQueryInput.GetText(),
			}

			// save
			reconnect, err := ev.app.Model.Save(qv)
			if err != nil {
				ev.app.MessageDialog().ShowError(Message{
					Title:   "Config save failed",
					Message: err.Error(),
					IsErr:   true,
				})
				return nil
			}

			// connect
			if reconnect {
				//retChan := make(chan struct{}, 1)
				//callback := func() {
				//	retChan <- struct{}{}
				//}
				err := ev.app.Model.Refresh()
				if err != nil {
					ev.app.MessageDialog().ShowError(Message{
						Title:   "Connect failed",
						Message: err.Error(),
						IsErr:   true,
					})
					//<-retChan
					return nil
				}
			}
			// query
			if err := ev.app.Model.DoQuery(); err != nil {
				ev.app.MessageDialog().ShowError(Message{
					Title:   "Query failed",
					Message: err.Error(),
					IsErr:   true,
				})
				return nil
			}

			ev.app.Main.Pop()
			return nil
		case "cancel":
			ev.Hide()
			ev.app.activateEdit()
			return nil
		}
	}
	return event
}

func (ev *EditView) Show() {
	ev.render()

	ev.app.Main.Push(ev)
}

func (ev *EditView) Hide() {
	ev.app.Main.Pop()
}

func (ev *EditView) render() {
	ev.timeInput.SetText(ev.Model.TimeRange.Spec())
	ev.logStreamInput.SetText(ev.Model.Origin, false)
	ev.queryInput.SetText(ev.Model.Pattern)
	ev.selectQueryInput.SetText(ev.Model.SelectQuery)
}
