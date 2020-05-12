package goconcurrentqueue

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

const (
	fixedFIFOQueueCapacity = 500
)

type FixedFIFOTestSuite struct {
	suite.Suite
	fifo *FixedFIFO
}

func (suite *FixedFIFOTestSuite) SetupTest() {
	suite.fifo = NewFixedFIFO(fixedFIFOQueueCapacity)
}

// ***************************************************************************************
// ** Run suite
// ***************************************************************************************

func TestFixedFIFOTestSuite(t *testing.T) {
	suite.Run(t, new(FixedFIFOTestSuite))
}

// ***************************************************************************************
// ** Enqueue && GetLen
// ***************************************************************************************

// single enqueue lock verification
func (suite *FixedFIFOTestSuite) TestEnqueueLockSingleGR() {
	suite.NoError(suite.fifo.Enqueue(1), "Unlocked queue allows to enqueue elements")

	suite.fifo.Lock()
	err := suite.fifo.Enqueue(1)
	suite.Error(err, "Locked queue does not allow to enqueue elements")

	// verify custom error: code: QueueErrorCodeLockedQueue
	customError, ok := err.(*QueueError)
	suite.True(ok, "Expected error type: QueueError")
	// verify custom error code
	suite.Equalf(QueueErrorCodeLockedQueue, customError.Code(), "Expected code: '%v'", QueueErrorCodeLockedQueue)
}

// single enqueue (1 element, 1 goroutine)
func (suite *FixedFIFOTestSuite) TestEnqueueLenSingleGR() {
	suite.fifo.Enqueue(testValue)
	len := suite.fifo.GetLen()
	suite.Equalf(1, len, "Expected number of elements in queue: 1, currently: %v", len)

	suite.fifo.Enqueue(5)
	len = suite.fifo.GetLen()
	suite.Equalf(2, len, "Expected number of elements in queue: 2, currently: %v", len)
}

// single enqueue at full capacity, 1 goroutine
func (suite *FixedFIFOTestSuite) TestEnqueueFullCapacitySingleGR() {
	total := 5
	suite.fifo = NewFixedFIFO(total)

	for i := 0; i < total; i++ {
		suite.NoError(suite.fifo.Enqueue(i), "no error expected when queue is not full")
	}

	err := suite.fifo.Enqueue(0)
	suite.Error(err, "error expected when queue is full")
	// verify custom error: code: QueueErrorCodeLockedQueue
	customError, ok := err.(*QueueError)
	suite.True(ok, "Expected error type: QueueError")
	// verify custom error code
	suite.Equalf(QueueErrorCodeFullCapacity, customError.Code(), "Expected code: '%v'", QueueErrorCodeFullCapacity)
}

func (suite *FixedFIFOTestSuite) TestEnqueueListenerToExpireSingleGR() {
	var (
		uselessChan = make(chan interface{})
		value = "my-test-value"
	)

	// let Enqueue knows there is a channel to send the next item instead of enqueueing it into the queue
	suite.fifo.waitForNextElementChan <- uselessChan

	// enqueues an item having an waiting channel but without a valid listener, so the item should be enqueued into the queue
	suite.fifo.Enqueue(value)
	// dequeues the item directly from the queue
	dequeuedValue, _ := suite.fifo.Dequeue()

	suite.Equal(value, dequeuedValue)
}

