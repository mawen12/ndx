package model

import "time"

type QueryContext struct {
	Pattern  string
	From     time.Time
	To       time.Time
	LineUtil int
}
