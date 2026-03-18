package viewv2

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/mawen12/ndx/internal"
	"github.com/mawen12/ndx/internal/model"
	"github.com/mawen12/ndx/pkg/histogram"
	"github.com/mawen12/ndx/pkg/timefmt"
	"github.com/rivo/tview"
)

type Components struct {
	holder *model.Holder[internal.ComponentKey, model.Component]

	app *App
}

func NewComponents() *Components {
	return &Components{
		holder: model.NewHolder[internal.ComponentKey, model.Component](),
	}
}

func (c *Components) Init(ctx context.Context) {
	c.app = extractApp(ctx)

	c.holder.Init(ctx)
}

func (c *Components) Add(cc model.Component) {
	c.holder.Add(cc.Name(), cc)
}

func (c *Components) MustGet(name internal.ComponentKey) model.Component {
	return c.holder.MustGet(name)
}

const (
	queryLabelMatch    = "awk pattern: "
	queryLabelMismatch = "awk pattern[yellow::b]*[-::-]:"
)

type TextDesc struct {
	*tview.TextView
	name     internal.ComponentKey
	initFunc func(t *TextDesc, ctx context.Context)
}

func NewTextDesc(text string, name internal.ComponentKey, initFunc func(t *TextDesc, ctx context.Context)) *TextDesc {
	return &TextDesc{
		TextView: tview.NewTextView().SetText(text).SetDynamicColors(true).SetScrollable(false),
		name:     name,
		initFunc: initFunc,
	}
}

func (t *TextDesc) Name() internal.ComponentKey {
	return t.name
}

func (t *TextDesc) Init(ctx context.Context) {
	if t.initFunc != nil {
		t.initFunc(t, ctx)
	}
}

type TextInput struct {
	*tview.InputField
	name     internal.ComponentKey
	initFunc func(t *TextInput, ctx context.Context)
}

func NewTextInput(name internal.ComponentKey, initFunc func(t *TextInput, ctx context.Context)) *TextInput {
	return &TextInput{
		InputField: tview.NewInputField(),
		name:       name,
		initFunc:   initFunc,
	}
}

func (t *TextInput) Name() internal.ComponentKey {
	return t.name
}

func (t *TextInput) Init(ctx context.Context) {
	if t.initFunc != nil {
		t.initFunc(t, ctx)
	}

	t.SetFocusFunc(func() {
		t.SetFieldStyle(tcell.StyleDefault.
			Foreground(tcell.ColorBlue).
			Background(tcell.ColorWhite).
			Bold(true))
	})

	t.SetBlurFunc(func() {
		t.SetFieldStyle(tcell.StyleDefault.
			Foreground(tcell.ColorWhite).
			Background(tcell.ColorBlue))
	})
}

type Query struct {
	*tview.InputField
}

func NewQuery() *Query {
	return &Query{
		InputField: tview.NewInputField(),
	}
}

func (q *Query) Name() internal.ComponentKey {
	return internal.QueryComponent
}

func (q *Query) Init(ctx context.Context) {
	app := extractApp(ctx)

	q.SetFocusFunc(func() {
		q.SetFieldStyle(tcell.StyleDefault.
			Foreground(tcell.ColorBlue).
			Background(tcell.ColorWhite).
			Bold(true))
	})

	q.SetBlurFunc(func() {
		q.SetFieldStyle(tcell.StyleDefault.
			Foreground(tcell.ColorWhite).
			Background(tcell.ColorBlue))
	})

	q.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			app.SetFocus(app.components.MustGet(internal.EditBtnComponent))
			return nil
		case tcell.KeyBacktab:
			app.SetFocus(app.components.MustGet(internal.TableComponent))
			return nil
		case tcell.KeyEnter:
			app.Config.Pattern = q.GetText()
			app.Query(context.Background())
			app.Render()
			return nil
		}
		return event
	})

	queryDesc := app.components.MustGet(internal.QueryLabelComponent).(*TextDesc)

	q.SetChangedFunc(func(text string) {
		if text == app.Config.Pattern {
			app.QueueUpdateDraw(func() {
				queryDesc.SetText(queryLabelMatch)
			})
		} else {
			app.QueueUpdateDraw(func() {
				queryDesc.SetText(queryLabelMismatch)
			})
		}
	})
}

