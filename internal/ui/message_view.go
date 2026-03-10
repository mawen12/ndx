package ui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/mawen12/ndx/internal/model"
	"github.com/mawen12/ndx/pkg/tviews"
	"github.com/rivo/tview"
)

type MessageView struct {
	*tview.Grid
	*model.CycleList

	frame *tview.Frame
	flex  *tview.Flex

	msgFlex *tview.Flex
	text    *tview.TextView

	btnFlex *tview.Flex
	okBtn   *tview.Button
	copyBtn *tview.Button

	app *App
}

func NewMessageView(app *App) *MessageView {
	mv := MessageView{
		Grid:      tview.NewGrid(),
		CycleList: model.NewCycleList(),
		flex:      tviews.NewFlexRow(),
		msgFlex:   tviews.NewFlexRow(),
		text:      tview.NewTextView().SetTextAlign(tview.AlignCenter).SetDynamicColors(true),
		btnFlex:   tviews.NewFlexColumn(),
		okBtn:     tview.NewButton("OK"),
		copyBtn:   tview.NewButton("Copy"),
		app:       app,
	}

	return &mv
}

func (mv *MessageView) Init() {
	mv.handle()

	mv.layout()
}

func (mv *MessageView) handle() {
	mv.PushBack(model.NewCyclePrimitive(mv.okBtn, "ok"))
	mv.PushBack(model.NewCyclePrimitive(mv.copyBtn, "copy"))
}

func (mv *MessageView) keyboard(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyTab:
		p := mv.Next().(tview.Primitive)
		mv.app.SetFocus(p)
	case tcell.KeyBacktab:
		p := mv.Prev().(tview.Primitive)
		mv.app.SetFocus(p)
	case tcell.KeyEnter:
		switch mv.Current().(model.CyclePrimitive).Name {
		case "ok":

		case "copy":

		}
	}

	return event
}

func (mv *MessageView) layout() {
	mv.msgFlex.AddItem(mv.text, 0, 1, true)

	mv.btnFlex.
		AddItem(nil, 0, 1, false).
		AddItem(mv.okBtn, len(mv.okBtn.GetLabel())+2*2, 0, true).
		AddItem(nil, 0, 1, false).
		AddItem(mv.copyBtn, len(mv.copyBtn.GetLabel())+2*2, 0, false).
		AddItem(nil, 0, 1, false)

	mv.flex.
		AddItem(mv.msgFlex, 1, 0, false).
		AddItem(nil, 1, 0, false).
		AddItem(mv.btnFlex, 1, 1, true)

	mv.frame = tview.NewFrame(mv.flex).SetBorders(0, 0, 0, 0, 0, 0)
	mv.frame.SetBorder(true).SetBorderPadding(1, 1, 1, 1)
	mv.frame.SetTitle("Log Query Error")

	mv.AddItem(mv.frame, 1, 1, 1, 1, 0, 0, true)
}

func (mv *MessageView) Show(message string) {
	width, height := calculateSize(message)
	mv.SetColumns(0, width+20, 0).
		SetRows(0, height+10, 0)

	mv.text.SetText(message)
	mv.frame.SetBackgroundColor(tcell.ColorDarkRed)
	mv.text.SetBackgroundColor(tcell.ColorDarkRed)

	mv.Reset()

	mv.app.Main.AddPage("message_view", mv, true, true)

	mv.app.SetFocus(mv.frame)
}

func calculateSize(content string) (int, int) {
	content = strings.TrimSpace(content)
	lines := strings.Split(content, "\n")
	maxLen := 0
	for _, line := range lines {
		if len(line) > maxLen {
			maxLen = len(line)
		}
	}

	return maxLen, len(lines)
}