// TestEnqueueLenMultipleGR enqueues elements concurrently
//
// Detailed steps:
//	1 - Enqueue totalGRs concurrently (from totalGRs different GRs)
//	2 - Verifies the len, it should be equal to totalGRs
//	3 - Verifies that all elements from 0 to totalGRs were enqueued
func (suite *FixedFIFOTestSuite) TestEnqueueLenMultipleGR() {
	var (
		totalGRs = 500
		wg       sync.WaitGroup
	)

	// concurrent enqueueing
	// multiple GRs concurrently enqueueing consecutive integers from 0 to (totalGRs - 1)
	for i := 0; i < totalGRs; i++ {
		wg.Add(1)
		go func(value int) {
			defer wg.Done()
			suite.fifo.Enqueue(value)
		}(i)
	}
	wg.Wait()

	// check that there are totalGRs elements enqueued
	totalElements := suite.fifo.GetLen()
	suite.Equalf(totalGRs, totalElements, "Total enqueued elements should be %v, currently: %v", totalGRs, totalElements)

	// checking that the expected elements (1, 2, 3, ... totalGRs-1 ) were enqueued
	var (
		tmpVal                interface{}
		val                   int
		err                   error
		totalElementsVerified int
	)

	// slice to check every element
	allElements := make([]bool, totalGRs)
	for i := 0; i < totalElements; i++ {
		tmpVal, err = suite.fifo.Dequeue()
		suite.NoError(err, "No error should be returned trying to get an existent element")

		val = tmpVal.(int)
		if allElements[val] {
			suite.Failf("Duplicated element", "Unexpected duplicated value: %v", val)
		} else {
			allElements[val] = true
			totalElementsVerified++
		}
	}
	suite.True(totalElementsVerified == totalGRs, "Enqueued elements are missing")
}

// call GetLen concurrently
func (suite *FixedFIFOTestSuite) TestGetLenMultipleGRs() {
	var (
		totalGRs               = 100
		totalElementsToEnqueue = 10
		wg                     sync.WaitGroup
	)

	for i := 0; i < totalElementsToEnqueue; i++ {
		suite.fifo.Enqueue(i)
	}

	for i := 0; i < totalGRs; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			total := suite.fifo.GetLen()
			suite.Equalf(totalElementsToEnqueue, total, "Expected len: %v", totalElementsToEnqueue)
		}()
	}
	wg.Wait()
}

// ***************************************************************************************
// ** GetCap
// ***************************************************************************************

// single GR getCapacity
func (suite *FixedFIFOTestSuite) TestGetCapSingleGR() {
	// initial capacity
	suite.Equal(fixedFIFOQueueCapacity, suite.fifo.GetCap(), "unexpected capacity")

	// new fifo with capacity == 10
	suite.fifo = NewFixedFIFO(10)
	suite.Equal(10, suite.fifo.GetCap(), "unexpected capacity")
}

// ***************************************************************************************
// ** Dequeue
// ***************************************************************************************

// single dequeue lock verification
func (suite *FixedFIFOTestSuite) TestDequeueLockSingleGR() {
	suite.fifo.Enqueue(1)
	_, err := suite.fifo.Dequeue()
	suite.NoError(err, "Unlocked queue allows to dequeue elements")

	suite.fifo.Enqueue(1)
	suite.fifo.Lock()
	_, err = suite.fifo.Dequeue()
	suite.Error(err, "Locked queue does not allow to dequeue elements")

	// verify custom error: code: QueueErrorCodeLockedQueue
	customError, ok := err.(*QueueError)
	suite.True(ok, "Expected error type: QueueError")
	// verify custom error code
	suite.Equalf(QueueErrorCodeLockedQueue, customError.Code(), "Expected code: '%v'", QueueErrorCodeLockedQueue)
}

// dequeue an empty queue
func (suite *FixedFIFOTestSuite) TestDequeueEmptyQueueSingleGR() {
	val, err := suite.fifo.Dequeue()

	// error expected
	suite.Errorf(err, "Can't dequeue an empty queue")

	// verify custom error type
	customError, ok := err.(*QueueError)
	suite.True(ok, "Expected error type: QueueError")
	// verify custom error code
	suite.Equalf(QueueErrorCodeEmptyQueue, customError.Code(), "Expected code: '%v'", QueueErrorCodeEmptyQueue)

	// no value expected
	suite.Equal(nil, val, "Can't get a value different than nil from an empty queue")
}

