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
	for _, queryConn := range qcs {
		sb.WriteString(queryConn.String())
	}
	return sb.String()
}

type Query struct {
	Conns       QueryConns
	Pattern     string
	TimeRange   *times.TimeRange
	SelectQuery string
}
