package model

import (
	"context"

	"github.com/rivo/tview"
)

type Primitive interface {
	tview.Primitive

	Name() string
}

type Igniter interface {
	Init(ctx context.Context) error

	Start()

	Stop()
}

type Component interface {
	Primitive
	Igniter
}
