package goconcurrentqueue

import (
	"sync"
	"testing"

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
	suite.Error(suite.fifo.Enqueue(1), "Locked queue does not allow to enqueue elements")
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

	// verify element's value
	suite.Equalf(val, nil, "Nil should be returner, currently returned: %v", val)
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
	suite.Error(suite.fifo.Remove(0), "Locked queue does not allow to remove elements")
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
	suite.Error(err, "Locked queue does not allow to dequeue elements")
}

// dequeue an empty queue
func (suite *FIFOTestSuite) TestDequeueEmptyQueueSingleGR() {
	val, err := suite.fifo.Dequeue()
	suite.Errorf(err, "Can't dequeue an empty queue")
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
// ** Run suite
// ***************************************************************************************

func TestFIFOTestSuite(t *testing.T) {
	suite.Run(t, new(FIFOTestSuite))
}
