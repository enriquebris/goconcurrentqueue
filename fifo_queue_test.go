package goconcurrentqueue

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

const (
	testValue = "test value"
)

type FIFOTestSuite struct {
	suite.Suite
	fifo *FIFO
}

func (suite *FIFOTestSuite) SetupTest() {
	suite.fifo = NewFIFO()
}

// ***************************************************************************************
// ** Queue initialization
// ***************************************************************************************

// no elements at initialization
func (suite *FIFOTestSuite) TestNoElementsAtInitialization() {
	len := suite.fifo.GetLen()
	suite.Equalf(0, len, "No elements expected at initialization, currently: %v", len)
}

// unlocked at initialization
func (suite *FIFOTestSuite) TestNoLockedAtInitialization() {
	suite.True(suite.fifo.IsLocked() == false, "Queue must be unlocked at initialization")
}

// ***************************************************************************************
// ** Enqueue && GetLen
// ***************************************************************************************

// single enqueue lock verification
func (suite *FIFOTestSuite) TestEnqueueLockSingleGR() {
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
func (suite *FIFOTestSuite) TestEnqueueLenSingleGR() {
	suite.fifo.Enqueue(testValue)
	len := suite.fifo.GetLen()
	suite.Equalf(1, len, "Expected number of elements in queue: 1, currently: %v", len)

	suite.fifo.Enqueue(5)
	len = suite.fifo.GetLen()
	suite.Equalf(2, len, "Expected number of elements in queue: 2, currently: %v", len)
}

// single enqueue and wait for next element
func (suite *FIFOTestSuite) TestEnqueueWaitForNextElementSingleGR() {
	waitForNextElement := make(chan interface{})
	// add the listener manually (ONLY for testings purposes)
	suite.fifo.waitForNextElementChan <- waitForNextElement

	value := 100
	// enqueue from a different GR to avoid blocking the listener channel
	go suite.fifo.Enqueue(value)
	// wait for the enqueued element
	result := <-waitForNextElement

	suite.Equal(value, result)
}

// TestEnqueueLenMultipleGR enqueues elements concurrently
//
// Detailed steps:
//	1 - Enqueue totalGRs concurrently (from totalGRs different GRs)
//	2 - Verifies the len, it should be equal to totalGRs
//	3 - Verifies that all elements from 0 to totalGRs were enqueued
func (suite *FIFOTestSuite) TestEnqueueLenMultipleGR() {
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
	sl2check := make([]bool, totalGRs)

	for i := 0; i < totalElements; i++ {
		tmpVal, err = suite.fifo.Get(i)
		suite.NoError(err, "No error should be returned trying to get an existent element")

		val = tmpVal.(int)
		if !sl2check[val] {
			totalElementsVerified++
			sl2check[val] = true
		} else {
			suite.Failf("Duplicated element", "Unexpected duplicated value: %v", val)
		}
	}
	suite.True(totalElementsVerified == totalGRs, "Enqueued elements are missing")
}

// call GetLen concurrently
func (suite *FIFOTestSuite) TestGetLenMultipleGRs() {
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
func (suite *FIFOTestSuite) TestGetCapSingleGR() {
	// initial capacity
	suite.Equal(cap(suite.fifo.slice), suite.fifo.GetCap(), "unexpected capacity")

	// checking after adding 2 items
	suite.fifo.Enqueue(1)
	suite.fifo.Enqueue(2)
	suite.Equal(cap(suite.fifo.slice), suite.fifo.GetCap(), "unexpected capacity")
}

// ***************************************************************************************
// ** Get
// ***************************************************************************************

// get an element
func (suite *FIFOTestSuite) TestGetLockSingleGR() {
	suite.fifo.Enqueue(1)
	_, err := suite.fifo.Get(0)
	suite.NoError(err, "Unlocked queue allows to get elements")

	suite.fifo.Lock()
	_, err = suite.fifo.Get(0)
	suite.Error(err, "Locked queue does not allow to get elements")

	// verify custom error: code: QueueErrorCodeLockedQueue
	customError, ok := err.(*QueueError)
	suite.True(ok, "Expected error type: QueueError")
	// verify custom error code
	suite.Equalf(QueueErrorCodeLockedQueue, customError.Code(), "Expected code: '%v'", QueueErrorCodeLockedQueue)
}

// get a valid element
func (suite *FIFOTestSuite) TestGetSingleGR() {
	suite.fifo.Enqueue(testValue)
	val, err := suite.fifo.Get(0)

	// verify error (should be nil)
	suite.NoError(err, "No error should be enqueueing an element")

	// verify element's value
	suite.Equalf(testValue, val, "Different element returned: %v", val)
}

// get a invalid element
func (suite *FIFOTestSuite) TestGetInvalidElementSingleGR() {
	suite.fifo.Enqueue(testValue)
	val, err := suite.fifo.Get(1)

	// verify error
	suite.Error(err, "An error should be returned after ask for a no existent element")
	// verify custom error: code: QueueErrorCodeLockedQueue
	customError, ok := err.(*QueueError)
	suite.True(ok, "Expected error type: QueueError")
	// verify custom error code
	suite.Equalf(QueueErrorCodeIndexOutOfBounds, customError.Code(), "Expected code: '%v'", QueueErrorCodeIndexOutOfBounds)

	// verify element's value
	suite.Equalf(val, nil, "Nil should be returned, currently returned: %v", val)
}

// call Get concurrently
func (suite *FIFOTestSuite) TestGetMultipleGRs() {
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
			val, err := suite.fifo.Get(5)

			suite.NoError(err, "No error should be returned trying to get an existent element")
			suite.Equal(5, val.(int), "Expected element's value: 5")
		}()
	}
	wg.Wait()

	total := suite.fifo.GetLen()
	suite.Equalf(totalElementsToEnqueue, total, "Expected len: %v", totalElementsToEnqueue)
}