// dequeue all elements
func (suite *FixedFIFOTestSuite) TestDequeueSingleGR() {
	suite.fifo.Enqueue(testValue)
	suite.fifo.Enqueue(5)

	// dequeue the first element
	val, err := suite.fifo.Dequeue()
	suite.NoError(err, "Unexpected error")
	suite.Equal(testValue, val, "Wrong element's value")
	len := suite.fifo.GetLen()
	suite.Equal(1, len, "Incorrect number of queue elements")

	// get the second element
	val, err = suite.fifo.Dequeue()
	suite.NoError(err, "Unexpected error")
	suite.Equal(5, val, "Wrong element's value")
	len = suite.fifo.GetLen()
	suite.Equal(0, len, "Incorrect number of queue elements")

}

// dequeue an item after closing the empty queue's channel
func (suite *FixedFIFOTestSuite) TestDequeueClosedChannelSingleGR() {
	// enqueue a dummy item
	suite.fifo.Enqueue(1)
	// close the internal queue's channel
	close(suite.fifo.queue)
	// dequeue the dummy item
	suite.fifo.Dequeue()

	// dequeue after the queue's channel was closed
	val, err := suite.fifo.Dequeue()

	suite.Nil(val, "nil value expected after internal channel was closed")
	suite.Error(err, "error expected after internal queue channel was closed")
	// verify custom error: code: QueueErrorCodeLockedQueue
	customError, ok := err.(*QueueError)
	suite.True(ok, "Expected error type: QueueError")
	// verify custom error code
	suite.Equalf(QueueErrorCodeInternalChannelClosed, customError.Code(), "Expected code: '%v'", QueueErrorCodeInternalChannelClosed)
}

// TestDequeueMultipleGRs dequeues elements concurrently
//
// Detailed steps:
//	1 - Enqueues totalElementsToEnqueue consecutive integers
//	2 - Dequeues totalElementsToDequeue concurrently from totalElementsToDequeue GRs
//	3 - Verifies the final len, should be equal to totalElementsToEnqueue - totalElementsToDequeue
//	4 - Verifies that the next dequeued element's value is equal to totalElementsToDequeue
func (suite *FixedFIFOTestSuite) TestDequeueMultipleGRs() {
	var (
		wg                     sync.WaitGroup
		totalElementsToEnqueue = 100
		totalElementsToDequeue = 90
	)

	for i := 0; i < totalElementsToEnqueue; i++ {
		suite.fifo.Enqueue(i)
	}

	for i := 0; i < totalElementsToDequeue; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := suite.fifo.Dequeue()
			suite.NoError(err, "Unexpected error during concurrent Dequeue()")
		}()
	}
	wg.Wait()

	// check len, should be == totalElementsToEnqueue - totalElementsToDequeue
	totalElementsAfterDequeue := suite.fifo.GetLen()
	suite.Equal(totalElementsToEnqueue-totalElementsToDequeue, totalElementsAfterDequeue, "Total elements on queue (after Dequeue) does not match with expected number")

	// check current first element
	val, err := suite.fifo.Dequeue()
	suite.NoError(err, "No error should be returned when dequeuing an existent element")
	suite.Equalf(totalElementsToDequeue, val, "The expected last element's value should be: %v", totalElementsToEnqueue-totalElementsToDequeue)
}

// ***************************************************************************************
// ** DequeueOrWaitForNextElement
// ***************************************************************************************

// single GR Locked queue
func (suite *FixedFIFOTestSuite) TestDequeueOrWaitForNextElementLockSingleGR() {
	suite.fifo.Lock()
	result, err := suite.fifo.DequeueOrWaitForNextElement()
	suite.Nil(result, "No value expected if queue is locked")
	suite.Error(err, "Locked queue does not allow to enqueue elements")

	// verify custom error: code: QueueErrorCodeLockedQueue
	customError, ok := err.(*QueueError)
	suite.True(ok, "Expected error type: QueueError")
	// verify custom error code
	suite.Equalf(QueueErrorCodeLockedQueue, customError.Code(), "Expected code: '%v'", QueueErrorCodeLockedQueue)
}

