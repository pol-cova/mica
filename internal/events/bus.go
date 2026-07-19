package events

import (
	"sync"

	"github.com/mica-dev/mica/internal/incidents"
)

// Bus transports updates to connected local workspaces. The incident timeline
// remains persisted; consumers must reload it after reconnecting.
type Bus struct {
	mu          sync.RWMutex
	next        uint64
	subscribers map[uint64]chan incidents.TimelineEvent
}

func New() *Bus { return &Bus{subscribers: make(map[uint64]chan incidents.TimelineEvent)} }
func (b *Bus) Publish(event incidents.TimelineEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, subscriber := range b.subscribers {
		select {
		case subscriber <- event:
		default:
		}
	}
}
func (b *Bus) Subscribe() (<-chan incidents.TimelineEvent, func()) {
	b.mu.Lock()
	id := b.next
	b.next++
	channel := make(chan incidents.TimelineEvent, 16)
	b.subscribers[id] = channel
	b.mu.Unlock()
	return channel, func() {
		b.mu.Lock()
		if current, ok := b.subscribers[id]; ok {
			delete(b.subscribers, id)
			close(current)
		}
		b.mu.Unlock()
	}
}
