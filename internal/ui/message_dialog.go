package ui

import (
	"context"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Message struct {
	Title   string
	Message string
	IsErr   bool
}

type MessageDialog struct {
	*tview.Modal
	app *App
}

func NewMessageDialog(app *App) *MessageDialog {
	md := &MessageDialog{
		Modal: tview.NewModal().
			AddButtons([]string{"OK", "Copy"}),
		app: app,
	}

	md.SetButtonStyle(tcell.Style{}.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite))
	md.SetButtonActivatedStyle(tcell.Style{}.Foreground(tcell.ColorBlue).Background(tcell.ColorWhite))

	return md
}

func (md *MessageDialog) Name() string {
	return "message_dialog"
}

func (md *MessageDialog) Init(ctx context.Context) error {
	md.Modal.SetDoneFunc(func(index int, label string) {
		switch index {
		case 0:
			md.app.Main.Pop()
		case 1:
		}
	})

	md.Modal.SetInputCapture(md.keyboard)
	return nil
}

func (md *MessageDialog) Start() {}

func (md *MessageDialog) Stop() {}

func (md *MessageDialog) keyboard(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEsc:
		md.app.Main.Pop()
	}

	return event
}

func (md *MessageDialog) ShowError(m Message) {
	md.SetTitle(m.Title)
	md.SetText(m.Message)

	if m.IsErr {
		md.SetBackgroundColor(tcell.ColorDarkRed)
		md.Box.SetBackgroundColor(tcell.ColorDarkRed)
		md.Box.SetBorderColor(tcell.ColorWhite)
	} else {
		md.SetBackgroundColor(tcell.ColorBlue)
		md.Box.SetBackgroundColor(tcell.ColorBlue)
		md.Box.SetBorderColor(tcell.ColorWhite)
	}

	md.app.Main.Push(md)
}
