package goconcurrentqueue

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type QueueErrorTestSuite struct {
	suite.Suite
	queueError *QueueError
}

// ***************************************************************************************
// ** Initialization
// ***************************************************************************************

// no elements at initialization
func (suite *QueueErrorTestSuite) TestInitialization() {
	queueError := NewQueueError("code", "message")

	suite.Equal("code", queueError.Code())
	suite.Equal("message", queueError.Error())
}

// ***************************************************************************************
// ** Run suite
// ***************************************************************************************

func TestQueueErrorTestSuite(t *testing.T) {
	suite.Run(t, new(QueueErrorTestSuite))
}
