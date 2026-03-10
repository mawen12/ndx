package model

import (
	"container/list"

	"github.com/rivo/tview"
)

type CyclePrimitive struct {
	tview.Primitive
	Name string
}

func NewCyclePrimitive(p tview.Primitive, name string) CyclePrimitive {
	return CyclePrimitive{
		Primitive: p,
		Name:      name,
	}
}

type CycleList struct {
	*list.List

	root *list.Element
}

func NewCycleList() *CycleList {
	return &CycleList{
		List: list.New(),
	}
}

func (n *CycleList) Reset() {
	n.root = n.List.Front()
}

func (n *CycleList) Current() any {
	if n.root == nil {
		panic("please reset it before use")
	}

	return n.root.Value
}

func (n *CycleList) Next() any {
	if n.root == nil {
		panic("please reset it before use")
	}

	if n.root == n.List.Back() {
		n.root = n.List.Front()
	} else {
		n.root = n.root.Next()
	}

	return n.root.Value
}

func (n *CycleList) Prev() any {
	if n.root == nil {
		panic("please reset it before use")
	}

	if n.root == n.List.Front() {
		n.root = n.List.Back()
	} else {
		n.root = n.root.Prev()
	}

	return n.root.Value
}
