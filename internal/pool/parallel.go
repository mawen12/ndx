package pool

import (
	"context"
	"sync"
)

type Param interface{}

type Result interface {
	Out() interface{}
	Err() error

	Param() Param
}

type result struct {
	out   interface{}
	err   error
	param Param
}

func (r *result) Out() interface{} {
	return r.out
}

func (r *result) Err() error {
	return r.err
}

func (r *result) Param() Param {
	return r.param
}

func ParallelRun[T Param](ctx context.Context, ps []T, action func(ctx context.Context, in T) (out interface{}, err error)) []Result {
	var wg sync.WaitGroup
	ch := make(chan result, len(ps))

	for _, param := range ps {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ret, err := action(ctx, param)
			ch <- result{out: ret, err: err, param: param}
		}()
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	results := make([]Result, 0, len(ps))
	for {
		select {
		case <-ctx.Done():
			return results
		case ret := <-ch:
			results = append(results, &ret)
			if len(results) == len(ps) {
				return results
			}
		}
	}
}