// single GR DequeueOrWaitForNextElement with a previous enqueued element
func (suite *FixedFIFOTestSuite) TestDequeueOrWaitForNextElementWithEnqueuedElementSingleGR() {
	value := 100
	len := suite.fifo.GetLen()
	suite.fifo.Enqueue(value)

	result, err := suite.fifo.DequeueOrWaitForNextElement()

	suite.NoError(err)
	suite.Equal(value, result)
	// length must be exactly the same as it was before
	suite.Equal(len, suite.fifo.GetLen())
}

// single GR DequeueOrWaitForNextElement 1 element
func (suite *FixedFIFOTestSuite) TestDequeueOrWaitForNextElementWithEmptyQueue() {
	var (
		value  = 100
		result interface{}
		err    error
		done   = make(chan struct{})
	)

	wait4gr := make(chan struct{})
	// waiting for next enqueued element
	go func(ready chan struct{}) {
		// let flow know this goroutine is ready
		wait4gr <- struct{}{}

		result, err = suite.fifo.DequeueOrWaitForNextElement()
		done <- struct{}{}
	}(wait4gr)

	// wait until listener goroutine is ready to dequeue/wait
	<-wait4gr

	// enqueue an element
	go func() {
		suite.fifo.Enqueue(value)
	}()

	select {
	// wait for the dequeued element
	case <-done:
		suite.NoError(err)
		suite.Equal(value, result)

	// the following comes first if more time than expected passed while waiting for the dequeued element
	case <-time.After(2 * time.Second):
		suite.Fail("too much time waiting for the enqueued element")

	}
}

// single GR calling DequeueOrWaitForNextElement (WaitForNextElementChanCapacity + 1) times, last one should return error
func (suite *FixedFIFOTestSuite) TestDequeueOrWaitForNextElementWithFullWaitingChannel() {
	// enqueue WaitForNextElementChanCapacity listeners to future enqueued elements
	for i := 0; i < WaitForNextElementChanCapacity; i++ {
		suite.fifo.waitForNextElementChan <- make(chan interface{})
	}

	result, err := suite.fifo.DequeueOrWaitForNextElement()
	suite.Nil(result)
	suite.Error(err)
	// verify custom error: code: QueueErrorCodeEmptyQueue
	customError, ok := err.(*QueueError)
	suite.True(ok, "Expected error type: QueueError")
	// verify custom error code
	suite.Equalf(QueueErrorCodeEmptyQueue, customError.Code(), "Expected code: '%v'", QueueErrorCodeEmptyQueue)
}

// multiple GRs, calling DequeueOrWaitForNextElement from different GRs and enqueuing the expected values later
func (suite *FixedFIFOTestSuite) TestDequeueOrWaitForNextElementMultiGR() {
	var (
		wg,
		// waitgroup to wait until goroutines are running
		wg4grs sync.WaitGroup
		// channel to enqueue dequeued values
		dequeuedValues = make(chan int, WaitForNextElementChanCapacity)
		// map[dequeued_value] = times dequeued
		mp = make(map[int]int)
	)

	for i := 0; i < WaitForNextElementChanCapacity; i++ {
		wg4grs.Add(1)
		go func() {
			wg4grs.Done()
			// wait for the next enqueued element
			result, err := suite.fifo.DequeueOrWaitForNextElement()
			// no error && no nil result
			suite.NoError(err)
			suite.NotNil(result)

			// send each dequeued element into the dequeuedValues channel
			resultInt, _ := result.(int)
			dequeuedValues <- resultInt

			// let the wg.Wait() know that this GR is done
			wg.Done()
		}()
	}

	// wait here until all goroutines (from above) are up and running
	wg4grs.Wait()

	// enqueue all needed elements
	for i := 0; i < WaitForNextElementChanCapacity; i++ {
		wg.Add(1)
		suite.fifo.Enqueue(i)
		// save the enqueued value as index
		mp[i] = 0
	}

	// wait until all GRs dequeue the elements
	wg.Wait()
	// close dequeuedValues channel in order to only read the previous enqueued values (from the channel)
	close(dequeuedValues)

	// verify that all enqueued values were dequeued
	for v := range dequeuedValues {
		val, ok := mp[v]
		suite.Truef(ok, "element dequeued but never enqueued: %v", val)
		// increment the m[p] value meaning the value p was dequeued
		mp[v] = val + 1
	}
	// verify there are no duplicates
	for k, v := range mp {
		suite.Equalf(1, v, "%v was dequeued %v times", k, v)
	}
}

