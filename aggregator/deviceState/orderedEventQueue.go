// a run queue that guarantees ordered execution
package deviceState

import (
	"container/heap"
	"errors"

	"go.uber.org/zap"
)

type orderedEvent struct {
	handler func()
	eid     int64
}

type eventHeap []*orderedEvent

type OrderedEventQueue struct {
	lastid int64
	events eventHeap
	log    *zap.Logger
}

func (eh eventHeap) Len() int { return len(eh) }

func (eh eventHeap) Less(i, j int) bool {
	return eh[i].eid < eh[j].eid
}

func (eh eventHeap) Swap(i, j int) {
	eh[i], eh[j] = eh[j], eh[i]
}

func (eh *eventHeap) Push(x interface{}) {
	item := x.(*orderedEvent)
	*eh = append(*eh, item)
}

func (eh *eventHeap) Pop() interface{} {
	old := *eh
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // avoid memory leak
	*eh = old[0 : n-1]
	return item
}

func (eq *OrderedEventQueue) Push(eid int64, handler func()) error {
	switch {
	case eid <= eq.lastid:
		return errors.New("OrderedEventQueue: eid is in the past")
	case eq.lastid+1 == eid:
		handler()
		eq.lastid += 1
		if eq.log != nil {
			eq.log.Debug("OrderedEventQueue: running event", zap.Int64("eid", eid))
		}
		for len(eq.events) > 0 && eq.lastid+1 == eq.events[0].eid {
			heap.Pop(&eq.events).(*orderedEvent).handler()
			if eq.log != nil {
				eq.log.Debug("OrderedEventQueue: running event", zap.Int64("eid", eq.lastid))
			}
			eq.lastid += 1
		}
	default:
		heap.Push(&eq.events, &orderedEvent{eid: eid, handler: handler})
	}
	return nil
}
