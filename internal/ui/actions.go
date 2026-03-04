package ui

import (
	"sync"

	"github.com/gdamore/tcell/v2"
)

type ActionHandler func(key *tcell.EventKey) *tcell.EventKey

type ActionOpts struct {
	Visible bool
	HotKey  bool
}

type KeyAction struct {
	Description string
	Action      ActionHandler
	Opts        ActionOpts
}

type KeyMap map[tcell.Key]KeyAction

type KeyActions struct {
	actions KeyMap
	mx      sync.RWMutex
}

func NewKeyAction(d string, a ActionHandler, visible bool) KeyAction {
	return NewKeyActionWithOpts(d, a, ActionOpts{
		Visible: visible,
	})
}

func NewKeyActionWithOpts(d string, a ActionHandler, opts ActionOpts) KeyAction {
	return KeyAction{
		Description: d,
		Action:      a,
		Opts:        opts,
	}
}

func NewKeyActionsFromMap(mm KeyMap) *KeyActions {
	return &KeyActions{
		actions: mm,
	}
}
