package ui

import (
	"fmt"

	"github.com/mawen12/ndx/internal/model"
	"github.com/rivo/tview"
)

type Pages struct {
	*tview.Pages
	*model.Stack
}

func NewPages() *Pages {
	p := Pages{
		Pages: tview.NewPages(),
		Stack: model.NewStack(),
	}

	p.AddListener(&p)

	return &p
}

func (p *Pages) Show(c model.Component) {
	p.SwitchToPage(componentID(c))
}

func (p *Pages) addAndShow(c model.Component) {
	p.add(c)
	p.Show(c)
}

func (p *Pages) add(c model.Component) {
	p.AddPage(componentID(c), c, true, true)
}

func (p *Pages) delete(c model.Component) {
	p.RemovePage(componentID(c))
}

func (p *Pages) StackPushed(c model.Component) {
	p.addAndShow(c)
}

func (p *Pages) StackPopped(old, _ model.Component) {
	p.delete(old)
}

func (p *Pages) StackTop(top model.Component) {
	if top == nil {
		return
	}

	p.Show(top)
}

func componentID(c model.Component) string {
	return fmt.Sprintf("%s-%p", c.Name(), c)
}
