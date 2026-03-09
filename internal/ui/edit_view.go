package ui

import (
	"container/list"

	"github.com/gdamore/tcell/v2"
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

	frame *tview.Frame
	flex  *tview.Flex

	topFlex      *tview.Flex
	historyLabel *tview.TextView
	backBtn      *tview.Button
	andLabel     *tview.TextView
	fwdBtn       *tview.Button

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

	linkedList *list.List
	navigator  *list.Element

	app *App
}

func NewEditView(app *App) *EditView {
	ev := EditView{
		flex: tview.NewFlex().SetDirection(tview.FlexRow),

		topFlex:      tview.NewFlex().SetDirection(tview.FlexColumn),
		historyLabel: tview.NewTextView().SetText("Query History:"),
		backBtn:      tview.NewButton("Back <Ctrl+K>"),
		andLabel:     tview.NewTextView().SetText("and"),
		fwdBtn:       tview.NewButton("Forth <Ctrl+J>"),

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

		linkedList: list.New(),
		app:        app,
	}

	return &ev
}

func (ev *EditView) Init() {
	ev.navigate()

	ev.layout()
}

type EditViewElement struct {
	tview.Primitive
	Name string
}

func NewEditViewElement(p tview.Primitive, Name string) EditViewElement {
	return EditViewElement{
		Primitive: p,
		Name:      Name,
	}
}

func (ev *EditView) navigate() {
	ev.linkedList.PushBack(NewEditViewElement(ev.timeInput, "time_input"))
	ev.linkedList.PushBack(NewEditViewElement(ev.logStreamInput, "logStream_input"))
	ev.linkedList.PushBack(NewEditViewElement(ev.queryInput, "query_input"))
	ev.linkedList.PushBack(NewEditViewElement(ev.selectQueryInput, "selectQuery_input"))
	ev.linkedList.PushBack(NewEditViewElement(ev.backBtn, "back_btn"))
	ev.linkedList.PushBack(NewEditViewElement(ev.fwdBtn, "fwd_btn"))
	ev.navigator = ev.linkedList.Front()

	ev.timeInput.SetInputCapture(ev.keyboard)
	ev.logStreamInput.SetInputCapture(ev.keyboard)
	ev.queryInput.SetInputCapture(ev.keyboard)
	ev.selectQueryInput.SetInputCapture(ev.keyboard)
	ev.backBtn.SetInputCapture(ev.keyboard)
	ev.fwdBtn.SetInputCapture(ev.keyboard)
}

func (ev *EditView) layout() {
	ev.topFlex.
		AddItem(ev.historyLabel, 15, 0, false).
		AddItem(ev.backBtn, 15, 0, false).
		AddItem(nil, 1, 0, false).
		AddItem(ev.andLabel, 3, 0, false).
		AddItem(nil, 1, 0, false).
		AddItem(ev.fwdBtn, 16, 0, false).
		AddItem(nil, 0, 1, false)

	ev.timeFlex.
		AddItem(ev.timeInput, 0, 1, true).
		AddItem(nil, 1, 0, false).
		AddItem(ev.timezoneLabel, 0, 0, false)

	ev.selectQueryFlex.
		AddItem(ev.selectQueryInput, 0, 1, false).
		AddItem(nil, 1, 0, false).
		AddItem(ev.selectQueryEditBtn, 6, 0, false)

	ev.flex.
		AddItem(ev.topFlex, 1, 0, false).
		AddItem(nil, 1, 0, false).
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

	ev.Grid = tview.NewGrid().
		SetColumns(0, 105, 0).
		SetRows(0, 20, 0).
		AddItem(ev.frame, 1, 1, 1, 1, 0, 0, true)
}

func (ev *EditView) keyboard(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEsc:
		ev.Hide()
		ev.app.activateEdit()
	case tcell.KeyTab:
		p := ev.next()
		ev.app.SetFocus(p)
	case tcell.KeyBacktab:
		p := ev.prev()
		ev.app.SetFocus(p)
	case tcell.KeyEnter:
		switch ev.current().Name {
		case "time_input", "logStream_input", "query_input", "selectQuery_input":

		case "back_btn":

		case "fwd_btn":

		}
	}

	return event
}

func (ev *EditView) current() EditViewElement {
	return ev.navigator.Value.(EditViewElement)
}

func (ev *EditView) next() tview.Primitive {
	if ev.navigator == ev.linkedList.Back() {
		ev.navigator = ev.linkedList.Front()
	} else {
		ev.navigator = ev.navigator.Next()
	}

	return ev.navigator.Value.(EditViewElement)
}

func (ev *EditView) prev() tview.Primitive {
	if ev.navigator == ev.linkedList.Front() {
		ev.navigator = ev.linkedList.Back()
	} else {
		ev.navigator = ev.navigator.Prev()
	}

	return ev.navigator.Value.(EditViewElement)
}

func (ev *EditView) Show() {
	ev.queryInput.SetText(ev.app.model.Pattern)

	ev.navigator = ev.linkedList.Front()

	ev.app.Main.AddPage("edit_view", ev, true, true)

	ev.app.SetFocus(ev.frame)
}

func (ev *EditView) Hide() {
	ev.app.Main.RemovePage("edit_view")
}
