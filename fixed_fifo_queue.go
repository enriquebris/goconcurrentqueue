package goconcurrentqueue

import (
	"context"
)

// Fixed capacity FIFO (First In First Out) concurrent queue
type FixedFIFO struct {
	queue    chan interface{}
	lockChan chan struct{}
	// queue for watchers that will wait for next elements (if queue is empty at DequeueOrWaitForNextElement execution )
	waitForNextElementChan chan chan interface{}
}

func NewFixedFIFO(capacity int) *FixedFIFO {
	queue := &FixedFIFO{}
	queue.initialize(capacity)

	return queue
}

func (st *FixedFIFO) initialize(capacity int) {
	st.queue = make(chan interface{}, capacity)
	st.lockChan = make(chan struct{}, 1)
	st.waitForNextElementChan = make(chan chan interface{}, WaitForNextElementChanCapacity)
}

// Enqueue enqueues an element. Returns error if queue is locked or it is at full capacity.
func (st *FixedFIFO) Enqueue(value interface{}) error {
	if st.IsLocked() {
		return NewQueueError(QueueErrorCodeLockedQueue, "The queue is locked")
	}

	// check if there is a listener waiting for the next element (this element)
	select {
	case listener := <-st.waitForNextElementChan:
		// verify whether it is possible to notify the listener (it could be the listener is no longer
		// available because the context expired: DequeueOrWaitForNextElementContext)
		select {
		// sends the element through the listener's channel instead of enqueueing it
		case listener <- value:
		default:
			// push the element into the queue instead of sending it through the listener's channel (which is not available at this moment)
			return st.enqueueIntoQueue(value)
		}

	default:
		// enqueue the element into the queue
		return st.enqueueIntoQueue(value)
	}

	return nil
}

// enqueueIntoQueue enqueues the given item directly into the regular queue
func (st *FixedFIFO) enqueueIntoQueue(value interface{}) error {
	select {
	case st.queue <- value:
	default:
		return NewQueueError(QueueErrorCodeFullCapacity, "FixedFIFO queue is at full capacity")
	}

	return nil
}

// Dequeue dequeues an element. Returns error if: queue is locked, queue is empty or internal channel is closed.
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

// DequeueOrWaitForNextElement dequeues an element (if exist) or waits until the next element gets enqueued and returns it.
// Multiple calls to DequeueOrWaitForNextElement() would enqueue multiple "listeners" for future enqueued elements.
func (st *FixedFIFO) DequeueOrWaitForNextElement() (interface{}, error) {
	return st.DequeueOrWaitForNextElementContext(context.Background())
}

// DequeueOrWaitForNextElementContext dequeues an element (if exist) or waits until the next element gets enqueued and returns it.
// Multiple calls to DequeueOrWaitForNextElementContext() would enqueue multiple "listeners" for future enqueued elements.
// When the passed context expires this function exits and returns the context' error
func (st *FixedFIFO) DequeueOrWaitForNextElementContext(ctx context.Context) (interface{}, error) {
	if st.IsLocked() {
		return nil, NewQueueError(QueueErrorCodeLockedQueue, "The queue is locked")
	}

	select {
	case value, ok := <-st.queue:
		if ok {
			return value, nil
		}
		return nil, NewQueueError(QueueErrorCodeInternalChannelClosed, "internal channel is closed")
	case <-ctx.Done():
		return nil, ctx.Err()
	// queue is empty, add a listener to wait until next enqueued element is ready
	default:
		// channel to wait for next enqueued element
		waitChan := make(chan interface{})

		select {
		// enqueue a watcher into the watchForNextElementChannel to wait for the next element
		case st.waitForNextElementChan <- waitChan:
			// return the next enqueued element, if any
			select {
			case item := <-waitChan:
				return item, nil
			case <-ctx.Done():
				return nil, ctx.Err()
			// try again to get the element from the regular queue (in case waitChan doesn't provide any item)
			case value, ok := <-st.queue:
				if ok {
					return value, nil
				}
				return nil, NewQueueError(QueueErrorCodeInternalChannelClosed, "internal channel is closed")
			}
		default:
			// too many watchers (waitForNextElementChanCapacity) enqueued waiting for next elements
			return nil, NewQueueError(QueueErrorCodeEmptyQueue, "empty queue and can't wait for next element")
		}

		//return nil, NewQueueError(QueueErrorCodeEmptyQueue, "empty queue")
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