type TimeLabel struct {
	*tview.TextView
}

func NewTimeLabel() *TimeLabel {
	return &TimeLabel{
		TextView: tview.NewTextView().SetScrollable(false).SetTextAlign(tview.AlignCenter),
	}
}

func (t *TimeLabel) Name() internal.ComponentKey {
	return internal.TimeLabelComponent
}

func (t *TimeLabel) Init(ctx context.Context) {
	app := extractApp(ctx)

	t.SetText(app.Config.TimeRange.String())
}

type EditBtn struct {
	*tview.Button
}

func NewEditBtn() *EditBtn {
	return &EditBtn{
		Button: tview.NewButton("Edit"),
	}
}

func (e *EditBtn) Name() internal.ComponentKey {
	return internal.EditBtnComponent
}

func (e *EditBtn) Init(ctx context.Context) {
	app := extractApp(ctx)

	e.SetStyle(tcell.StyleDefault.
		Foreground(tcell.ColorWhite).
		Background(tcell.ColorBlue).
		Bold(false))

	e.SetActivatedStyle(tcell.StyleDefault.
		Foreground(tcell.ColorBlue).
		Background(tcell.ColorWhite).
		Bold(true))

	e.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			app.SetFocus(app.components.MustGet(internal.HistogramComponent))
			return nil
		case tcell.KeyBacktab:
			app.SetFocus(app.components.MustGet(internal.QueryComponent))
			return nil
		}
		return event
	})

	e.SetSelectedFunc(func() {
		app.Show(internal.KeyEditModal)
	})
}

type Histogram struct {
	*histogram.Histogram
}

func NewHistogram() *Histogram {
	return &Histogram{
		Histogram: histogram.NewHistogram(),
	}
}

func (h *Histogram) Name() internal.ComponentKey {
	return internal.HistogramComponent
}

func (h *Histogram) Init(ctx context.Context) {
	app := extractApp(ctx)

	h.SetBinSize(60)

	h.SetXFormat(func(v int) string {
		tz := time.Local

		t := time.Unix(int64(v), 0).In(tz)
		if t.Hour() == 0 && t.Minute() == 0 {
			return t.In(tz).Format(fmt.Sprintf("[yellow]%s[-]", timefmt.MonthDay))
		}
		return t.In(tz).Format(timefmt.HourMinute)
	})

	h.SetCursorFormat(func(from int, to *int, width int) string {
		tz := time.Local
		fromTime := time.Unix(int64(from), 0).In(tz)

		if to == nil {
			return fromTime.In(tz).Format(timefmt.MonthDayHourMinute)
		}

		toTime := time.Unix(int64(*to), 0).In(tz)

		formatter := func(t time.Time) string {
			return t.In(tz).Format(timefmt.MonthDayHourMinute)
		}

		return fmt.Sprintf("%s - %s (%s)", formatter(fromTime), formatter(toTime), strings.TrimSuffix(toTime.Sub(fromTime).String(), "0s"))
	})

	h.SetXMarker(func(from, to, numChars int) []int {
		tz := time.Local
		return getXMarksForHistogram(tz, from, to, numChars)
	})

	h.SetDataBinsSnapper(snapDataBinsInChartDots)

	h.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			app.SetFocus(app.components.MustGet(internal.TableComponent))
		case tcell.KeyBacktab:
			app.SetFocus(app.components.MustGet(internal.EditBtnComponent))
		case tcell.KeyEsc:
			app.SetFocus(app.components.MustGet(internal.TableComponent))
		case tcell.KeyRune:
			switch event.Rune() {
			case ':':
				cmd := app.components.MustGet(internal.CmdComponent).(*Cmd)
				cmd.SetPrev(h)
				app.SetFocus(app.components.MustGet(internal.CmdComponent))
			case 'i', 'a':
				app.SetFocus(app.components.MustGet(internal.QueryComponent))
			}
		}

		return event
	})

	h.SetSelectedFunc(func(from, to int) {
		panic("histogram selected not implemented")
	})

}

