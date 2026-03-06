package ui

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
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
	prev tview.Primitive
}

func NewCommand(app *App) *Command {
	c := Command{
		InputField: tview.NewInputField(),
		app:        app,
	}

	c.SetInputCapture(c.keyboard)

	c.SetFocusFunc(c.activate)

	c.SetBlurFunc(c.inactivate)

	return &c
}

func (c *Command) Name() string {
	return "cmd"
}

func (c *Command) ShowQueryDuration(duration time.Duration) {
	c.SetText(fmt.Sprintf("Query cost %d", duration.Milliseconds()))
}

func (c *Command) SetFocus(prev tview.Primitive) {
	c.prev = prev
	c.app.SetFocus(c)
}

func (c *Command) keyboard(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEsc:
		c.app.SetFocus(c.prev)
	}
	return event
}

func (c *Command) activate() {
	c.SetFieldStyle(tcell.Style{}.Background(tcell.ColorWhite).Foreground(tcell.ColorBlue).Bold(true))
	c.SetText(":")
}

func (c *Command) inactivate() {
	c.SetFieldStyle(tcell.Style{}.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite).Bold(true))
	c.SetText("")
}
