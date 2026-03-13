package viewv2

import (
	"fmt"
	"log/slog"

	"github.com/rivo/tview"
)

type Pages struct {
	*tview.Pages

	app   *App
	stack *Stack
}

func NewPages(app *App) *Pages {
	p := &Pages{
		Pages: tview.NewPages(),
		app:   app,
		stack: NewStack(),
	}

	p.stack.AddListener(p)

	return p
}

func (p *Pages) Show(c Component) {
	p.stack.Push(c)
}

func (p *Pages) Hide() {
	p.stack.Pop()
}

// stack listener

func (p *Pages) Pushed(c Component) {
	p.add(c)
	p.show(c)
}

func (p *Pages) Poped(old, new Component) {
	p.delete(old)
	old.Stop()
	p.Top(new)
}

func (p *Pages) Top(c Component) {
	if c != nil {
		p.show(c)
	}
}

func (p *Pages) add(c Component) {
	p.AddPage(componentID(c), c, true, true)
}

func (p *Pages) delete(c Component) {
	p.RemovePage(componentID(c))
}

func (p *Pages) show(c Component) {
	if !c.Modal() {
		p.SwitchToPage(componentID(c))
	}

	c.Start()
	p.app.SetFocus(c)
	slog.Info("Show and focus component", "component", c.Name())
}

func componentID(c Component) string {
	return fmt.Sprintf("%s-%p", c.Name(), c)
}