const (
	minute = time.Minute
	hour   = time.Hour
	day    = time.Hour * 24
	month  = day * 30
	year   = day * 365
)

var snaps = []time.Duration{
	minute * 1,
	minute * 2,
	minute * 5,
	minute * 10,
	minute * 15,
	minute * 20,
	minute * 30,
	hour * 1,
	hour * 2,
	hour * 3,
	hour * 6,
	hour * 12,
	day * 1,
	day * 2,
	day * 7,
	month * 1,
	year * 1,
}

func getXMarksForHistogram(tz *time.Location, from, to int, numChars int) []int {
	const minCharsDistanceBetweenMarks = 15
	numMarks := numChars / minCharsDistanceBetweenMarks

	fromTime := time.Unix(int64(from), 0).In(tz)
	toTime := time.Unix(int64(to), 0).In(tz)

	marksTime := getXMarksForTimeRange(tz, fromTime, toTime, numMarks)
	ret := make([]int, 0, len(marksTime))
	for _, v := range marksTime {
		ret = append(ret, int(v.Unix()))
	}
	return ret
}

func getXMarksForTimeRange(tz *time.Location, from, to time.Time, maxNumMarks int) []time.Time {
	if !from.Before(to) || maxNumMarks <= 0 {
		return nil
	}

	duration := to.Sub(from)
	step := chooseStep(duration, maxNumMarks)
	if step == 0 {
		return nil
	}

	start := truncateAlignedToMidnight(from, step, tz)
	if start.Before(from) {
		start = start.Add(step)
	}

	var marks []time.Time
	for t := start; !t.After(to); t = t.Add(step) {
		marks = append(marks, t)
		if len(marks) >= maxNumMarks {
			break
		}
	}
	return marks
}

func chooseStep(duration time.Duration, maxNumMarks int) time.Duration {
	for _, step := range snaps {
		if int(duration/step) <= maxNumMarks {
			return step
		}
	}

	return snaps[len(snaps)-1]
}

func truncateAlignedToMidnight(t time.Time, d time.Duration, loc *time.Location) time.Time {
	t = t.In(loc)

	_, offset := t.Zone()
	if offset == 0 {
		return t.Truncate(d)
	}

	midnight := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)

	sinceMidnight := t.Sub(midnight)
	truncatedSinceMidnight := sinceMidnight.Truncate(d)
	return midnight.Add(truncatedSinceMidnight)
}

func snapDataBinsInChartDots(dataBinsInChartDot int) int {
	for _, snap := range snaps {
		minutes := int(snap / minute)

		if dataBinsInChartDot <= minutes {
			return minutes
		}
	}

	return int(snaps[len(snaps)-1] / minute)
}

type Table struct {
	*tview.Table
}

func NewTable() *Table {
	return &Table{
		Table: tview.NewTable(),
	}
}

func (t *Table) Name() internal.ComponentKey {
	return internal.TableComponent
}

func (t *Table) Init(ctx context.Context) {
	app := extractApp(ctx)

	t.SetFocusFunc(func() {
		app.QueueUpdateDraw(func() {
			t.SetSelectable(true, false)
		})
	})

	t.SetBlurFunc(func() {
		app.QueueUpdateDraw(func() {
			t.SetSelectable(false, false)
		})
	})

	t.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			app.SetFocus(app.components.MustGet(internal.QueryComponent))
		case tcell.KeyBacktab:
			app.SetFocus(app.components.MustGet(internal.HistogramComponent))
		default:
			switch event.Rune() {
			case 'i':
				app.SetFocus(app.components.MustGet(internal.QueryComponent))
			case ':':
				cmd := app.components.MustGet(internal.CmdComponent).(*Cmd)
				cmd.SetPrev(t)
				app.SetFocus(cmd)
			}
		}
		return event
	})
}

