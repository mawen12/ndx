package model

import (
	"context"
	"fmt"
	"sync"
)

type Holder[K comparable, V Igniter] struct {
	mx       sync.RWMutex
	initOnce sync.Once
	holders  map[K]V
}

func NewHolder[K comparable, V Igniter]() *Holder[K, V] {
	return &Holder[K, V]{
		holders: make(map[K]V),
	}
}

func (h *Holder[K, V]) Init(ctx context.Context) {
	h.initOnce.Do(func() {
		for _, value := range h.holders {
			value.Init(ctx)
		}
	})
}

func (h *Holder[K, V]) Add(key K, c V) {
	h.mx.Lock()
	defer h.mx.Unlock()

	if _, ok := h.holders[key]; ok {
		panic(fmt.Sprintf("Cannot register %v, because it is already exists", key))
	}
	h.holders[key] = c
}

func (h *Holder[K, V]) Get(key K) (c V, ok bool) {
	h.mx.RLock()
	defer h.mx.RUnlock()

	c, ok = h.holders[key]
	return
}

func (h *Holder[K, V]) MustGet(key K) V {
	h.mx.RLock()
	defer h.mx.RUnlock()

	c, ok := h.holders[key]
	if !ok {
		panic(fmt.Sprintf("Holder doesn't exist %v please add it.", key))
	}
	return c
}
