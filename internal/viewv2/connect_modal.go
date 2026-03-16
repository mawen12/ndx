package viewv2

import (
	"context"
	"fmt"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/mawen12/ndx/internal"
	"github.com/mawen12/ndx/internal/config"
	"github.com/mawen12/ndx/internal/model"
	"github.com/mawen12/ndx/internal/pool"
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
	messageChan <-chan model.Notice
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
	//content := c.app.components.MustGet(internal.ConnectModalContentComponent).(*TextDesc)

	conns, err := config.ParseConns(c.app.ConfigView.Conns)
	if err != nil {
		c.SetText(fmt.Sprintf("[ERROR]: conns(%s) is invalid", c.app.ConfigView.Conns))
		//content.SetText(fmt.Sprintf("[ERROR]: conns(%s) is invalid", c.app.ConfigView.Conns))
		return
	}

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
			case notice, ok := <-c.messageChan:
				if !ok {
					return
				}

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

				c.app.QueueUpdateDraw(func() {
					//content.SetText(cms.String())
					c.SetText(cms.String())
				})
			}
		}
	})

	c.connectTask.Start(conns)
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

//func (cm *ConnectModal) Draw(screen tcell.Screen) {
//	// cm.SetRect(x, y, width/2, height/2)
//	screenWidth, screenHeight := screen.Size()
//
//	// cm.SetRect(0, 0, screenWidth, screenHeight) // 全屏
//
//	cm.SetRect(screenWidth/4, screenHeight/4, screenWidth/2, screenHeight/2) // 左右
//
//	cm.Box.DrawForSubclass(screen, cm)
//	x, y, width, height := cm.GetInnerRect()
//	cm.frame.SetRect(x, y, width, height)
//	cm.frame.Draw(screen)
//}

//func (cm *ConnectModal) Focus(delegate func(p tview.Primitive)) {
//	delegate(cm.form)
//}
//
//func (cm *ConnectModal) Blur() {
//	cm.form.Blur()
//}
//
//func (cm *ConnectModal) HasFocus() bool {
//	return cm.form.HasFocus()
//}

//func (cm *ConnectModal) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
//	return cm.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
//		// Pass mouse events on to the form.
//		consumed, capture = cm.form.MouseHandler()(action, event, setFocus)
//		if !consumed && action == tview.MouseLeftDown && cm.InRect(event.Position()) {
//			setFocus(cm)
//			consumed = true
//		}
//		return
//	})
//}
//
//func (m *ConnectModal) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
//	return m.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
//		if m.frame.HasFocus() {
//			if handler := m.frame.InputHandler(); handler != nil {
//				handler(event, setFocus)
//				return
//			}
//		}
//	})
//}

type ConnectTask struct {
	mu  sync.Mutex
	app *App

	cancel context.CancelFunc

	messageChan chan model.Notice

	startOnce sync.Once
	closeOnce sync.Once
}

func NewConnectTask() *ConnectTask {
	return &ConnectTask{
		messageChan: make(chan model.Notice, 10),
	}
}

func (t *ConnectTask) Reset() {
	t.Close()

	t.mu.Lock()
	defer t.mu.Unlock()

	t.messageChan = make(chan model.Notice, 10)
	t.startOnce = sync.Once{}
	t.closeOnce = sync.Once{}
}

func (t *ConnectTask) Chan() <-chan model.Notice {
	return t.messageChan
}

func (t *ConnectTask) Start(conns config.QueryConns) {
	t.startOnce.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())

		t.mu.Lock()
		t.cancel = cancel
		t.mu.Unlock()

		t.app.Background(func() {
			defer close(t.messageChan)

			p := pool.NewPool(conns)

			callback := func(conn, message string, finished bool) {
				t.messageChan <- model.Notice{Conn: conn, Message: message, Finished: finished}
			}

			err := p.Connect(ctx, callback)
			if err == nil {
				t.app.Pool = p
			}

			// query

			//ret, err := p.Query(ctx, *t.app.Config)
			//if err != nil {
			//	t.messageChan <- model.Notice{Conn: "", Message: err.Error(), Finished: true}
			//}
			//
			//
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
