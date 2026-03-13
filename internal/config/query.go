package config

import (
	"strconv"
	"strings"

	"github.com/mawen12/ndx/pkg/times"
)

type QueryConn struct {
	Origin   string
	Scheme   string
	Host     string
	Port     uint16
	User     string
	Password string
	LogFile  string
}

func (qc *QueryConn) String() string {
	var sb strings.Builder

	if qc.Scheme != "" {
		sb.WriteString(qc.Scheme)
		sb.WriteString("://")
	}

	if qc.User != "" {
		sb.WriteString(qc.User)
		if qc.Password != "" {
			sb.WriteString(":")
			sb.WriteString(qc.Password)
		}
		sb.WriteString("/")
	}

	if qc.Host != "" {
		sb.WriteString(qc.Host)
		if qc.Port != 0 {
			sb.WriteByte(':')
			sb.WriteString(strconv.Itoa(int(qc.Port)))
		}
	}

	if qc.LogFile != "" {
		sb.WriteString(qc.LogFile)
	}

	return sb.String()
}

type QueryConns []*QueryConn

func (qcs QueryConns) String() string {
	var sb strings.Builder
	for i, queryConn := range qcs {
		sb.WriteString(queryConn.String())
		if i != len(qcs)-1 {
			sb.WriteByte(',')
		}
	}
	return sb.String()
}

func (qcs QueryConns) Pretty() string {
	var sb strings.Builder
	for i, queryConn := range qcs {
		sb.WriteString(queryConn.String())
		if i != len(qcs)-1 {
			sb.WriteByte(',')
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}

type Query struct {
	Origin      string
	Conns       QueryConns
	Pattern     string
	TimeRange   *times.TimeRange
	SelectQuery string
}

func (q *Query) Save(new Query) {
	q.Origin = new.Origin
	q.Conns = new.Conns
	q.Pattern = new.Pattern
	q.TimeRange = new.TimeRange
	q.SelectQuery = new.SelectQuery
}
