[![godoc reference](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/enriquebris/goconcurrentqueue) ![version](https://img.shields.io/badge/version-v0.3.0-yellowgreen.svg?style=flat "goconcurrentqueue v0.3.0")  [![Go Report Card](https://goreportcard.com/badge/github.com/enriquebris/goconcurrentqueue)](https://goreportcard.com/report/github.com/enriquebris/goconcurrentqueue)  [![Build Status](https://api.travis-ci.org/enriquebris/goconcurrentqueue.svg?branch=master)](https://travis-ci.org/enriquebris/goconcurrentqueue) [![codecov](https://codecov.io/gh/enriquebris/goconcurrentqueue/branch/master/graph/badge.svg)](https://codecov.io/gh/enriquebris/goconcurrentqueue)

# goconcurrentqueue - Concurrent queues
Concurrent-safe queue. Multiple goroutines (GR) could access the queue simultaneously without adding a race condition.

## Installation

Execute
```bash
go get github.com/enriquebris/goconcurrentqueue
```

## Documentation
Visit [goconcurrentqueue at godoc.org](https://godoc.org/github.com/enriquebris/goconcurrentqueue)

## Qeueues

- First In First Out (FIFO)
    - [FIFO](#fifo)
    - [FixedFIFO](#fixedfifo)
    - [Benchmarks FixedFIFO vs FIFO](#benchmarks-fixedfifo-vs-fifo)

### FIFO

**FIFO**: concurrent-safe auto expandable queue.

#### pros
Client is able to enqueue as many items as needed.

#### cons
It is slightly slower than FixedFIFO.

### FixedFIFO

**FixedFIFO**: concurrent-safe fixed capacity queue.

#### pros
FixedFIFO is, at least, 2x faster than [FIFO](#fifo) in concurrent scenarios (multiple GR accessing the queue simultaneously).

#### cons
It has a fixed capacity meaning that no more items than this capacity could coexist at the same time. 

## Benchmarks FixedFIFO vs FIFO

### Enqueue

![concurrent-safe FixedFIFO vs FIFO . operation: enqueue](web/FixedFIFO-vs-FIFO-enqueue.png "concurrent-safe FixedFIFO vs FIFO . operation: enqueue")

### Dequeue

![concurrent-safe FixedFIFO vs FIFO . operation: dequeue](web/FixedFIFO-vs-FIFO-dequeue.png "concurrent-safe FixedFIFO vs FIFO . operation: dequeue")

## Get started

### Fifo queue simple usage

```go
package main

import (
	"fmt"

	"github.com/enriquebris/goconcurrentqueue"
)

type AnyStruct struct {
	Field1 string
	Field2 int
}

func main() {
	fifoQueue := goconcurrentqueue.NewFIFO()

	// enqueue two elements (different types)
	fifoQueue.Enqueue(AnyStruct{"one", 1})
	fifoQueue.Enqueue("Paris")

	// dequeue the first element
	item, _ := fifoQueue.Dequeue()

	fmt.Println(item)
}
```

### FIFO queue detailed example

```go
package main

import (
	"fmt"
	"github.com/enriquebris/goconcurrentqueue"
)

func main() {
	// instantiate the FIFO queue
	fifoQueue := goconcurrentqueue.NewFIFO()

	totalElementsToEnqueue := 100

	// print total enqueued elements
	fmt.Printf("Total enqueued elements at queue instantiation: %v\n", fifoQueue.GetLen())

	fmt.Printf("\n(step 1) - Enqueue %v elements\n", totalElementsToEnqueue)
	// enqueue n elements ( n ==> totalElementsToEnqueue )
	for i := 1; i <= totalElementsToEnqueue; i++ {
		fifoQueue.Enqueue(i)
	}
	// print total enqueued elements
	fmt.Printf("Total enqueued elements: %v\n", fifoQueue.GetLen())

	// dequeue a element
	fmt.Println("\n(step 2) - Dequeue 1 element")
	element, err := fifoQueue.Dequeue()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Dequeued element's value: %v\n", element)
	}
	// print total enqueued elements
	fmt.Printf("Total enqueued elements: %v\n", fifoQueue.GetLen())

	// get the value of the first element, the next to be dequeued
	fmt.Println("\n(step 3) - Get element at index 0 (not dequeue)")
	element, err = fifoQueue.Get(0)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Element at first position (0): %v\n", element)
	}
	// print total enqueued elements
	fmt.Printf("Total enqueued elements: %v\n", fifoQueue.GetLen())

	// remove an arbitrary element (based on the index)
	fmt.Println("\n(step 4) - Remove element at index 1")
	err = fifoQueue.Remove(1)
	if err != nil {
		fmt.Printf("Error at queue.Remove(...): '%v'\n", err.Error())
	}

	// print total enqueued elements
	fmt.Printf("Total enqueued elements: %v\n", fifoQueue.GetLen())
}
```

## History

### v0.3.0

- Added FixedFIFO queue's implementation (at least 2x faster than FIFO for multiple GRs)
- Added benchmarks for both FIFO / FixedFIFO
- Added GetCap() to Queue interface
- Removed Get() and Remove() methods from Queue interface

### v0.2.0

- Added Lock/Unlock/IsLocked methods to control operations locking

### v0.1.0

- First In First Out (FIFO) queue added
