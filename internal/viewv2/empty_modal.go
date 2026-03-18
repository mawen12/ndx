package viewv2

import (
	"github.com/rivo/tview"
)

type EmptyModal struct {
	*tview.Modal
}

func NewEmptyModal() *EmptyModal {
	return &EmptyModal{
		Modal: tview.NewModal(),
	}
}

func (c *EmptyModal) Start() {
}

func (c *EmptyModal) Stop() {
}

func (c *EmptyModal) IsModal() bool {
	return true
}
