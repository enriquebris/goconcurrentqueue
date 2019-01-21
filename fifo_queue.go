package goconcurrentqueue

import (
	"fmt"
	"sync"
)

// FIFO (First In First Out) concurrent queue
type FIFO struct {
	slice   []interface{}
	rwmutex sync.RWMutex
}

// NewFIFO returns a new FIFO concurrent queue
func NewFIFO() *FIFO {
	ret := &FIFO{}
	ret.initialize()

	return ret
}

func (st *FIFO) initialize() {
	st.slice = make([]interface{}, 0)
}

// Enqueue enqueues an element
func (st *FIFO) Enqueue(value interface{}) {
	st.rwmutex.Lock()
	defer st.rwmutex.Unlock()

	st.slice = append(st.slice, value)
}

// Dequeue dequeues an element
func (st *FIFO) Dequeue() (interface{}, error) {
	st.rwmutex.Lock()
	defer st.rwmutex.Unlock()

	len := len(st.slice)
	if len == 0 {
		return nil, fmt.Errorf("queue is empty")
	}

	elementToReturn := st.slice[0]
	st.slice = st.slice[1:]

	return elementToReturn, nil
}

// Get returns an element's value and keeps the element at the queue
func (st *FIFO) Get(index int) (interface{}, error) {
	st.rwmutex.RLock()
	defer st.rwmutex.RUnlock()

	if len(st.slice) <= index {
		return nil, fmt.Errorf("index out of bounds: %v", index)
	}

	return st.slice[index], nil
}

// Remove removes an element from the queue
func (st *FIFO) Remove(index int) error {
	st.rwmutex.Lock()
	defer st.rwmutex.Unlock()

	if len(st.slice) <= index {
		return fmt.Errorf("index out of bounds: %v", index)
	}

	// remove the element
	st.slice = append(st.slice[:index], st.slice[index+1:]...)

	return nil
}

// GetLen returns the number of enqueued elements
func (st *FIFO) GetLen() int {
	st.rwmutex.RLock()
	defer st.rwmutex.RUnlock()

	return len(st.slice)
}
