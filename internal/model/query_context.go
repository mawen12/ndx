package model

import (
	"time"
)

type QueryContext struct {
	pattern   string
	timeRange time.Time

	QueryFunc func()
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

func (qc *QueryContext) DoQuery() {
	if qc.QueryFunc != nil {
		qc.QueryFunc()
	}
}
