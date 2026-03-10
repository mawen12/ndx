package times

import (
	"strings"
	"time"
)

type TimeOrDuration struct {
	Time     time.Time
	Duration time.Duration
}

func NewTimeOrDuration(tz *time.Location, layout, s string) (TimeOrDuration, error) {
	dur, err := time.ParseDuration(s)
	if err == nil {
		return TimeOrDuration{
			Duration: dur,
		}, nil
	}

	t, err := time.ParseInLocation(layout, s, tz)
	if err == nil {
		return TimeOrDuration{Time: t}, nil
	}

	return TimeOrDuration{}, err
}

func (t TimeOrDuration) IsZero() bool {
	return t.Time.IsZero() && t.Duration == 0
}

func (t TimeOrDuration) In(loc *time.Location) TimeOrDuration {
	if t.Time.IsZero() {
		return t
	}

	t.Time = t.Time.In(loc)
	return t
}

func (t *TimeOrDuration) IsAbsolute() bool {
	return !t.Time.IsZero()
}

func (t *TimeOrDuration) AbsoluteTime(relativeTo time.Time) time.Time {
	if relativeTo.IsZero() {
		panic("relativeTo can't be zero")
	}

	if !t.Time.IsZero() {
		return t.Time
	}

	return relativeTo.Add(t.Duration)
}

func (t *TimeOrDuration) Format(layout string) string {
	if !t.Time.IsZero() {
		return t.Time.Format(layout)
	}

	return formatDuration(t.Duration)
}

func (t *TimeOrDuration) String() string {
	if !t.Time.IsZero() {
		return t.Time.String()
	}
	return formatDuration(t.Duration)
}

func formatDuration(dur time.Duration) string {
	ret := dur.String()

	if strings.HasSuffix(ret, "h0m0s") {
		return ret[:len(ret)-4]
	} else if strings.HasSuffix(ret, "m0s") {
		return ret[:len(ret)-2]
	}

	return ret
}
