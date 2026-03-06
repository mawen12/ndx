package pool

import (
	"time"

	"github.com/mawen12/ndx/internal/model"
)

type MergedResult struct {
	Stat     map[int64]int
	Lines    []model.LogLine
	Duration time.Duration
}
