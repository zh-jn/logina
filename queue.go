package logina

import "sync"

var InitQueueSize = 1000
var MaxQueueSize = 10000

type OverflowStrategy int

const (
	OverflowStrategy_DiscardNewest OverflowStrategy = iota
	OverflowStrategy_DiscardOldest
	OverflowStrategy_Unknown
)

type PushResult int

const (
	PushResult_Success PushResult = iota
	PushResult_DiscardNewest
	PushResult_DiscardOldest
	PushResult_Fail
)

type EntryQueue struct {
	q        []*Entry
	front    int
	tail     int
	strategy OverflowStrategy
	mu       sync.Mutex
}

func NewQueue(strategy OverflowStrategy, sizes ...int) *EntryQueue {
	argsN := len(sizes)
	if argsN > 2 {
		argsN = 2
	}
	switch argsN {
	case 2:
		MaxQueueSize = sizes[1]
		fallthrough
	case 1:
		InitQueueSize = sizes[0]
	}

	ret := &EntryQueue{
		q:        make([]*Entry, InitQueueSize, MaxQueueSize),
		front:    0,
		tail:     0,
		strategy: strategy,
	}
	return ret
}

func (queue *EntryQueue) Full() bool {
	l := len(queue.q)
	if (queue.tail+1)%l == queue.front {
		return true
	}
	return false
}

func (queue *EntryQueue) Empty() bool {
	return queue.front == queue.tail
}

func (queue *EntryQueue) Push(e *Entry) PushResult {
	if e == nil {
		return PushResult_Fail
	}
	queue.mu.Lock()
	l := len(queue.q)
	if queue.Full() {
		switch queue.strategy {
		case OverflowStrategy_DiscardNewest:
			queue.mu.Unlock()
			return PushResult_DiscardNewest
		case OverflowStrategy_DiscardOldest:
			queue.q[queue.tail] = e
			queue.front = (queue.front + 1) % l
			queue.tail = (queue.tail + 1) % l
			queue.mu.Unlock()
			return PushResult_DiscardOldest
		default:
			queue.mu.Unlock()
			return PushResult_Fail
		}
	}
	queue.q[queue.tail] = e
	queue.tail = (queue.tail + 1) % l
	queue.mu.Unlock()
	return PushResult_Success
}

func (queue *EntryQueue) Pop() *Entry {
	queue.mu.Lock()
	l := len(queue.q)
	if queue.Empty() {
		queue.mu.Unlock()
		return nil
	}
	ret := queue.q[queue.front]
	queue.front = (queue.front + 1) % l
	queue.mu.Unlock()
	return ret
}
