package goconcurrentqueue

import (
	"testing"
)

// ***************************************************************************************
// ** Enqueue
// ***************************************************************************************

// single goroutine - enqueue 1 element
func BenchmarkFIFOEnqueueSingleGR(b *testing.B) {
	fifo := NewFIFO()
	for i := 0; i < b.N; i++ {
		fifo.Enqueue(i)
	}
}

// single goroutine - enqueue 100 elements
func BenchmarkFIFOEnqueue100SingleGR(b *testing.B) {
	fifo := NewFIFO()

	for i := 0; i < b.N; i++ {
		for c := 0; c < 100; c++ {
			fifo.Enqueue(c)
		}
	}
}

// multiple goroutines - enqueue 100 elements per gr
func BenchmarkFIFOEnqueue100MultipleGRs(b *testing.B) {
	fifo := NewFIFO()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for c := 0; c < 100; c++ {
				fifo.Enqueue(c)
			}
		}
	})
}

// multiple goroutines - enqueue 1000 elements per gr
func BenchmarkFIFOEnqueue1000MultipleGRs(b *testing.B) {
	fifo := NewFIFO()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for c := 0; c < 1000; c++ {
				fifo.Enqueue(c)
			}
		}
	})
}

// single goroutine - enqueue 1000 elements
func BenchmarkFIFOEnqueue1000SingleGR(b *testing.B) {
	fifo := NewFIFO()

	for i := 0; i < b.N; i++ {
		for c := 0; c < 1000; c++ {
			fifo.Enqueue(c)
		}
	}
}

// ***************************************************************************************
// ** Dequeue
// ***************************************************************************************

// single goroutine - dequeue 100 elements
func BenchmarkFIFODequeue100SingleGR(b *testing.B) {
	// do not measure the queue initialization
	b.StopTimer()
	fifo := NewFIFO()

	for i := 0; i < b.N; i++ {
		// do not measure the enqueueing process
		b.StopTimer()
		for i := 0; i < 100; i++ {
			fifo.Enqueue(i)
		}

		// measure the dequeueing process
		b.StartTimer()
		for i := 0; i < 100; i++ {
			fifo.Dequeue()
		}
	}
}

// single goroutine 1000 - dequeue 1000 elements
func BenchmarkFIFODequeue1000SingleGR(b *testing.B) {
	// do not measure the queue initialization
	b.StopTimer()
	fifo := NewFIFO()

	for i := 0; i < b.N; i++ {
		// do not measure the enqueueing process
		b.StopTimer()
		for i := 0; i < 1000; i++ {
			fifo.Enqueue(i)
		}

		// measure the dequeueing process
		b.StartTimer()
		for i := 0; i < 1000; i++ {
			fifo.Dequeue()
		}
	}
}

// multiple goroutines - dequeue 100 elements per gr
func BenchmarkFIFODequeue100MultipleGRs(b *testing.B) {
	b.StopTimer()
	fifo := NewFIFO()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			b.StopTimer()
			for c := 0; c < 100; c++ {
				fifo.Enqueue(c)
			}

			b.StartTimer()
			for c := 0; c < 100; c++ {
				fifo.Dequeue()
			}
		}
	})
}

// multiple goroutines - dequeue 1000 elements per gr
func BenchmarkFIFODequeue1000MultipleGRs(b *testing.B) {
	b.StopTimer()
	fifo := NewFIFO()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			b.StopTimer()
			for c := 0; c < 1000; c++ {
				fifo.Enqueue(c)
			}

			b.StartTimer()
			for c := 0; c < 1000; c++ {
				fifo.Dequeue()
			}
		}
	})
}
