package model

import (
	"context"

	"github.com/mawen12/ndx/internal"
	"github.com/rivo/tview"
)

type Igniter interface {
	Init(ctx context.Context)
}

type Primitive[Name ~string] interface {
	tview.Primitive
	Igniter
	Name() Name
}

type Component interface {
	Primitive[internal.ComponentKey]
}

type Page interface {
	Primitive[internal.PageKey]

	Start()
	Stop()
	IsModal() bool
}
