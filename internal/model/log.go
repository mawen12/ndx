package model

import "time"

type LogLine interface {
	Time() time.Time
	LogFile() string
	LogNumber() int
	Message() string
	Contexts() map[string]string
	Level() string
	OriginalLine() string
}

type QueryResult interface {
	Lines() []LogLine
	Statistics() map[int64]int
	Err() error
	Duration() time.Duration
}

type ResultStream interface {
	Read() QueryResult
	NextLine() error
}
