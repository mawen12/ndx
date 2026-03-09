package model

import "github.com/mawen12/ndx/pkg/times"

type Query struct {
	ConnString  string
	Pattern     string
	TimeRange   *times.TimeRange
	SelectQuery string
}

func NewQuery(connString, pattern string, timeRange *times.TimeRange, selectQuery string) *Query {
	return &Query{
		ConnString:  connString,
		Pattern:     pattern,
		TimeRange:   timeRange,
		SelectQuery: selectQuery,
	}
}
