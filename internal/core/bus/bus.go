package bus

import "sync"

type EventBus struct {
	mu   sync.RWMutex
	subs map[EventType][]chan Event
}

func NewEventBus() *EventBus {
	return &EventBus{
		subs: make(map[EventType][]chan Event),
	}
}

func (b *EventBus) Subscribe(eventType EventType) <-chan Event {
	ch := make(chan Event, 100)

	b.mu.Lock()
	b.subs[eventType] = append(b.subs[eventType], ch)
	b.mu.Unlock()

	return ch
}

func (b *EventBus) Publish(event Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, ch := range b.subs[event.Type] {
		select {
		case ch <- event:
		default:
			// drop if slow consumer (important for packet systems)
		}
	}
}

func (b *EventBus) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, chans := range b.subs {
		for _, ch := range chans {
			close(ch)
		}
	}
}

func (b *EventBus) Unsubscribe(eventType EventType, ch <-chan Event) {
	b.mu.Lock()
	defer b.mu.Unlock()

	chans, ok := b.subs[eventType]
	if !ok {
		return
	}

	for i, v := range chans {
		// Check if this is the channel we want to remove
		if v == ch {
			// Remove the element by shifting the slice
			b.subs[eventType] = append(chans[:i], chans[i+1:]...)

			// Close the channel to free up the subscriber goroutine
			close(v)
			break
		}
	}
}