func (t *Table) ShowLogs(lines []model.LogLine) {
	t.Clear()

	newTableCellFunc := func(line model.LogLine) []*tview.TableCell {
		lines := strings.Split(line.OriginalLine(), "\000")
		tcs := make([]*tview.TableCell, len(lines))
		for i, l := range lines {
			tc := tview.NewTableCell(tview.Escape(l)).
				SetSelectable(true).
				SetAttributes(tcell.AttrBold).
				SetAttributes(tview.AlignLeft).
				SetSelectedStyle(tcell.Style{}.Background(tcell.ColorWhite).Foreground(tcell.ColorBlue)).
				SetStyle(tcell.Style{}.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite))

			tc.SetReference(line)

			tcs[i] = tc
		}

		return tcs
	}

	var rowIndex int
	for _, line := range lines {
		for _, cell := range newTableCellFunc(line) {
			t.SetCell(rowIndex, 0, cell)
			rowIndex++
		}
	}

	t.Select(len(lines)-1, 0)
	t.ScrollToEnd()
}

type StatusLine struct {
	*tview.Flex
}

func NewStatusLine() *StatusLine {
	return &StatusLine{
		Flex: tview.NewFlex(),
	}
}

func (s *StatusLine) Name() internal.ComponentKey {
	return internal.StatusLineComponent
}

func (s *StatusLine) Init(ctx context.Context) {
	app := extractApp(ctx)

	s.AddItem(app.components.MustGet(internal.StatusLineLeftComponent), 0, 1, false).
		AddItem(nil, 1, 0, false).
		AddItem(app.components.MustGet(internal.StatusLineRightComponent), 30, 0, false)
}

type Cmd struct {
	*tview.InputField

	prev tview.Primitive
}

func NewCmd() *Cmd {
	return &Cmd{
		InputField: tview.NewInputField(),
	}
}

func (c *Cmd) Name() internal.ComponentKey {
	return internal.CmdComponent
}

func (c *Cmd) Init(ctx context.Context) {
	app := extractApp(ctx)

	c.SetFocusFunc(func() {
		app.QueueUpdateDraw(func() {
			c.SetFieldStyle(tcell.Style{}.Background(tcell.ColorWhite).Foreground(tcell.ColorBlue).Bold(true))
			c.SetText(":")
		})
	})

	c.SetBlurFunc(func() {
		app.QueueUpdateDraw(func() {
			c.SetFieldStyle(tcell.Style{}.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite).Bold(false))
			c.SetText("")
		})
	})

	c.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			app.SetFocus(c.prev)
			return nil
		}
		return event
	})
}

func (c *Cmd) SetPrev(p tview.Primitive) {
	c.prev = p
}

func (c *Cmd) ShowQueryDuration(duration time.Duration) {
	c.SetText(fmt.Sprintf("Query cost %dms", duration.Milliseconds()))
}

type LogStreamsInput struct {
	*tview.TextArea
}

func NewLogStreamsInput() *LogStreamsInput {
	return &LogStreamsInput{
		TextArea: tview.NewTextArea().SetSize(0, 0).SetWrap(true).SetWordWrap(false),
	}
}

func (l *LogStreamsInput) Name() internal.ComponentKey {
	return internal.EditViewLogStreamComponent
}

func (l *LogStreamsInput) Init(ctx context.Context) {
	app := extractApp(ctx)

	l.SetTextStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlue))

	l.SetFocusFunc(func() {
		app.QueueUpdateDraw(func() {
			l.SetTextStyle(tcell.StyleDefault.Foreground(tcell.ColorBlue).Background(tcell.ColorWhite).Bold(true))
		})
	})

	l.SetBlurFunc(func() {
		app.QueueUpdateDraw(func() {
			l.SetTextStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlue))
		})
	})

	l.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			app.Hide()
			return nil
		case tcell.KeyTab:
			app.SetFocus(app.components.MustGet(internal.EditViewSelectQueryComponent))
			return nil
		case tcell.KeyBacktab:
			app.SetFocus(app.components.MustGet(internal.EditViewQueryComponent))
			return nil
		case tcell.KeyEnter:
			app.Show(internal.KeyConnectModal)
			return nil
		}
		return event
	})

	l.SetText(app.Config.Origin, false)

	l.SetChangedFunc(func() {
		app.ConfigView.SetConns(l.GetText())
	})
}
