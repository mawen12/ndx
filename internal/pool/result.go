package pool

import "github.com/mawen12/ndx/internal/conn"

type MergedResult struct {
	Stat  map[int64]int
	Lines []conn.LineInfo
}
