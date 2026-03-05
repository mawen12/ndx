package ui

import (
	"log/slog"

	"github.com/gdamore/tcell/v2"
	"github.com/mawen12/ndx/internal/model"
	"github.com/rivo/tview"
)

/*
Command 命令行组件，格式为 :开头

	h / help
		help 命令，弹出 dialog 框

	time <time>
		将该时间应用到条件查询中

	w / write
		将日志保存到文件中，路径为 /tmp/last_ndx

	set

	q / quit
		退出

	reconnect
		重新连接

	disconnect
		断开拦截

	refresh!
		强制刷新

	querydebug / qdebug / debug

	version / about
		展示版本信息
*/
type Command struct {
	*tview.InputField

	app  *App
	prev model.Focusable
}

func NewCommand(app *App) *Command {
	c := Command{
		InputField: tview.NewInputField(),
		app:        app,
	}

	c.SetInputCapture(c.eventHandle)

	return &c
}

func (c *Command) Name() string {
	return "cmd"
}

func (c *Command) eventHandle(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEsc:
		c.UnFocus()
		c.prev.SetFocus(c)
	}
	return event
}

func (c *Command) SetFocus(prev model.Focusable) {
	slog.Info("into command", "prev", prev.Name())
	c.prev = prev
	c.app.SetFocus(c)
	c.SetText(":")
}

func (c *Command) UnFocus() {
	c.SetText("")
}