// ***************************************************************************************
// ** Remove
// ***************************************************************************************

// single remove lock verification
func (suite *FIFOTestSuite) TestRemoveLockSingleGR() {
	suite.fifo.Enqueue(1)
	suite.NoError(suite.fifo.Remove(0), "Unlocked queue allows to remove elements")

	suite.fifo.Enqueue(1)
	suite.fifo.Lock()
	err := suite.fifo.Remove(0)
	suite.Error(err, "Locked queue does not allow to remove elements")

	// verify custom error: code: QueueErrorCodeLockedQueue
	customError, ok := err.(*QueueError)
	suite.True(ok, "Expected error type: QueueError")
	// verify custom error code
	suite.Equalf(QueueErrorCodeLockedQueue, customError.Code(), "Expected code: '%v'", QueueErrorCodeLockedQueue)
}

// remove elements
func (suite *FIFOTestSuite) TestRemoveSingleGR() {
	suite.fifo.Enqueue(testValue)
	suite.fifo.Enqueue(5)

	// removing first element
	err := suite.fifo.Remove(0)
	suite.NoError(err, "Unexpected error")

	// get element at index 0
	val, err2 := suite.fifo.Get(0)
	suite.NoError(err2, "Unexpected error")
	suite.Equal(5, val, "Queue returned the wrong element")

	// remove element having a wrong index
	err3 := suite.fifo.Remove(1)
	suite.Errorf(err3, "The index of the element to remove is out of bounds")
	// verify custom error: code: QueueErrorCodeLockedQueue
	customError, ok := err3.(*QueueError)
	suite.True(ok, "Expected error type: QueueError")
	// verify custom error code
	suite.Equalf(QueueErrorCodeIndexOutOfBounds, customError.Code(), "Expected code: '%v'", QueueErrorCodeIndexOutOfBounds)
}

// TestRemoveMultipleGRs removes elements concurrently.
//
// Detailed steps:
//	1 - Enqueues totalElementsToEnqueue consecutive elements (0, 1, 2, 3, ... totalElementsToEnqueue - 1)
//	2 - Hits fifo.Remove(1) concurrently from totalElementsToRemove different GRs
//	3 - Verifies the final len == totalElementsToEnqueue - totalElementsToRemove
//	4 - Verifies that final 2nd element == (1 + totalElementsToRemove)
func (suite *FIFOTestSuite) TestRemoveMultipleGRs() {
	var (
		wg                     sync.WaitGroup
		totalElementsToEnqueue = 100
		totalElementsToRemove  = 90
	)

	for i := 0; i < totalElementsToEnqueue; i++ {
		suite.fifo.Enqueue(i)
	}

	for i := 0; i < totalElementsToRemove; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := suite.fifo.Remove(1)
			suite.NoError(err, "Unexpected error during concurrent Remove(n)")
		}()
	}
	wg.Wait()

	// check len, should be == totalElementsToEnqueue - totalElementsToRemove
	totalElementsAfterRemove := suite.fifo.GetLen()
	suite.Equal(totalElementsToEnqueue-totalElementsToRemove, totalElementsAfterRemove, "Total elements on list does not match with expected number")

	// check current 2nd element (index 1) on the queue
	val, err := suite.fifo.Get(1)
	suite.NoError(err, "No error should be returned when getting an existent element")
	suite.Equalf(1+totalElementsToRemove, val, "The expected value at position 1 (2nd element) should be: %v", 1+totalElementsToRemove)
}

