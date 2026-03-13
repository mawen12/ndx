package ui

import (
	"errors"
	"time"

	"github.com/mawen12/ndx/internal/config"
	"github.com/mawen12/ndx/internal/model"
	"github.com/mawen12/ndx/pkg/times"
)

type ViewModel struct {
	*config.Query

	QueryFunc func() error

	RefreshFunc func() error
}

func (vm *ViewModel) DoQuery() error {
	return vm.QueryFunc()
}

func (vm *ViewModel) Refresh() error {
	return vm.RefreshFunc()
}

func (vm *ViewModel) Save(qv model.QueryView) (bool, error) {
	if qv.Conns == "" {
		return false, errors.New("'conns' can not be empty")
	}

	if qv.TimeRange == "" {
		return false, errors.New("'timeRange' can not be empty")
	}

	query := config.Query{
		Origin:      qv.Conns,
		Conns:       vm.Conns,
		TimeRange:   vm.TimeRange,
		Pattern:     qv.Pattern,
		SelectQuery: qv.SelectQuery,
	}

	reconnect := false
	if qv.Conns != vm.Conns.String() {
		conns, err := config.ParseConns(qv.Conns)
		if err != nil {
			return false, err
		}

		reconnect = true
		query.Conns = conns
	}

	if qv.TimeRange != vm.TimeRange.Spec() {
		tr, err := times.ParseFromTimeStr(time.UTC, qv.TimeRange)
		if err != nil {
			return false, err
		}

		query.TimeRange = tr
	}

	vm.Query.Save(query)

	return reconnect, nil
}
