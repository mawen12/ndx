package view

import (
	"context"

	"github.com/mawen12/ndx/internal/model"
	"github.com/mawen12/ndx/internal/ui"
)

type PageStack struct {
	*ui.Pages

	app *App
}

func NewPageStack() *PageStack {
	return &PageStack{
		Pages: ui.NewPages(),
	}
}

func (p *PageStack) Init(ctx context.Context) (err error) {
	if p.app, err = extractApp(ctx); err != nil {
		return err
	}

	p.AddListener(p)

	return nil
}

func (p *PageStack) StackPushed(c model.Component) {
	c.Start()
	p.app.SetFocus(c)
}

func (p *PageStack) StackPopped(old, new model.Component) {
	old.Stop()
	p.StackTop(new)
}

func (p *PageStack) StackTop(top model.Component) {
	if top == nil {
		return
	}
	top.Start()
	p.app.SetFocus(top)
}