// DequeueOrWaitForNextElement once queue's channel is closed
func (suite *FixedFIFOTestSuite) TestDequeueOrWaitForNextElementClosedChannel() {
	close(suite.fifo.queue)

	result, err := suite.fifo.DequeueOrWaitForNextElement()
	suite.Nil(result, "no result expected if internal queue's channel is closed")
	suite.Error(err, "error expected if internal queue's channel is closed")

	// verify custom error: code: QueueErrorCodeEmptyQueue
	customError, ok := err.(*QueueError)
	suite.True(ok, "Expected error type: QueueError")
	// verify custom error code
	suite.Equalf(QueueErrorCodeInternalChannelClosed, customError.Code(), "Expected code: '%v'", QueueErrorCodeInternalChannelClosed)
}

// ***************************************************************************************
// ** Lock / Unlock / IsLocked
// ***************************************************************************************

// single lock
func (suite *FixedFIFOTestSuite) TestLockSingleGR() {
	suite.fifo.Lock()
	suite.True(suite.fifo.IsLocked(), "fifo.isLocked has to be true after fifo.Lock()")
}

func (suite *FixedFIFOTestSuite) TestMultipleLockSingleGR() {
	for i := 0; i < 5; i++ {
		suite.fifo.Lock()
	}

	suite.True(suite.fifo.IsLocked(), "queue must be locked after Lock() operations")
}

// single unlock
func (suite *FixedFIFOTestSuite) TestUnlockSingleGR() {
	suite.fifo.Lock()
	suite.fifo.Unlock()
	suite.True(suite.fifo.IsLocked() == false, "fifo.isLocked has to be false after fifo.Unlock()")

	suite.fifo.Unlock()
	suite.True(suite.fifo.IsLocked() == false, "fifo.isLocked has to be false after fifo.Unlock()")
}

// ***************************************************************************************
// ** Context
// ***************************************************************************************

// context canceled while waiting for element to be added
func (suite *FixedFIFOTestSuite) TestContextCanceledAfter1Second() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		result interface{}
		err    error
		done   = make(chan struct{})
	)

	go func() {
		result, err = suite.fifo.DequeueOrWaitForNextElementContext(ctx)
		done <- struct{}{}
	}()

	select {
	// wait for the dequeue to finish
	case <-done:
		suite.True(err == context.DeadlineExceeded, "Canceling the context passed to fifo.DequeueOrWaitForNextElementContext() must cancel the dequeue wait", err)
		suite.Nil(result)

	// the following comes first if more time than expected happened while waiting for the dequeued element
	case <-time.After(2 * time.Second):
		suite.Fail("DequeueOrWaitForNextElementContext did not return immediately after context was canceled")
	}
}

// passing a already-canceled context to DequeueOrWaitForNextElementContext
func (suite *FixedFIFOTestSuite) TestContextAlreadyCanceled() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var (
		result interface{}
		err    error
		done   = make(chan struct{})
	)

	go func() {
		result, err = suite.fifo.DequeueOrWaitForNextElementContext(ctx)
		done <- struct{}{}
	}()

	select {
	// wait for the dequeue to finish
	case <-done:
		suite.True(err == context.Canceled, "Canceling the context passed to fifo.DequeueOrWaitForNextElementContext() must cancel the dequeue wait", err)
		suite.Nil(result)

	// the following comes first if more time than expected happened while waiting for the dequeued element
	case <-time.After(2 * time.Second):
		suite.Fail("DequeueOrWaitForNextElementContext did not return immediately after context was canceled")
	}
}