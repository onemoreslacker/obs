package clients

type clientError struct{ msg string }

func (e clientError) Error() string { return e.msg }

var (
	ErrRequestFailed       = clientError{msg: "error: request failed"}
	ErrNonRetryableRequest = clientError{msg: "error: request is non-retryable"}
)
