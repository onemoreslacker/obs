package notifier

type notifierError struct{ msg string }

func (e notifierError) Error() string { return e.msg }

var (
	ErrEmptyUpdates = notifierError{msg: "no new activity via given link"}
)
