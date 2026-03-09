package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/mawen12/ndx/pkg/histogram"
	"github.com/mawen12/ndx/pkg/timefmt"
)

type Histogram struct {
	*histogram.Histogram

	app *App
}

func NewHistogram(app *App) *Histogram {
	h := Histogram{
		Histogram: histogram.NewHistogram(),
		app:       app,
	}

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

	h.SetInputCapture(h.keyboard)

	h.SetSelectedFunc(h.selected)

	return &h
}

func (h *Histogram) keyboard(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyTab:
		h.app.activateTable()
	case tcell.KeyBacktab:
		h.app.activateEdit()
	case tcell.KeyEsc:
		h.app.activateTable()
	case tcell.KeyRune:
		switch event.Rune() {
		case ':':
			h.app.activateCmd(h)
		case 'i', 'a':
			h.app.activateQuery()
		}
	}

	return event
}

func (h *Histogram) selected(from, to int) {
	panic("not implemented")
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
