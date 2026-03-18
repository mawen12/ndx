package viewv2

import (
	"context"
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/mawen12/ndx/internal"
)

type MessageModal struct {
	*EmptyModal

	app     *App
	message string
}

func NewMessageModal() *MessageModal {
	return &MessageModal{
		EmptyModal: NewEmptyModal(),
	}
}

func (c *MessageModal) Name() internal.PageKey {
	return internal.KeyMessageModal
}

func (c *MessageModal) Init(ctx context.Context) {
	c.app = extractApp(ctx)

	c.AddButtons([]string{"OK"})

	c.SetDoneFunc(func(index int, label string) {
		c.app.Hide()
	})

	c.SetBackgroundColor(tcell.ColorBlack)
}

func (c *MessageModal) SetMessage(message string) {
	c.SetText(message)
}

func (c *MessageModal) SetErrorMessage(message string) {
	c.SetText(fmt.Sprintf(`[red]ERROR: [-]%s`, message))
}
