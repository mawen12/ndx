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

	IsConnectFunc func() bool

	QueryFunc func() error

	RefreshFunc func() error
}

func (vm *ViewModel) DoQuery() error {
	return vm.QueryFunc()
}

func (vm *ViewModel) Refresh() error {
	return vm.RefreshFunc()
}

func (vm *ViewModel) Save(qv model.QueryView) error {
	if qv.Conns == "" {
		return errors.New("'conns' can not be empty")
	}

	if qv.TimeRange == "" {
		return errors.New("'timeRange' can not be empty")
	}

	query := config.Query{
		Pattern:     qv.Pattern,
		SelectQuery: qv.SelectQuery,
	}

	reconnect := false
	if qv.Conns != vm.Conns.String() {
		conns, err := config.ParseConns(qv.Conns)
		if err != nil {
			return err
		}

		reconnect = true
		query.Conns = conns
	} else {
		query.Conns = vm.Conns
	}

	if qv.TimeRange != vm.TimeRange.Spec() {
		tr, err := times.ParseFromTimeStr(time.UTC, qv.TimeRange)
		if err != nil {
			return err
		}

		query.TimeRange = tr
	} else {
		query.TimeRange = vm.TimeRange
	}

	// connect
	vm.Pattern = query.Pattern
	vm.TimeRange = query.TimeRange
	vm.Conns = query.Conns

	vm.SelectQuery = query.SelectQuery

	if !vm.IsConnectFunc() || reconnect {
		if err := vm.Refresh(); err != nil {
			return err
		}
	}

	// query
	return vm.DoQuery()
}
