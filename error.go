package goconcurrentqueue

const (
	QueueErrorCodeEmptyQueue            = "empty-queue"
	QueueErrorCodeLockedQueue           = "locked-queue"
	QueueErrorCodeIndexOutOfBounds      = "index-out-of-bounds"
	QueueErrorCodeFullCapacity          = "full-capacity"
	QueueErrorCodeInternalChannelClosed = "internal-channel-closed"
	QueueErrorCodeIndexesMatch          = "indexes-match"
	QueueErrorCodeIndexFirstPosition    = "index-first-position"
	QueueErrorCodeIndexLastPosition     = "index-last-position"
)

type QueueError struct {
	code    string
	message string
}

func NewQueueError(code string, message string) *QueueError {
	return &QueueError{
		code:    code,
		message: message,
	}
}

func (st *QueueError) Error() string {
	return st.message
}

func (st *QueueError) Code() string {
	return st.code
}
