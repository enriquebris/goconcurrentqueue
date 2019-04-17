package goconcurrentqueue

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"
)

type FixedFIFOTestSuite struct {
	suite.Suite
	fifo *FixedFIFO
}

func (suite *FixedFIFOTestSuite) SetupTest() {
	suite.fifo = NewFixedFIFO(500)
}

// ***************************************************************************************
// ** Run suite
// ***************************************************************************************

func TestFixedFIFOTestSuite(t *testing.T) {
	suite.Run(t, new(FIFOTestSuite))
}

// ***************************************************************************************
// ** Enqueue && GetLen
// ***************************************************************************************

// single enqueue lock verification
func (suite *FixedFIFOTestSuite) TestEnqueueLockSingleGR() {
	suite.NoError(suite.fifo.Enqueue(1), "Unlocked queue allows to enqueue elements")

	suite.fifo.Lock()
	suite.Error(suite.fifo.Enqueue(1), "Locked queue does not allow to enqueue elements")
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
