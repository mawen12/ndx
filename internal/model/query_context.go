package model

import (
	"github.com/mawen12/ndx/pkg/times"
)

type QueryContext struct {
	Pattern   string
	TimeRange *times.TimeRange

	QueryFunc func()
}
