package goconcurrentqueue

// Fixed capacity FIFO (First In First Out) concurrent queue
type FixedFIFO struct {
	queue    chan interface{}
	lockChan chan struct{}
}

func NewFixedFIFO(capacity int) *FixedFIFO {
	queue := &FixedFIFO{}
	queue.initialize(capacity)

	return queue
}

func (st *FixedFIFO) initialize(capacity int) {
	st.queue = make(chan interface{}, capacity)
	st.lockChan = make(chan struct{}, 1)
}

func (st *FixedFIFO) Enqueue(value interface{}) error {
	if st.IsLocked() {
		return NewQueueError(QueueErrorCodeLockedQueue, "The queue is locked")
	}

	select {
	case st.queue <- value:
		return nil
	default:
		return NewQueueError(QueueErrorCodeFullCapacity, "FixedFIFO queue is at full capacity")
	}
}

func (st *FixedFIFO) Dequeue() (interface{}, error) {
	if st.IsLocked() {
		return nil, NewQueueError(QueueErrorCodeLockedQueue, "The queue is locked")
	}

	select {
	case value, ok := <-st.queue:
		if ok {
			return value, nil
		}
		return nil, NewQueueError(QueueErrorCodeInternalChannelClosed, "internal channel is closed")
	default:
		return nil, NewQueueError(QueueErrorCodeEmptyQueue, "empty queue")
	}
}

// GetLen returns queue's length (total enqueued elements)
func (st *FixedFIFO) GetLen() int {
	st.Lock()
	defer st.Unlock()

	return len(st.queue)
}

// GetCap returns the queue's capacity
func (st *FixedFIFO) GetCap() int {
	st.Lock()
	defer st.Unlock()

	return cap(st.queue)
}

func (st *FixedFIFO) Lock() {
	// non-blocking fill the channel
	select {
	case st.lockChan <- struct{}{}:
	default:
	}
}

func (st *FixedFIFO) Unlock() {
	// non-blocking flush the channel
	select {
	case <-st.lockChan:
	default:
	}
}

func (st *FixedFIFO) IsLocked() bool {
	return len(st.lockChan) >= 1
}
