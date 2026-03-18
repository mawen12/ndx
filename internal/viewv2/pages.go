package viewv2

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mawen12/ndx/internal"
	"github.com/mawen12/ndx/internal/model"
	"github.com/rivo/tview"
)

type Pages struct {
	*tview.Pages

	stack  *model.Stack
	holder *model.Holder[internal.PageKey, model.Page]

	app *App
}

func NewPages() *Pages {
	p := &Pages{
		Pages:  tview.NewPages(),
		stack:  model.NewStack(),
		holder: model.NewHolder[internal.PageKey, model.Page](),
	}

	p.stack.AddListener(p)

	return p
}

func (p *Pages) Init(ctx context.Context) {
	p.app = extractApp(ctx)

	p.holder.Init(ctx)
}

func (p *Pages) Add(c model.Page) {
	p.holder.Add(c.Name(), c)
}

func (p *Pages) MustGet(name internal.PageKey) model.Page {
	return p.holder.MustGet(name)
}

func (p *Pages) Show(name internal.PageKey) {
	p.stack.Push(p.holder.MustGet(name))
}

func (p *Pages) Hide() {
	p.stack.Pop()
}

// stack listener

func (p *Pages) Pushed(c model.Page) {
	p.addPage(c)
	p.showPage(c)
}

func (p *Pages) Popped(old, new model.Page) {
	p.deletePage(old)
	old.Stop()
	p.Top(new)
}

func (p *Pages) Top(c model.Page) {
	if c != nil {
		p.showPage(c)
	}
}

// internal method

func (p *Pages) addPage(c model.Page) {
	p.AddPage(pageID(c), c, true, true)
}

func (p *Pages) deletePage(c model.Page) {
	p.RemovePage(pageID(c))
}

func (p *Pages) showPage(c model.Page) {
	if !c.IsModal() {
		p.SwitchToPage(pageID(c))
	}

	c.Start()
	p.app.SetFocus(c)
	slog.Info("Show and focus component", "component", c.Name())
}

func pageID(c model.Page) string {
	return fmt.Sprintf("%s-%p", c.Name(), c)
}