// ***************************************************************************************
// ** Dequeue
// ***************************************************************************************

// single dequeue lock verification
func (suite *FIFOTestSuite) TestDequeueLockSingleGR() {
	suite.fifo.Enqueue(1)
	_, err := suite.fifo.Dequeue()
	suite.NoError(err, "Unlocked queue allows to dequeue elements")

	suite.fifo.Enqueue(1)
	suite.fifo.Lock()
	_, err = suite.fifo.Dequeue()
	// error expected
	suite.Error(err, "Locked queue does not allow to dequeue elements")
	// verify custom error: code: QueueErrorCodeLockedQueue
	customError, ok := err.(*QueueError)
	suite.True(ok, "Expected error type: QueueError")
	// verify custom error code
	suite.Equalf(QueueErrorCodeLockedQueue, customError.Code(), "Expected code: '%v'", QueueErrorCodeLockedQueue)
}

// dequeue an empty queue
func (suite *FIFOTestSuite) TestDequeueEmptyQueueSingleGR() {
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
func (suite *FIFOTestSuite) TestDequeueSingleGR() {
	suite.fifo.Enqueue(testValue)
	suite.fifo.Enqueue(5)

	// get the first element
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

// TestDequeueMultipleGRs dequeues elements concurrently
//
// Detailed steps:
//	1 - Enqueues totalElementsToEnqueue consecutive integers
//	2 - Dequeues totalElementsToDequeue concurrently from totalElementsToDequeue GRs
//	3 - Verifies the final len, should be equal to totalElementsToEnqueue - totalElementsToDequeue
//	4 - Verifies that the next dequeued element's value is equal to totalElementsToDequeue
func (suite *FIFOTestSuite) TestDequeueMultipleGRs() {
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
func (suite *FIFOTestSuite) TestDequeueOrWaitForNextElementLockSingleGR() {
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
func (suite *FIFOTestSuite) TestDequeueOrWaitForNextElementWithEnqueuedElementSingleGR() {
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
func (suite *FIFOTestSuite) TestDequeueOrWaitForNextElementWithEmptyQueue() {
	var (
		value  = 100
		result interface{}
		err    error
		done   = make(chan struct{})
	)

	// waiting for next enqueued element
	go func() {
		result, err = suite.fifo.DequeueOrWaitForNextElement()
		done <- struct{}{}
	}()

	// enqueue an element
	go func() {
		suite.fifo.Enqueue(value)
	}()

	select {
	// wait for the dequeued element
	case <-done:
		suite.NoError(err)
		suite.Equal(value, result)

	// the following comes first if more time than expected happened while waiting for the dequeued element
	case <-time.After(2 * time.Second):
		suite.Fail("too much time waiting for the enqueued element")

	}
}

// single GR calling DequeueOrWaitForNextElement (WaitForNextElementChanCapacity + 1) times, last one should return error
func (suite *FIFOTestSuite) TestDequeueOrWaitForNextElementWithFullWaitingChannel() {
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
func (suite *FIFOTestSuite) TestDequeueOrWaitForNextElementMultiGR() {
	var (
		wg sync.WaitGroup
		// channel to enqueue dequeued values
		dequeuedValues = make(chan int, WaitForNextElementChanCapacity)
		// map[dequeued_value] = times dequeued
		mp = make(map[int]int)
	)

	for i := 0; i < WaitForNextElementChanCapacity; i++ {
		go func() {
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
	// verify that there are no duplicates
	for k, v := range mp {
		suite.Equalf(1, v, "%v was dequeued %v times", k, v)
	}
}

// multiple GRs, calling DequeueOrWaitForNextElement from different GRs and enqueuing the expected values later
func (suite *FIFOTestSuite) TestDequeueOrWaitForNextElementMultiGR2() {
	var (
		done       = make(chan int, 10)
		total      = 2000
		results    = make(map[int]struct{})
		totalOk    = 0
		totalError = 0
	)

	go func(fifo Queue, done chan int, total int) {
		for i := 0; i < total; i++ {
			go func(queue Queue, done chan int) {
				rawValue, err := fifo.DequeueOrWaitForNextElement()
				if err != nil {
					fmt.Println(err)
					// error
					done <- -1
				} else {
					val, _ := rawValue.(int)
					done <- val
				}
			}(fifo, done)

			go func(queue Queue, value int) {
				if err := fifo.Enqueue(value); err != nil {
					fmt.Println(err)
				}
			}(fifo, i)
		}
	}(suite.fifo, done, total)

	i := 0
	for {
		v := <-done
		if v != -1 {
			totalOk++
			_, ok := results[v]
			suite.Falsef(ok, "duplicated value %v", v)

			results[v] = struct{}{}
		} else {
			totalError++
		}

		i++
		if i == total {
			break
		}
	}

	suite.Equal(total, totalError+totalOk)
}

// call DequeueOrWaitForNextElement(), wait some time and enqueue an item
func (suite *FIFOTestSuite) TestDequeueOrWaitForNextElementGapSingleGR() {
	var (
		expectedValue = 50
		done          = make(chan struct{}, 3)
	)

	// DequeueOrWaitForNextElement()
	go func(fifo *FIFO, done chan struct{}) {
		val, err := fifo.DequeueOrWaitForNextElement()
		suite.NoError(err)
		suite.Equal(expectedValue, val)
		done <- struct{}{}
	}(suite.fifo, done)

	// wait and Enqueue function
	go func(fifo *FIFO, done chan struct{}) {
		time.Sleep(time.Millisecond * dequeueOrWaitForNextElementInvokeGapTime * dequeueOrWaitForNextElementInvokeGapTime)
		suite.NoError(fifo.Enqueue(expectedValue))
		done <- struct{}{}
	}(suite.fifo, done)

	for i := 0; i < 2; i++ {
		select {
		case <-done:
		case <-time.After(2 * time.Millisecond * dequeueOrWaitForNextElementInvokeGapTime * dequeueOrWaitForNextElementInvokeGapTime):
			suite.FailNow("Too much time waiting for the value")
		}
	}
}

// ***************************************************************************************
// ** Lock / Unlock / IsLocked
// ***************************************************************************************

// single lock
func (suite *FIFOTestSuite) TestLockSingleGR() {
	suite.fifo.Lock()
	suite.True(suite.fifo.isLocked == true, "fifo.isLocked has to be true after fifo.Lock()")
}

// single unlock
func (suite *FIFOTestSuite) TestUnlockSingleGR() {
	suite.fifo.Lock()
	suite.fifo.Unlock()
	suite.True(suite.fifo.isLocked == false, "fifo.isLocked has to be false after fifo.Unlock()")
}

// single isLocked
func (suite *FIFOTestSuite) TestIsLockedSingleGR() {
	suite.True(suite.fifo.isLocked == suite.fifo.IsLocked(), "fifo.IsLocked() has to be equal to fifo.isLocked")

	suite.fifo.Lock()
	suite.True(suite.fifo.isLocked == suite.fifo.IsLocked(), "fifo.IsLocked() has to be equal to fifo.isLocked")

	suite.fifo.Unlock()
	suite.True(suite.fifo.isLocked == suite.fifo.IsLocked(), "fifo.IsLocked() has to be equal to fifo.isLocked")
}

// ***************************************************************************************
// ** Context
// ***************************************************************************************

// context canceled while waiting for element to be added
func (suite *FIFOTestSuite) TestContextCanceledAfter1Second() {
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
func (suite *FIFOTestSuite) TestContextAlreadyCanceled() {
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

// ***************************************************************************************
// ** Run suite
// ***************************************************************************************

func TestFIFOTestSuite(t *testing.T) {
	suite.Run(t, new(FIFOTestSuite))
}
