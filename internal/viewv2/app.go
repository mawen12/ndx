package viewv2

import (
	"context"
	"sync/atomic"

	"github.com/gdamore/tcell/v2"
	"github.com/mawen12/ndx/internal"
	"github.com/mawen12/ndx/internal/config"
	"github.com/mawen12/ndx/internal/model"
	"github.com/mawen12/ndx/internal/pool"
	"github.com/rivo/tview"
)

type App struct {
	*tview.Application

	pages      *Pages
	components *Components

	Config     *config.Query
	ConfigView *model.QueryView
	Pool       *pool.Pool
	Result     pool.MergedResult

	routines atomic.Int64
}

func NewApp(config *config.Query) *App {
	app := &App{
		Application: tview.NewApplication(),
		pages:       NewPages(),
		components:  NewComponents(),
		Config:      config,
		ConfigView: &model.QueryView{
			Conns:       config.Origin,
			Pattern:     config.Pattern,
			TimeRange:   config.TimeRange.Spec(),
			SelectQuery: config.SelectQuery,
		},
	}

	return app
}

func (app *App) Init() {
	ctx := context.WithValue(context.Background(), internal.KeyApp, app)

	// register component
	app.components.Add(NewTextDesc(queryLabelMatch, internal.QueryLabelComponent, nil))
	app.components.Add(NewQuery())
	app.components.Add(NewTimeLabel())
	app.components.Add(NewEditBtn())
	app.components.Add(NewHistogram())
	app.components.Add(NewTable())
	app.components.Add(NewTextDesc("", internal.StatusLineLeftComponent, nil))
	app.components.Add(NewTextDesc("", internal.StatusLineRightComponent, func(t *TextDesc, ctx context.Context) {
		t.SetTextAlign(tview.AlignRight)
	}))
	app.components.Add(NewStatusLine())
	app.components.Add(NewCmd())

	app.components.Add(NewTextDesc(timeLabelText, internal.EditViewTimeLabelComponent, nil))
	app.components.Add(NewTextInput(internal.EditViewTimeComponent, func(t *TextInput, ctx context.Context) {
		t.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyEsc:
				app.Hide()
				return nil
			case tcell.KeyTab:
				app.SetFocus(app.components.MustGet(internal.EditViewQueryComponent))
				return nil
			case tcell.KeyBacktab:
				app.SetFocus(app.components.MustGet(internal.EditViewSelectQueryComponent))
				return nil
			case tcell.KeyEnter:
				app.Show(internal.KeyConnectModal)
				return nil
			}
			return event
		})

		t.SetText(app.Config.TimeRange.Spec())
	}))
	app.components.Add(NewTextDesc(queryLabelText, internal.EditViewQueryLabelComponent, nil))
	app.components.Add(NewTextInput(internal.EditViewQueryComponent, func(t *TextInput, ctx context.Context) {
		t.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyEsc:
				app.Hide()
				return nil
			case tcell.KeyTab:
				app.SetFocus(app.components.MustGet(internal.EditViewLogStreamComponent))
				return nil
			case tcell.KeyBacktab:
				app.SetFocus(app.components.MustGet(internal.EditViewTimeComponent))
				return nil
			case tcell.KeyEnter:
				app.Show(internal.KeyConnectModal)
				return nil
			}
			return event
		})

		t.SetText(app.Config.Pattern)
	}))
	app.components.Add(NewTextDesc(logStreamLabelText, internal.EditViewLogStreamLabelComponent, nil))
	app.components.Add(NewLogStreamsInput())
	app.components.Add(NewTextDesc(selectQueryLabelText, internal.EditViewSelectQueryLabelComponent, nil))
	app.components.Add(NewTextInput(internal.EditViewSelectQueryComponent, func(t *TextInput, ctx context.Context) {
		t.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyEsc:
				app.Hide()
				return nil
			case tcell.KeyTab:
				app.SetFocus(app.components.MustGet(internal.EditViewTimeComponent))
				return nil
			case tcell.KeyBacktab:
				app.SetFocus(app.components.MustGet(internal.EditViewLogStreamComponent))
				return nil
			case tcell.KeyEnter:
				app.Show(internal.KeyConnectModal)
				return nil
			}
			return event
		})

		t.SetText(app.Config.SelectQuery)
	}))

	app.components.Add(NewTextDesc("", internal.ConnectModalContentComponent, nil))

	// init component
	app.components.Init(ctx)

	// register page/modal
	app.pages.Add(NewMainPage())
	app.pages.Add(NewEditModal())
	app.pages.Add(NewConnectModal())

	// init page
	app.pages.Init(ctx)
}

func (app *App) Run() error {
	app.SetRoot(app.pages, true)
	app.Show(internal.KeyMainPage)

	if err := app.Application.Run(); err != nil {
		return err
	}

	return nil
}

func (app *App) QueueUpdate(f func()) {
	if app.Application == nil {
		return
	}

	app.Background(func() {
		app.Application.QueueUpdate(f)
	})
}

func (app *App) QueueUpdateDraw(f func()) {
	if app.Application == nil {
		return
	}

	app.Background(func() {
		app.Application.QueueUpdateDraw(f)
	})
}

func (app *App) Close() error {
	app.Application.Stop()
	return nil
}

func (app *App) Show(name internal.PageKey) {
	app.pages.Show(name)
}

func (app *App) Hide() {
	app.pages.Hide()
}

func (app *App) Background(handle func()) {
	go func() {
		app.routines.Add(1)
		defer app.routines.Add(-1)

		handle()
	}()
}

func (app *App) Connect(ctx context.Context, callback func(conn, message string, finished bool)) error {
	conns, err := config.ParseConns(app.ConfigView.Conns)
	if err != nil {
		panic("app connect not implemented")
		return err
	}

	app.Pool = pool.NewPool(conns)

	err = app.Pool.Connect(ctx, callback)
	if err != nil {
		//t.app.Pool = p
		panic("app connect not implemented")
		return err
	}
	return nil
}

func (app *App) Query(ctx context.Context) error {
	if app.Pool == nil {
		panic("app query not implemented")
	}

	ret, err := app.Pool.Query(ctx, *app.Config)
	if err != nil {
		return err
	}

	app.Result = ret
	return nil
}

func extractApp(ctx context.Context) *App {
	app, ok := ctx.Value(internal.KeyApp).(*App)
	if !ok {
		panic("no application found in context")
	}
	return app
}
