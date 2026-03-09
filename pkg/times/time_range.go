package times

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mawen12/ndx/pkg/timefmt"
)

type TimeRange struct {
	From, To TimeOrDuration

	ActualFrom, ActualTo time.Time
	ActualQuery          time.Time
}

func NewDefaultTimeRange() *TimeRange {
	r := TimeRange{}

	r.SetRange(TimeOrDuration{Duration: -1 * time.Hour}, TimeOrDuration{})

	return &r
}

func NewTimeRange(from, to TimeOrDuration) *TimeRange {
	r := TimeRange{}

	r.SetRange(from, to)

	return &r
}

func ParseFromTimeStr(tz *time.Location, timeStr string) (*TimeRange, error) {
	parts := strings.Split(timeStr, " to ")
	if len(parts) == 0 {
		return nil, errors.New("time can't be empty, try -1h")
	}

	var from, to TimeOrDuration
	var err error

	fromStr := parts[0]
	from, err = NewTimeOrDuration(tz, timefmt.MonthDayHourMinute, fromStr)
	if err != nil {
		return nil, fmt.Errorf("invalid 'from' duration: %w", err)
	}

	to = TimeOrDuration{}
	if len(parts) > 1 {
		toStr := parts[1]
		if len(fromStr) > 5 && len(toStr) < 5 {
			toStr = fromStr[:6] + toStr
		}

		var err error
		to, err = NewTimeOrDuration(tz, timefmt.MonthDayHourMinute, toStr)
		if err != nil {
			return nil, fmt.Errorf("invalid 'to' duration: %w", err)
		}
	}

	return NewTimeRange(from, to), nil
}

func (r *TimeRange) SetRange(from, to TimeOrDuration) {
	r.SetFrom(from)
	r.SetTo(to)

	if r.ActualFrom.After(r.ActualTo) {
		r.ActualFrom, r.ActualTo = r.ActualTo, r.ActualFrom
	}
}

func (r *TimeRange) SetFrom(from TimeOrDuration) {
	r.From = from

	// fix the future value
	if !r.From.IsAbsolute() && r.From.Duration > 0 {
		r.From.Duration = -r.From.Duration
	}

	r.ActualFrom = r.From.AbsoluteTime(time.Now())

	r.ActualFrom = truncateCeil(r.ActualFrom, time.Minute)
}

func (r *TimeRange) SetTo(to TimeOrDuration) {
	r.To = to

	// fix the future value
	if !r.To.IsAbsolute() && r.To.Duration > 0 {
		r.To.Duration = -r.To.Duration
	}

	if r.To.IsZero() {
		r.ActualTo = time.Now()
		r.ActualTo = truncateCeil(r.ActualTo, time.Minute)
		r.ActualQuery = time.Time{}
	} else {
		r.ActualTo = r.To.AbsoluteTime(time.Now())
		r.ActualTo = truncateCeil(r.ActualTo, time.Minute)
		r.ActualQuery = r.ActualTo
	}
}

// Spec for edit_view timeLabel display
func (r *TimeRange) Spec() string {
	fromStr := r.From.Format(timefmt.MonthDayHourMinute)
	if r.To.IsZero() {
		return fromStr
	}
	format := timefmt.MonthDayHourMinute
	_, fm, fd := r.From.Time.Date()
	_, tm, td := r.To.Time.Date()
	if fm == tm && fd == td {
		format = timefmt.HourMinute
	}

	return fmt.Sprintf("%s to %s", fromStr, r.To.Format(format))
}

// String for timeLabel display
func (r *TimeRange) String() string {
	rangeDuration := r.ActualTo.Sub(r.ActualFrom)

	if !r.To.IsZero() {
		return fmt.Sprintf("%s to %s (%s)", r.From.Format("Jan2 15:04"), r.To.Format("Jan2 15:04"), formatDuration(rangeDuration))
	} else if r.From.IsAbsolute() {
		return fmt.Sprintf("%s to now (%s)", r.From.Format("Jan2 15:04"), formatDuration(rangeDuration))
	} else {
		return fmt.Sprintf("last %s", TimeOrDuration{Duration: -r.From.Duration})
	}
}

func truncateCeil(t time.Time, dur time.Duration) time.Time {
	t2 := t.Truncate(dur)
	if t2.Equal(t) {
		return t
	}

	return t2.Add(dur)
}
