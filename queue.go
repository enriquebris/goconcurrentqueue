package goconcurrentqueue

// Queue interface with basic && common queue functions
type Queue interface {
	// Enqueue element
	Enqueue(interface{}) error
	// Dequeue element
	Dequeue() (interface{}, error)
	// Get number of enqueued elements
	GetLen() int
	// Get queue's capacity
	GetCap() int

	// Lock the queue. No enqueue/dequeue/remove/get operations will be allowed after this point.
	Lock()
	// Unlock the queue.
	Unlock()
	// Return true whether the queue is locked
	IsLocked() bool
}
