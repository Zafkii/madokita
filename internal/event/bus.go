package event

import "sync/atomic"

type Handler func(payload any)

type entry struct {
	id      uint64
	handler Handler
}

type Bus struct {
	listeners map[string][]entry
	counter   uint64
}

func NewBus() *Bus {
	return &Bus{
		listeners: make(map[string][]entry),
	}
}

func (b *Bus) On(event string, handler Handler) func() {
	id := atomic.AddUint64(&b.counter, 1)
	b.listeners[event] = append(b.listeners[event], entry{id: id, handler: handler})
	return func() {
		b.Off(event, id)
	}
}

func (b *Bus) Off(event string, id uint64) {
	if entries, ok := b.listeners[event]; ok {
		for i, e := range entries {
			if e.id == id {
				b.listeners[event] = append(entries[:i], entries[i+1:]...)
				return
			}
		}
	}
}

func (b *Bus) Emit(event string, payload any) {
	if entries, ok := b.listeners[event]; ok {
		for _, e := range entries {
			e.handler(payload)
		}
	}
}

func (b *Bus) Clear() {
	b.listeners = make(map[string][]entry)
}
