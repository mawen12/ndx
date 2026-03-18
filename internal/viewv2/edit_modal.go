package viewv2

import (
	"context"

	"github.com/gdamore/tcell/v2"
	"github.com/mawen12/ndx/internal"
	"github.com/rivo/tview"
)

const (
	timeLabelText = `Time range in the format "[yellow]<time>[ to <time>][-]", where [yellow]<time>[-] is either absolute like "[yellow]Mar27 12:00[-]" or relative like "[yellow]-2h30m[-]" (relative to current time). If the "to" part is omitted, current time is used.`

	queryLabelText = `awk pattern. Examples: "[yellow]/foo bar/[-]", or "[yellow]( /foo bar/ || /other stuff/ ) && !/baz/[-]"`

	logStreamLabelText = `Logstreams. Comma-separated strings in the format "[yellow][user@]myserver.com[:port[:/path/to/logfile]][-]" Examples: "[yellow]user@myserver.com[-]", or "[yellow]user@myserver.com:22:/var/log/syslog[-]"`

	selectQueryLabelText = `Select field expression. Example: "[yellow]time STICKY, message, lstream, level_name AS level, *[-]".`
)

type EditModal struct {
	*tview.Box
	frame *tview.Frame
	flex  *tview.Flex

	app *App

	timeLabel        *TextDesc
	time             *TextInput
	queryLabel       *TextDesc
	query            *TextInput
	logStreamsLabel  *TextDesc
	logStreams       *LogStreamsInput
	selectQueryLabel *TextDesc
	selectQuery      *TextInput
}

func NewEditModal() *EditModal {
	e := &EditModal{
		Box:  tview.NewBox().SetBorder(true),
		flex: tview.NewFlex().SetDirection(tview.FlexRow),
	}

	e.frame = tview.NewFrame(e.flex)

	return e
}

func (e *EditModal) Name() internal.PageKey {
	return internal.KeyEditModal
}

func (e *EditModal) Init(ctx context.Context) {
	e.app = extractApp(ctx)

	e.timeLabel = e.app.components.MustGet(internal.EditViewTimeLabelComponent).(*TextDesc)
	e.time = e.app.components.MustGet(internal.EditViewTimeComponent).(*TextInput)
	e.queryLabel = e.app.components.MustGet(internal.EditViewQueryLabelComponent).(*TextDesc)
	e.query = e.app.components.MustGet(internal.EditViewQueryComponent).(*TextInput)
	e.logStreamsLabel = e.app.components.MustGet(internal.EditViewLogStreamLabelComponent).(*TextDesc)
	e.logStreams = e.app.components.MustGet(internal.EditViewLogStreamComponent).(*LogStreamsInput)
	e.selectQueryLabel = e.app.components.MustGet(internal.EditViewLogStreamLabelComponent).(*TextDesc)
	e.selectQuery = e.app.components.MustGet(internal.EditViewSelectQueryComponent).(*TextInput)

	e.flex.
		AddItem(e.app.components.MustGet(internal.EditViewTimeLabelComponent), 2, 0, false).
		AddItem(e.app.components.MustGet(internal.EditViewTimeComponent), 1, 0, true).
		AddItem(nil, 1, 0, false).
		AddItem(e.app.components.MustGet(internal.EditViewQueryLabelComponent), 1, 0, false).
		AddItem(e.app.components.MustGet(internal.EditViewQueryComponent), 1, 0, false).
		AddItem(nil, 1, 0, false).
		AddItem(e.app.components.MustGet(internal.EditViewLogStreamLabelComponent), 2, 0, false).
		AddItem(e.app.components.MustGet(internal.EditViewLogStreamComponent), 0, 1, false).
		AddItem(nil, 1, 0, false).
		AddItem(e.app.components.MustGet(internal.EditViewSelectQueryLabelComponent), 1, 0, false).
		AddItem(e.app.components.MustGet(internal.EditViewSelectQueryComponent), 1, 0, false).
		AddItem(nil, 1, 0, false)
}

func (e *EditModal) Start() {
	e.time.SetText(e.app.ConfigView.TimeRange)
	e.query.SetText(e.app.ConfigView.Pattern)
	e.logStreams.SetText(e.app.ConfigView.Conns, false)
	e.selectQuery.SetText(e.app.ConfigView.SelectQuery)
}

func (e *EditModal) Stop() {
}

func (e *EditModal) IsModal() bool {
	return true
}

func (m *EditModal) Draw(screen tcell.Screen) {
	// cm.SetRect(x, y, width/2, height/2)
	screenWidth, screenHeight := screen.Size()

	// cm.SetRect(0, 0, screenWidth, screenHeight) // 全屏

	//m.SetRect(screenWidth/4, screenHeight/4, screenWidth/2, screenHeight/2) // 左右

	m.SetRect(screenWidth/8, screenHeight/8, screenWidth/2+screenWidth/4, screenHeight/2+screenHeight/4) // 左右

	m.Box.DrawForSubclass(screen, m)
	x, y, width, height := m.GetInnerRect()
	m.frame.SetRect(x, y, width, height)
	m.frame.Draw(screen)
}

func (m *EditModal) Focus(delegate func(p tview.Primitive)) {
	delegate(m.flex)
}

func (m *EditModal) Blur() {
	m.flex.Blur()
}

func (m *EditModal) HasFocus() bool {
	return m.flex.HasFocus()
}

func (m *EditModal) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return m.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		// Pass mouse events on to the form.
		consumed, capture = m.flex.MouseHandler()(action, event, setFocus)
		if !consumed && action == tview.MouseLeftDown && m.InRect(event.Position()) {
			setFocus(m)
			consumed = true
		}
		return
	})
}

func (m *EditModal) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return m.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		if m.frame.HasFocus() {
			if handler := m.frame.InputHandler(); handler != nil {
				handler(event, setFocus)
				return
			}
		}
	})
}

func (m *EditModal) Query() {

}
