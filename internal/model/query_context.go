package model

import (
	"github.com/mawen12/ndx/pkg/times"
)

type QueryContext struct {
	Pattern   string
	TimeRange *times.TimeRange

	QueryFunc func()
}

func NewQueryContext() *QueryContext {
	return &QueryContext{
		Pattern:   "",
		TimeRange: times.NewDefaultTimeRange(),
	}
}

func (qc *QueryContext) DoQuery() {
	if qc.QueryFunc != nil {
		qc.QueryFunc()
	}
}
