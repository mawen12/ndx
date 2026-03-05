package view

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/mawen12/ndx/internal/conn"
	"github.com/rivo/tview"
)

type MainPage struct {
	*tview.Flex

	app *App
}

func NewMainPage(app *App) *MainPage {
	mp := MainPage{
		Flex: tview.NewFlex().SetDirection(tview.FlexRow),
		app:  app,
	}

	top := tview.NewFlex().SetDirection(tview.FlexColumn)
	top.AddItem(app.QueryLabel(), 13, 0, false)
	top.AddItem(nil, 1, 0, false)
	top.AddItem(app.Query(), 0, 1, true)
	top.AddItem(nil, 1, 0, false)
	top.AddItem(app.Time(), 6, 0, false)
	top.AddItem(nil, 1, 0, false)
	top.AddItem(app.Edit(), 6, 0, false)

	mp.AddItem(top, 1, 0, true)
	mp.AddItem(app.Table(), 0, 1, false)
	mp.AddItem(app.StatusLine(), 1, 0, false)
	mp.AddItem(app.Cmd(), 1, 0, false)

	app.Query().AddListener(&mp)

	return &mp
}

func (mp *MainPage) Enter(old, new string) {

}

func (mp *MainPage) Done(pattern string) {
	// 查询数据
	result := mp.app.conn.Exec(context.Background(), pattern, time.Now()).Read()
	if result.Err != nil {
		slog.Error("Query Failed", "pattern", pattern, "error", result.Err)
		return
	}
	slog.Info("Query Success", "pattern", pattern, "records", len(result.Lines), "stat", result.Stat)

	// clear old data
	mp.app.Table().Clear()

	// handle new data
	newTableCell := func(line conn.LineInfo) *tview.TableCell {
		tc := tview.NewTableCell(tview.Escape(line.OriginLine)).
			SetSelectable(true).
			SetAttributes(tcell.AttrBold).
			SetAttributes(tview.AlignLeft).
			SetSelectedStyle(tcell.Style{}.Background(tcell.ColorWhite).Foreground(tcell.ColorBlue)).
			SetStyle(tcell.Style{}.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite))

		tc.SetReference(line)

		return tc
	}

	for i, line := range result.Lines {
		mp.app.Table().SetCell(i, 0, newTableCell(line))
	}

	mp.app.Table().Select(len(result.Lines)-1, 0)
	mp.app.Table().ScrollToEnd()

	// show status line
	slog.Info("query cost", "cost", result.Cost.Milliseconds())

	//mp.app.StatusLine().ShowRight(fmt.Sprintf("%d / %d", mp.app.conn.))
	mp.app.Cmd().SetText(fmt.Sprintf("Query took: %dms", result.Cost.Milliseconds()))

}
