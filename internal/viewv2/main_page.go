package viewv2

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/mawen12/ndx/pkg/histogram"
	"github.com/mawen12/ndx/pkg/timefmt"
	"github.com/rivo/tview"
)

type MainPage struct {
	*tview.Flex
	app *App

	topBar     *TopBar
	histogram  *histogram.Histogram
	logTable   *tview.Table
	statusLine *StatusLine
	cmd        *tview.InputField
}

func NewMainPage(app *App) *MainPage {
	m := &MainPage{
		Flex: tview.NewFlex().SetDirection(tview.FlexRow),
		app:  app,

		topBar:    NewTopBar(app),
		histogram: histogram.NewHistogram(),
		logTable:  tview.NewTable(),
		cmd:       tview.NewInputField(),
	}

	m.histogram.SetBinSize(60).SetDataBinsSnapper(snapDataBinsInChartDots).
		SetXFormat(func(v int) string {
			tz := time.Local

			t := time.Unix(int64(v), 0).In(tz)
			if t.Hour() == 0 && t.Minute() == 0 {
				return t.In(tz).Format(fmt.Sprintf("[yellow]%s[-]", timefmt.MonthDay))
			}
			return t.In(tz).Format(timefmt.HourMinute)
		}).
		SetCursorFormat(func(from int, to *int, width int) string {
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
		}).
		SetXMarker(func(from, to, numChars int) []int {
			tz := time.Local
			return getXMarksForHistogram(tz, from, to, numChars)
		}).
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyTab:

			case tcell.KeyBacktab:

			case tcell.KeyEsc:

			case tcell.KeyRune:
				switch event.Rune() {
				case ':':

				case 'i', 'a':

				}
			}

			return event
		})

	m.logTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			//m.app.SetFocus(m.editBtn)
			return nil
		case tcell.KeyBacktab:
			return nil
		case tcell.KeyEnter:
			return nil
		}
		return event
	})

	m.cmd.SetFocusFunc(func() {

	})

	m.AddItem(m.topBar, 1, 0, true).
		AddItem(nil, 0, 1, false)

	return m
}

func (m *MainPage) Name() string {
	return "main_page"
}

func (m *MainPage) Start() {
	m.topBar.query.SetText(m.app.Config.Pattern)
	m.topBar.timeLabel.SetText(m.app.Config.TimeRange.String())
}

func (m *MainPage) Stop() {

}

func (m *MainPage) Modal() bool {
	return false
}

type TopBar struct {
	*tview.Flex
	app *App

	queryLabel *tview.TextView
	query      *tview.InputField
	timeLabel  *tview.TextView
	editBtn    *tview.Button
}

func NewTopBar(app *App) *TopBar {
	t := &TopBar{
		Flex: tview.NewFlex().SetDirection(tview.FlexColumn),
		app:  app,

		queryLabel: tview.NewTextView(),
		query:      tview.NewInputField(),
		timeLabel:  tview.NewTextView(),
		editBtn:    tview.NewButton("Edit"),
	}

	t.queryLabel.SetText("awk pattern: ").SetDynamicColors(true).SetScrollable(false).SetTextAlign(tview.AlignLeft)

	t.query.
		SetChangedFunc(func(text string) {
			label := "awk pattern: "
			//if text != app.Model.Query {
			//	label = "awk pattern[yellow::b]*[-::-]:"
			//}
			app.QueueUpdateDraw(func() {
				t.queryLabel.SetText(label)
			})
		}).
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyEsc:
				t.app.SetFocus(t.editBtn)
				return nil
			case tcell.KeyTab:
				t.app.SetFocus(t.editBtn)
				return nil
			case tcell.KeyBacktab:
				return nil
			case tcell.KeyEnter:
				return nil
			}
			return event
		})

	t.timeLabel.SetText("-1h").SetScrollable(false).SetTextAlign(tview.AlignCenter).
		SetChangedFunc(func() {
			t.ResizeItem(t.timeLabel, tview.TaggedStringWidth(t.timeLabel.GetLabel()), 0)
		})

	t.editBtn.SetTitleAlign(tview.AlignCenter).
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyEsc:
				t.app.SetFocus(t.editBtn)
				return nil
			case tcell.KeyTab:
				t.app.SetFocus(t.editBtn)
				return nil
			case tcell.KeyBacktab:
				t.app.SetFocus(t.query)
				return nil
			case tcell.KeyEnter:
				return nil
			}
			return event
		})

	t.AddItem(t.queryLabel, 13, 0, false).
		AddItem(nil, 1, 0, false).
		AddItem(t.query, 0, 1, true).
		AddItem(nil, 1, 0, false).
		AddItem(t.timeLabel, 6, 0, false).
		AddItem(nil, 1, 0, false).
		AddItem(t.editBtn, 6, 0, false)

	return t
}

type StatusLine struct {
	*tview.Flex

	left, right *tview.TextView
}

func NewStatusLine() *StatusLine {
	s := &StatusLine{
		Flex: tview.NewFlex(),

		left:  tview.NewTextView(),
		right: tview.NewTextView(),
	}

	s.right.SetTextAlign(tview.AlignRight)

	return s
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
