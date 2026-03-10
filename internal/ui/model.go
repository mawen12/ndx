package ui

import (
	"errors"

	"github.com/mawen12/ndx/internal/config"
	"github.com/mawen12/ndx/internal/model"
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

func (vm *ViewModel) Save(qv model.QueryView) error {
	if qv.Conns == "" {
		return errors.New("'conns' can not be empty")
	}

	if qv.TimeRange == "" {
		return errors.New("'timeRange can not be empty'")
	}

	// reconnect
	if vm.Conns.String() != qv.Conns {
	}

	return nil
}
