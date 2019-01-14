package goconcurrentqueue

type Queue interface {
	// Enqueue element
	Enqueue(interface{})
	// Dequeue element
	Dequeue() (interface{}, error)
	// Get number of enqueued elements
	GetLen() int
	// Get an element's value and keep it at the queue
	Get(int) (interface{}, error)
	// Remove any element from the queue
	Remove(index int) error
}
