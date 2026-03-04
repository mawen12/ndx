package conn

import (
	"context"
	"errors"
	"time"

	"github.com/mawen12/ndx/internal/proto"
)

type ResultReader struct {
	ndConn *NdxConn
	ctx    context.Context

	lines  []LineInfo
	stat   map[int64]int
	closed bool
	err    error
}

type LineInfo struct {
	Time          time.Time
	LogFilename   string
	LogLinenumber int
	Msg           string
	Contexts      map[string]string
	Level         string

	OriginLine string
}

// 2026-03-02 09:49:35,189 INFO [main] [] com.github.mawen12.PrometheusSpringBoot3App
func (l *LineInfo) Parse(loc *time.Location) {
	timestamp, err := time.ParseInLocation("2006-01-02 15:04:05.000", l.OriginLine[:23], loc)
	if err == nil {
		l.Time = timestamp
	}

	level := l.OriginLine[24:29]
	if level[len(level)-1] == ' ' {
		l.Level = level[:len(level)-1]
	}
}

type Result struct {
	Lines []LineInfo
	Stat  map[int64]int
	Err   error
}

func (rr *ResultReader) Read() *Result {
	br := &Result{}

	rr.stat = make(map[int64]int)

	err := rr.NextLine()
	if rr.err == nil && err != nil {
		rr.err = err
	}

	br.Lines = rr.lines
	br.Stat = rr.stat

	br.Err = rr.err

	return br
}

func (rr *ResultReader) NextLine() error {
	for !rr.closed && rr.err == nil {
		msg, err := rr.ndConn.receiveMessage()
		if err != nil {
			return err
		}

		switch msg := msg.(type) {
		case *proto.DataLine:
			lineInfo := LineInfo{
				LogLinenumber: msg.CurNR,
				LogFilename:   rr.ndConn.config.LogFile,
				Msg:           msg.Line,
				OriginLine:    msg.Line,
			}
			lineInfo.Parse(rr.ndConn.location)
			rr.lines = append(rr.lines, lineInfo)
		case *proto.Stat:
			t, err := time.ParseInLocation("2006-01-02 15:04", msg.MinuteKey, rr.ndConn.location)
			if err != nil {
				return err
			}

			rr.stat[t.Unix()] = msg.LineCount
		case *proto.ReadyForQuery:
			rr.ndConn.unlock()
			rr.closed = true
			return nil
		case *proto.ErrorResponse:
			rr.err = errors.New(msg.Message)
			rr.closed = true
			return nil
		}
	}

	return nil
}
