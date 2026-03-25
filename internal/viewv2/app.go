package viewv2

import (
	"context"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/mawen12/ndx/internal"
	"github.com/mawen12/ndx/internal/config"
	"github.com/mawen12/ndx/internal/model"
	"github.com/mawen12/ndx/internal/pool"
	"github.com/mawen12/ndx/pkg/times"
	"github.com/rivo/tview"
)

type App struct {
	*tview.Application
	Width, Height int

	pages      *Pages
	components *Components

	Config     *config.Query
	ConfigView *model.QueryView
	Pool       *pool.Pool
	Result     pool.MergedResult

	ShouldRender bool
	routines     atomic.Int64
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
				//app.Show(internal.KeyConnectModal)

				query, err := app.Parse()
				if err != nil {
					app.pages.MustGet(internal.KeyMessageModal).(*MessageModal).SetErrorMessage(err.Error())
					app.Show(internal.KeyMessageModal)
					return nil
				}

				if err = app.Connect(context.Background(), *query); err != nil {
					app.pages.MustGet(internal.KeyMessageModal).(*MessageModal).SetErrorMessage(err.Error())
					app.Show(internal.KeyMessageModal)
					return nil
				}

				app.SetQuery(query)

				if err = app.Query(context.Background(), false); err != nil {
					app.pages.MustGet(internal.KeyMessageModal).(*MessageModal).SetErrorMessage(err.Error())
					app.Show(internal.KeyMessageModal)
					return nil
				}
				app.Hide()
				app.Render()
				return nil
			}
			return event
		})

		t.SetChangedFunc(app.ConfigView.SetTimeRange)

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

		t.SetChangedFunc(app.ConfigView.SetPattern)

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

		t.SetChangedFunc(app.ConfigView.SetSelectQuery)

		t.SetText(app.Config.SelectQuery)
	}))

	app.components.Add(NewTextDesc("", internal.ConnectModalContentComponent, nil))

	// init component
	app.components.Init(ctx)

	// register page/modal
	app.pages.Add(NewMainPage())
	app.pages.Add(NewEditModal())
	app.pages.Add(NewConnectModal())
	app.pages.Add(NewChannelModal())
	app.pages.Add(NewMessageModal())

	// init page
	app.pages.Init(ctx)
}

func (app *App) Run() error {
	app.SetRoot(app.pages, true)
	app.Show(internal.KeyMainPage)
	app.Show(internal.KeyEditModal)

	app.Application.SetBeforeDrawFunc(app.updateSize)

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
	app.QueueUpdateDraw(func() {
		app.pages.Show(name)
	})
}

func (app *App) Hide() {
	app.QueueUpdateDraw(func() {
		app.pages.Hide()
	})
}

func (app *App) HideTwice() {
	app.QueueUpdateDraw(func() {
		app.pages.Hide()
		app.pages.Hide()
	})
}

func (app *App) Background(handle func()) {
	go func() {
		app.routines.Add(1)
		defer app.routines.Add(-1)

		handle()
	}()
}

func (app *App) SetQuery(query *config.Query) {
	app.Config = query

	app.pages.MustGet(internal.KeyMainPage).(*MainPage).RefreshQuery()
}

func (app *App) Parse() (query *config.Query, err error) {
	query = &config.Query{}
	slog.Info("Parsing the config view", "timeRange", app.ConfigView.TimeRange, "query", app.ConfigView.Pattern, "conns", app.ConfigView.Conns, "selectQuery", app.ConfigView.SelectQuery)

	query.TimeRange, err = times.ParseFromTimeStr(time.UTC, app.ConfigView.TimeRange)
	if err != nil {
		return nil, err
	}

	query.Conns, err = config.ParseConns(app.ConfigView.Conns)
	if err != nil {
		return nil, err
	}

	query.Origin = app.ConfigView.Conns
	query.Pattern = app.ConfigView.Pattern
	query.SelectQuery = app.ConfigView.SelectQuery
	return
}

func (app *App) Connect(ctx context.Context, query config.Query) error {
	if app.Pool != nil {
		app.Pool.Close()
	}

	app.Pool = pool.NewPool(query.Conns)

	return app.Pool.Connect(ctx)
}

func (app *App) Query(ctx context.Context, loadEarlier bool) error {
	if app.Pool == nil {
		panic("app query not implemented")
	}

	queryContext := model.QueryContext{
		Pattern: app.Config.Pattern,
		From:    app.Config.TimeRange.ActualFrom,
		To:      app.Config.TimeRange.ActualQuery,
	}

	// read first record
	if loadEarlier && len(app.Result.Lines) > 0 {
		queryContext.LineUtil = app.Result.Lines[0].LogNumber()
	}

	ret, err := app.Pool.Query(ctx, queryContext)
	if err != nil {
		return err
	}

	if loadEarlier {
		app.Result.Lines = append(ret.Lines, app.Result.Lines...)
		app.Result.Duration = ret.Duration
	} else {
		app.Result = ret
	}

	app.ShouldRender = true
	return nil
}

func (app *App) Render() {
	histogram := app.components.MustGet(internal.HistogramComponent).(*Histogram)
	table := app.components.MustGet(internal.TableComponent).(*Table)
	cmd := app.components.MustGet(internal.CmdComponent).(*Cmd)

	app.QueueUpdateDraw(func() {
		histogram.SetRange(int(app.Config.TimeRange.ActualFrom.Unix()), int(app.Config.TimeRange.ActualTo.Unix()))
		histogram.SetData(app.Result.Stat)

		//table.ShowLogs(app.Result.Lines)
		table.ShowLogsV2(app.Result.Lines)

		cmd.ShowQueryDuration(app.Result.Duration)

		app.SetFocus(table)
	})
}

func (app *App) updateSize(screen tcell.Screen) bool {
	app.Width, app.Height = screen.Size()
	return false
}

func extractApp(ctx context.Context) *App {
	app, ok := ctx.Value(internal.KeyApp).(*App)
	if !ok {
		panic("no application found in context")
	}
	return app
}
