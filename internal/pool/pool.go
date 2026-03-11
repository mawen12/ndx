package pool

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"sync"
	"time"

	"github.com/mawen12/ndx/internal/config"
	"github.com/mawen12/ndx/internal/model"
)

type Pool struct {
	cs map[string]*Conn

	connected bool
	cancel    context.CancelFunc
	listeners []StatListener
}

func NewPool(conns config.QueryConns) *Pool {
	p := Pool{}

	p.cs = make(map[string]*Conn)
	for _, part := range conns {
		p.cs[part.Origin] = NewConn(part)
	}

	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel

	go p.run(ctx)

	return &p
}

func (p *Pool) Connect(ctx context.Context, callback func(string, string, bool)) error {
	for _, conn := range p.cs {
		cb := func(message string, finished bool) {
			callback(conn.Conn.Origin, message, finished)
		}

		err := conn.Connect(ctx, cb)
		if err != nil {
			cb(err.Error(), true)
			return err
		}

		cb("", true)
	}

	p.connected = true

	return nil
}

func (p *Pool) IsConnected() bool {
	return p.connected
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
				//for connString, c := range s.ClosedConns {
				//	slog.Info("reconnecting", "connString", connString)

				//if c.err != nil {
				//	slog.Error("failed to reconnect", "connString", connString, "err", c.err)
				//} else {
				//	slog.Info("reconnected", "connString", connString, "err", c.err)
				//}
				//}
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

	ClosedConns map[string]*Conn
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
	s.ClosedConns = make(map[string]*Conn)
	for _, conn := range p.cs {
		if conn.LogClient != nil {
			if conn.IsClosed() {
				s.Closed++
				s.ClosedConns[conn.Conn.Origin] = conn
			} else if conn.IsBusy() {
				s.Busy++
			} else {
				s.Idle++
			}
		}

	}
	return
}

func (p *Pool) Close() {
	if !p.connected {
		return
	}

	p.cancel()
	for _, c := range p.cs {
		if c.LogClient != nil {
			_ = c.Close()
		}
	}
	p.connected = false
}

type execResult struct {
	connString string
	Result     model.QueryResult
}

func (p *Pool) Query(ctx context.Context, qc config.Query) (MergedResult, error) {
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
