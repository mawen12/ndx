package viewv2

import (
	"context"
	"log/slog"

	"github.com/gdamore/tcell/v2"
	"github.com/mawen12/ndx/internal"
	"github.com/rivo/tview"
)

type ChannelMessage struct {
	Message  string
	Finished bool
	Success  bool
}

type ChannelModal struct {
	*tview.Modal

	app         *App
	cancel      context.CancelFunc
	messageChan <-chan string
}

func NewChannelModal() *ChannelModal {
	return &ChannelModal{
		Modal: tview.NewModal(),
	}
}

func (c *ChannelModal) Name() internal.PageKey {
	return internal.KeyChannelModal
}

func (c *ChannelModal) Init(ctx context.Context) {
	c.app = extractApp(ctx)

	c.AddButtons([]string{"Cancel"})

	c.SetDoneFunc(func(index int, label string) {
		switch index {
		case 0:
			c.app.Hide()
		}
	})

	c.SetBackgroundColor(tcell.ColorBlack)
}

func (c *ChannelModal) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel

	c.app.Background(func() {
		for {
			select {
			case <-ctx.Done():
				return
			case message, ok := <-c.messageChan:
				if !ok {
					return
				}

				slog.Info("Receive message", "message", message)

				c.app.QueueUpdateDraw(func() {
					c.SetText(message)
				})
			}
		}
	})
}

func (c *ChannelModal) Stop() {
	if c.cancel != nil {
		c.cancel()
	}
}

func (c *ChannelModal) IsModal() bool {
	return true
}

func (c *ChannelModal) SetMessageChan(messageChan <-chan string) {
	c.messageChan = messageChan
}
