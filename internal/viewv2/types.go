package viewv2

import (
	"sync"

	"github.com/rivo/tview"
)

type Component interface {
	tview.Primitive
	Name() string

	Start()
	Stop()
	Modal() bool
}

const (
	StackPush StackEvent = iota
	StackPop
)

type StackEvent int

type StackListener interface {
	Pushed(Component)
	Poped(old, new Component)
	Top(Component)
}

type Stack struct {
	components []Component
	listeners  []StackListener
	mx         sync.RWMutex
}

func NewStack() *Stack {
	return new(Stack)
}

func (s *Stack) Empty() bool {
	s.mx.RLock()
	defer s.mx.RUnlock()

	return len(s.components) == 0
}

func (s *Stack) Top() Component {
	s.mx.RLock()
	defer s.mx.RUnlock()

	len := len(s.components)
	if len != 0 {
		return s.components[len-1]
	}
	return nil
}

func (s *Stack) Push(c Component) {
	if top := s.Top(); top != nil {
		top.Stop()
	}

	s.mx.Lock()
	s.components = append(s.components, c)
	s.mx.Unlock()

	s.notify(StackPush, c)
}

func (s *Stack) Pop() (Component, bool) {
	if s.Empty() {
		return nil, false
	}

	s.mx.Lock()
	old := s.components[len(s.components)-1]
	s.components = s.components[:len(s.components)-1]
	s.mx.Unlock()

	s.notify(StackPop, old)
	return old, true
}

func (s *Stack) AddListener(l StackListener) {
	s.listeners = append(s.listeners, l)
	if !s.Empty() {
		l.Top(s.Top())
	}
}

func (s *Stack) RemoveListener(l StackListener) {
	for i, listener := range s.listeners {
		if listener == l {
			s.listeners = append(s.listeners[:i], s.listeners[i+1:]...)
			return
		}
	}
}

func (s *Stack) notify(event StackEvent, c Component) {
	top := s.Top()

	switch event {
	case StackPop:
		for _, l := range s.listeners {
			l.Poped(c, top)
		}
	case StackPush:
		for _, l := range s.listeners {
			l.Pushed(c)
		}
	}
}
