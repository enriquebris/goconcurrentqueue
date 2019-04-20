package goconcurrentqueue

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type NotImplementedErrorTestSuite struct {
	suite.Suite
	err error
}

func (suite *NotImplementedErrorTestSuite) TestConstructor() {
	err := newNotImplementedError()
	_, ok := err.(NotImplementedError)

	suite.True(ok, "returned error could be used as a NotImplementedError (after type assertion)")
}

func (suite *NotImplementedErrorTestSuite) TestError() {
	suite.err = newNotImplementedError()
	suite.Equal("not implemented", suite.err.Error(), "unexpected message")
}

// ***************************************************************************************
// ** Run suite
// ***************************************************************************************

func TestNotImplementedErrorTestSuite(t *testing.T) {
	suite.Run(t, new(NotImplementedErrorTestSuite))
}
