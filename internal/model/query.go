package model

type QueryView struct {
	Conns       string
	Pattern     string
	TimeRange   string
	SelectQuery string
}

func (q *QueryView) SetConns(conns string) {
	q.Conns = conns
}

func (q *QueryView) SetPattern(pattern string) {
	q.Pattern = pattern
}

func (q *QueryView) SetTimeRange(timeRange string) {
	q.TimeRange = timeRange
}

func (q *QueryView) SetSelectQuery(selectQuery string) {
	q.SelectQuery = selectQuery
}
