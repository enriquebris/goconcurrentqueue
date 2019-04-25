package goconcurrentqueue

import (
	"testing"
)

// ***************************************************************************************
// ** Enqueue
// ***************************************************************************************

// single goroutine - enqueue 1 element
func BenchmarkFixedFIFOEnqueueSingleGR(b *testing.B) {
	fifo := NewFixedFIFO(5)
	for i := 0; i < b.N; i++ {
		fifo.Enqueue(i)
	}
}

// single goroutine - enqueue 100 elements
func BenchmarkFixedFIFOEnqueue100SingleGR(b *testing.B) {
	fifo := NewFixedFIFO(100)

	for i := 0; i < b.N; i++ {
		for c := 0; c < 100; c++ {
			fifo.Enqueue(c)
		}
	}
}

// single goroutine - enqueue 1000 elements
func BenchmarkFixedFIFOEnqueue1000SingleGR(b *testing.B) {
	fifo := NewFixedFIFO(1000)

	for i := 0; i < b.N; i++ {
		for c := 0; c < 1000; c++ {
			fifo.Enqueue(c)
		}
	}
}

// multiple goroutines - enqueue 100 elements per gr
func BenchmarkFixedFIFOEnqueue100MultipleGRs(b *testing.B) {
	b.StopTimer()
	fifo := NewFixedFIFO(5000000)

	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for c := 0; c < 100; c++ {
				fifo.Enqueue(c)
			}
		}
	})
}

// multiple goroutines - enqueue 1000 elements per gr
func BenchmarkFixedFIFOEnqueue1000MultipleGRs(b *testing.B) {
	b.StopTimer()
	fifo := NewFixedFIFO(5000000)

	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for c := 0; c < 1000; c++ {
				fifo.Enqueue(c)
			}
		}
	})
}

// ***************************************************************************************
// ** Dequeue
// ***************************************************************************************

// _benchmarkFixedFIFODequeueNDingleGR is a helper for single GR dequeue
func _benchmarkFixedFIFODequeueNDingleGR(total int, b *testing.B) {
	fifo := NewFixedFIFO(total)

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		for i := 0; i < total; i++ {
			fifo.Enqueue(i)
		}

		b.StartTimer()
		for c := 0; c < total; c++ {
			fifo.Dequeue()
		}
	}
}

// single goroutine - dequeue 100 elements
func BenchmarkFixedFIFODequeue100SingleGR(b *testing.B) {
	_benchmarkFixedFIFODequeueNDingleGR(100, b)
}

// single goroutine - dequeue 1000 elements
func BenchmarkFixedFIFODequeue1000SingleGR(b *testing.B) {
	_benchmarkFixedFIFODequeueNDingleGR(1000, b)
}

// multiple goroutines - dequeue 100 elements per gr
func BenchmarkFixedFIFODequeue100MultipleGRs(b *testing.B) {
	b.StopTimer()
	fifo := NewFixedFIFO(5000000)

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
func BenchmarkFixedFIFODequeue1000MultipleGRs(b *testing.B) {
	b.StopTimer()
	fifo := NewFixedFIFO(5000000)

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
