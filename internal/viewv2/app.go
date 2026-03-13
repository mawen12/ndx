package viewv2

import (
	"github.com/mawen12/ndx/internal/config"
	"github.com/rivo/tview"
)

type App struct {
	*tview.Application
	*Pages

	Config *config.Query

	MainPage *MainPage
}

func NewApp(config *config.Query) *App {
	app := &App{
		Application: tview.NewApplication(),

		Config: config,
	}

	app.Pages = NewPages(app)
	//app.MainPage = NewMainPage(app)

	return app
}

func (app *App) Run() error {
	app.SetRoot(app.Pages, true)
	app.Show(app.MainPage)

	if err := app.Application.Run(); err != nil {
		return err
	}

	return nil
}

func (app *App) Close() error {
	app.Application.Stop()
	return nil
}
