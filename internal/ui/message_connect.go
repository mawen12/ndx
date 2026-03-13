package ui

import (
	"context"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/mawen12/ndx/internal/model"
	"github.com/rivo/tview"
)

type NoticeCh chan model.Notice

type MessageConnect struct {
	*tview.Modal
	app *App
}

func NewMessageConnect(app *App) *MessageConnect {
	md := &MessageConnect{
		Modal: tview.NewModal().
			AddButtons([]string{"OK", "Cancel"}),
		app: app,
	}

	md.SetTitle("Connecting")

	md.SetButtonStyle(tcell.Style{}.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite))
	md.SetButtonActivatedStyle(tcell.Style{}.Foreground(tcell.ColorBlue).Background(tcell.ColorWhite))

	return md
}

func (mc *MessageConnect) Name() string {
	return "message_connect"
}

func (mc *MessageConnect) Init(ctx context.Context) error {
	mc.Modal.SetInputCapture(mc.keyboard)
	return nil
}

func (mc *MessageConnect) Start() {}

func (mc *MessageConnect) Stop() {}

func (mc *MessageConnect) keyboard(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEsc:
		mc.app.Main.Pop()
	}

	return event
}

func (mc *MessageConnect) ShowError(ch NoticeCh, onOK, onCancel func()) {
	mc.Modal.SetDoneFunc(func(index int, label string) {
		switch index {
		case 0:
		case 1:
			onCancel()
			mc.app.Main.Pop()
		}
	})

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		var cms model.ConnectMessages

		for notice := range ch {
			if cms.Connecting == nil {
				cms.Connecting = &model.ConnectMessage{
					Connection: notice.Conn,
					Messages:   make([]string, 10),
				}
			}

			if !notice.Finished {
				cms.Connecting.Messages = append(cms.Connecting.Messages, notice.Message)
			} else {
				if notice.Success {
					cms.Connected = append(cms.Connected, model.KV{Key: cms.Connecting.Connection, Value: ""})
				} else {
					cms.Connected = append(cms.Connected, model.KV{Key: cms.Connecting.Connection, Value: notice.Message})
				}
				cms.Connecting = nil
			}

			mc.SetText(cms.String())
		}
	}()

	//ticker := time.NewTicker(100 * time.Millisecond)
	//
	//go func() {
	//	loadingChars := []string{"|", "/", "-", "\\"}
	//	index := 0
	//
	//	for {
	//		select {
	//		case <-ticker.C:
	//			mc.SetTitle(fmt.Sprintf("Connecting...%s", loadingChars[index]))
	//			index = (index + 1) % len(loadingChars)
	//		}
	//	}
	//}()

	go func() {
		wg.Wait()
		//ticker.Stop()

		mc.SetTitle("Connect Finish")
		mc.SetDoneFunc(func(index int, label string) {
			switch index {
			case 0:
				onOK()
			case 1:
				onCancel()
				mc.app.Main.Pop()
			}
		})
	}()

	mc.app.Main.Push(mc)
}
