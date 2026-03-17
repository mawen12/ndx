package viewv2

import (
	"context"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/mawen12/ndx/internal"
	"github.com/mawen12/ndx/internal/model"
	"github.com/rivo/tview"
)

type ConnectModal struct {
	//*tview.Box
	//frame *tview.Frame
	//form  *tview.Form
	*tview.Modal

	app         *App
	cancel      context.CancelFunc
	connectTask *ConnectTask
	messageChan <-chan ConnectMessage
}

func NewConnectModal() *ConnectModal {
	c := &ConnectModal{
		Modal: tview.NewModal(),
		//Box:         tview.NewBox(),
		//form:        tview.NewForm(),
		connectTask: NewConnectTask(),
	}

	//c.frame = tview.NewFrame(c.form)

	return c
}

func (c *ConnectModal) Name() internal.PageKey {
	return internal.KeyConnectModal
}

func (c *ConnectModal) Init(ctx context.Context) {
	c.app = extractApp(ctx)
	c.connectTask.app = c.app

	//c.form.AddFormItem(c.app.components.MustGet(internal.ConnectModalContentComponent).(*TextDesc))
	//
	//c.form.AddButton("Cancel", func() {
	//	c.app.Hide()
	//})

	c.AddButtons([]string{"Cancel"})

	c.SetDoneFunc(func(index int, label string) {
		switch index {
		case 0:
			c.app.Hide()
		}
	})

	c.SetTitle("Step 1: Connecting...").SetBorder(true)

	c.SetBackgroundColor(tcell.ColorBlack)
}

func (c *ConnectModal) Start() {
	c.connectTask.Reset()
	c.messageChan = c.connectTask.Chan()

	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel
	var cms model.ConnectMessages

	c.app.Background(func() {
		for {
			select {
			case <-ctx.Done():
				return
			case message, ok := <-c.messageChan:
				if !ok {
					return
				}

				switch {
				case message.Pause:
					if message.Err != nil {
						c.app.QueueUpdateDraw(func() {
							c.SetText(message.Err.Error())
						})
					}
					return
				case message.Title != "":
					c.app.QueueUpdateDraw(func() {
						c.SetTitle(message.Title)
						c.SetText("")
						cms.Connecting = nil
					})
				case message.Connecting != nil:
					if cms.Connecting == nil {
						cms.Connecting = &model.ConnectMessage{
							Connection: message.Connecting.Conn,
							Messages:   make([]string, 10),
						}
					}

					if message.Connecting.Finished {
						if message.Connecting.Success {
							cms.Connected = append(cms.Connected, model.KV{Key: cms.Connecting.Connection, Value: ""})
						} else {
							cms.Connected = append(cms.Connected, model.KV{Key: cms.Connecting.Connection, Value: message.Connecting.Message})
						}
						cms.Connecting = nil
					} else {
						cms.Connecting.Messages = append(cms.Connecting.Messages, message.Connecting.Message)
					}

					c.app.QueueUpdateDraw(func() {
						c.SetText(cms.String())
					})
				default:
					c.app.HideTwice()
					return
				}
			}
		}
	})

	c.connectTask.Start()
}

func (c *ConnectModal) Stop() {
	if c.cancel != nil {
		c.cancel()

		c.connectTask.Close()
	}
}

func (c *ConnectModal) IsModal() bool {
	return true
}

type ConnectMessage struct {
	Title      string
	Connecting *model.Notice
	Pause      bool
	Err        error
}

type ConnectTask struct {
	mu  sync.Mutex
	app *App

	cancel context.CancelFunc

	messageChan chan ConnectMessage

	startOnce sync.Once
	closeOnce sync.Once
}

func NewConnectTask() *ConnectTask {
	return &ConnectTask{
		messageChan: make(chan ConnectMessage, 10),
	}
}

func (t *ConnectTask) Reset() {
	t.Close()

	t.mu.Lock()
	defer t.mu.Unlock()

	t.messageChan = make(chan ConnectMessage, 10)
	t.startOnce = sync.Once{}
	t.closeOnce = sync.Once{}
}

func (t *ConnectTask) Chan() <-chan ConnectMessage {
	return t.messageChan
}

func (t *ConnectTask) Start() {
	t.startOnce.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())

		t.mu.Lock()
		t.cancel = cancel
		t.mu.Unlock()

		t.app.Background(func() {
			defer close(t.messageChan)

			callback := func(conn, message string, success, finsihed bool) {
				t.messageChan <- ConnectMessage{Connecting: &model.Notice{
					Conn:     conn,
					Message:  message,
					Success:  success,
					Finished: finsihed,
				}}
			}

			t.messageChan <- ConnectMessage{Title: "Step 1: Connecting"}
			err := t.app.Connect(ctx, callback)
			if err != nil {
				t.messageChan <- ConnectMessage{Pause: true, Err: err}
				return
			}

			// query
			t.messageChan <- ConnectMessage{Title: "Step 2: Querying..."}

			err = t.app.Query(ctx)
			if err != nil {
				t.messageChan <- ConnectMessage{Pause: true, Err: err}
			}

			t.messageChan <- ConnectMessage{}
		})
	})
}

func (t *ConnectTask) Close() {
	t.closeOnce.Do(func() {
		t.mu.Lock()
		cancel := t.cancel
		t.cancel = nil
		t.mu.Unlock()

		if cancel != nil {
			cancel()
		}
	})
}
