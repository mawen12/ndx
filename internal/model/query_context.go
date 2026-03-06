package model

import "time"

type QueryContext struct {
	pattern   string
	timeRange time.Time
}

func NewQueryContext() *QueryContext {
	return &QueryContext{}
}

func (qc *QueryContext) Pattern() string {
	return qc.pattern
}

func (qc *QueryContext) SetPattern(pattern string) {
	qc.pattern = pattern
}

func (qc *QueryContext) TimeRange() time.Time {
	return qc.timeRange
}

func (qc *QueryContext) SetTimeRange(tr time.Time) {
	qc.timeRange = tr
}
