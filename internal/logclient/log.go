package logclient

import (
	"context"
	"errors"
	"time"

	"github.com/mawen12/ndx/internal/model"
	"github.com/mawen12/ndx/internal/proto"
)

type LogLine struct {
	timestamp time.Time
	logFile   string
	logNumber int
	message   string
	contexts  map[string]string
	level     string

	originalLine string
}

func (l *LogLine) Time() time.Time {
	return l.timestamp
}

func (l *LogLine) LogFile() string {
	return l.logFile
}

func (l *LogLine) LogNumber() int {
	return l.logNumber
}

func (l *LogLine) Message() string {
	return l.message
}

func (l *LogLine) Contexts() map[string]string {
	return l.contexts
}

func (l *LogLine) Level() string {
	return l.level
}

func (l *LogLine) OriginalLine() string {
	return l.originalLine
}

type QueryResult struct {
	lines      []model.LogLine
	statistics map[int64]int
	err        error
	duration   time.Duration
}

func (qr *QueryResult) Lines() []model.LogLine {
	return qr.lines
}
func (qr *QueryResult) Statistics() map[int64]int {
	return qr.statistics
}
func (qr *QueryResult) Err() error {
	return qr.err
}
func (qr *QueryResult) Duration() time.Duration {
	return qr.duration
}

type ResultStream struct {
	*LogClient
	ctx context.Context

	lines  []model.LogLine
	stat   map[int64]int
	closed bool
	err    error
}

func (rs *ResultStream) Read() model.QueryResult {
	start := time.Now()
	br := &QueryResult{}

	rs.stat = make(map[int64]int)

	err := rs.NextLine()
	if rs.err == nil && err != nil {
		rs.err = err
	}

	br.lines = rs.lines
	br.statistics = rs.stat

	br.err = rs.err
	br.duration = time.Since(start)

	return br
}

func (rs *ResultStream) NextLine() error {
	for !rs.closed && rs.err == nil {
		msg, err := rs.LogClient.receiveMessage()
		if err != nil {
			return err
		}

		switch msg := msg.(type) {
		case *proto.DataLine:
			lineInfo := &LogLine{
				logNumber:    msg.CurNR,
				logFile:      rs.LogClient.config.LogFile,
				message:      msg.Line,
				originalLine: msg.Line,
			}
			//lineInfo.Parse(rs.ndConn.location)
			rs.lines = append(rs.lines, lineInfo)
		case *proto.Stat:
			t, err := time.ParseInLocation("2006-01-02 15:04", msg.MinuteKey, rs.LogClient.location)
			if err != nil {
				return err
			}

			rs.stat[t.Unix()] = msg.LineCount
		case *proto.ReadyForQuery:
			rs.LogClient.unlock()
			rs.closed = true
			return nil
		case *proto.ErrorResponse:
			rs.err = errors.New(msg.Message)
			rs.closed = true
			return nil
		}
	}

	return nil
}
