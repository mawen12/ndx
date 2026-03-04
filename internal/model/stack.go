package model

import "sync"

const (
	StackPush StackAction = 1 << iota

	StackPop
)

type StackAction int

type StackListener interface {
	StackPushed(Component)

	StackPopped(old, new Component)

	StackTop(Component)
}

type Stack struct {
	components []Component
	listeners  []StackListener
	mx         sync.RWMutex
}

func NewStack() *Stack {
	return new(Stack)
}

func (s *Stack) AddListener(l StackListener) {
	s.listeners = append(s.listeners, l)
	if !s.Empty() {
		l.StackTop(s.Top())
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

func (s *Stack) Flatten() []string {
	s.mx.RLock()
	defer s.mx.RUnlock()

	ss := make([]string, len(s.components))
	for i, c := range s.components {
		ss[i] = c.Name()
	}
	return ss
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

// Pop 弹出一个元素，bool 代表此次操作是否成功
func (s *Stack) Pop() (Component, bool) {
	if s.Empty() {
		return nil, false
	}

	s.mx.Lock()
	c := s.components[len(s.components)-1]
	c.Stop()
	s.components = s.components[:len(s.components)-1]
	s.mx.Unlock()

	s.notify(StackPop, c)
	return c, true
}

func (s *Stack) Peek() []Component {
	s.mx.RLock()
	defer s.mx.RUnlock()
	return s.components
}

func (s *Stack) Empty() bool {
	s.mx.RLock()
	defer s.mx.RUnlock()
	return len(s.components) == 0
}

func (s *Stack) IsLast() bool {
	s.mx.RLock()
	defer s.mx.RUnlock()
	return len(s.components) == 1
}

func (s *Stack) Previous() Component {
	s.mx.RLock()
	defer s.mx.RUnlock()

	if s.Empty() {
		return nil
	}
	if s.IsLast() {
		return s.Top()
	}
	return s.components[len(s.components)-2]
}

func (s *Stack) Top() Component {
	s.mx.RLock()
	defer s.mx.RUnlock()
	if s.Empty() {
		return nil
	}
	return s.components[len(s.components)-1]
}

func (s *Stack) notify(a StackAction, c Component) {
	top := s.Top()

	switch a {
	case StackPop:
		for _, l := range s.listeners {
			l.StackPopped(c, top)
		}
	case StackPush:
		for _, l := range s.listeners {
			l.StackPushed(c)
		}
	}
}
