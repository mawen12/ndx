package pool

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/mawen12/ndx/internal/model"
)

type Pool struct {
	cs map[string]Conn

	cancel    context.CancelFunc
	listeners []StatListener
}

func Connect(connString string) (*Pool, error) {
	parts := strings.Split(connString, ",")
	if len(parts) < 1 {
		return nil, ErrInvalidConnString
	}

	p := Pool{}

	p.cs = make(map[string]Conn)
	ctx := context.Background()
	for _, part := range parts {
		con := NewConn(ctx, part)
		p.cs[part] = *con
	}

	ctx, p.cancel = context.WithCancel(ctx)

	go p.run(ctx)

	return &p, nil
}

func (p *Pool) run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-ticker.C:
			s := p.Stat()
			p.notifyListeners(s)
			slog.Info("checking connections status", "stat", s)
			if s.Closed != 0 {
				for connString, c := range s.ClosedConns {
					slog.Info("reconnecting", "connString", connString)

					if c.err != nil {
						slog.Error("failed to reconnect", "connString", connString, "err", c.err)
					} else {
						slog.Info("reconnected", "connString", connString, "err", c.err)
					}
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

type Stat struct {
	Idle   int
	Busy   int
	Closed int

	ClosedConns map[string]Conn
}

type StatListener interface {
	OnStat(s Stat)
}

func (p *Pool) AddListener(listener StatListener) {
	p.listeners = append(p.listeners, listener)
}

func (p *Pool) RemoveListener(listener StatListener) {
	for i, l := range p.listeners {
		if l == listener {
			p.listeners = append(p.listeners[:i], p.listeners[i+1:]...)
			return
		}
	}
}

func (p *Pool) notifyListeners(s Stat) {
	for _, listener := range p.listeners {
		listener.OnStat(s)
	}
}

func (s Stat) String() string {
	return fmt.Sprintf("Idle: %d, Busy: %d, Closed: %d", s.Idle, s.Busy, s.Closed)
}

func (p *Pool) Stat() (s Stat) {
	s.ClosedConns = make(map[string]Conn)
	for _, conn := range p.cs {
		if conn.IsClosed() {
			s.Closed++
			s.ClosedConns[conn.connString] = conn
		} else if conn.IsBusy() {
			s.Busy++
		} else {
			s.Idle++
		}
	}
	return
}

func (p *Pool) Close() {
	p.cancel()
	for _, c := range p.cs {
		_ = c.Close()
	}
}

type execResult struct {
	connString string
	Result     model.QueryResult
}

func (p *Pool) Query(ctx context.Context, qc model.QueryContext) (MergedResult, error) {
	start := time.Now()
	resultCh := make(chan execResult)
	var wg sync.WaitGroup
	for connString, conn := range p.cs {
		wg.Add(1)
		go func() {
			defer wg.Done()

			ret := conn.Exec(ctx, qc.Pattern, qc.TimeRange.ActualFrom, qc.TimeRange.ActualQuery)
			resultCh <- execResult{connString: connString, Result: ret}
		}()
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	var errs []error
	mergedStat := make(map[int64]int)
	var mergedLines []model.LogLine
	for ch := range resultCh {
		if ch.Result.Err() != nil {
			errs = append(errs, ch.Result.Err())
			continue
		}

		// handle stat
		for m, c := range ch.Result.Statistics() {
			mergedStat[m] += c
		}

		// handle line
		for _, lineInfo := range ch.Result.Lines() {
			mergedLines = append(mergedLines, lineInfo)
		}
	}

	sort.SliceStable(mergedLines, func(i, j int) bool {
		return mergedLines[i].Time().Before(mergedLines[j].Time())
	})

	return MergedResult{
		Stat:     mergedStat,
		Lines:    mergedLines,
		Duration: time.Since(start),
	}, nil
}
