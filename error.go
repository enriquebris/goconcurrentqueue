package goconcurrentqueue

type NotImplementedError struct {
}

func newNotImplementedError() error {
	return NotImplementedError{}
}

func (st NotImplementedError) Error() string {
	return "not implemented"
}
