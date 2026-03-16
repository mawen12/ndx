package model

import "sync"

const (
	StackPush StackEvent = iota
	StackPop
)

type StackEvent int

type StackListener interface {
	Pushed(Page)
	Popped(old, new Page)
	Top(Page)
}

type Stack struct {
	pages     []Page
	listeners []StackListener
	mx        sync.RWMutex
}

func NewStack() *Stack {
	return new(Stack)
}

func (s *Stack) Empty() bool {
	s.mx.RLock()
	defer s.mx.RUnlock()

	return len(s.pages) == 0
}

func (s *Stack) Top() Page {
	s.mx.RLock()
	defer s.mx.RUnlock()

	len := len(s.pages)
	if len != 0 {
		return s.pages[len-1]
	}
	return nil
}

func (s *Stack) Push(c Page) {
	if top := s.Top(); top != nil {
		top.Stop()
	}

	s.mx.Lock()
	s.pages = append(s.pages, c)
	s.mx.Unlock()

	s.notify(StackPush, c)
}

func (s *Stack) Pop() (Page, bool) {
	if s.Empty() {
		return nil, false
	}

	s.mx.Lock()
	old := s.pages[len(s.pages)-1]
	s.pages = s.pages[:len(s.pages)-1]
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

func (s *Stack) notify(event StackEvent, c Page) {
	top := s.Top()

	switch event {
	case StackPop:
		for _, l := range s.listeners {
			l.Popped(c, top)
		}
	case StackPush:
		for _, l := range s.listeners {
			l.Pushed(c)
		}
	}
}
