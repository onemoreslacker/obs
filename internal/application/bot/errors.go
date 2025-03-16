package bot

type botError struct{ msg string }

func (e botError) Error() string { return e.msg }

var (
	ErrFailedToFormatLinks = botError{msg: "error: failed to format links"}
	ErrUserNotRegistered   = botError{msg: "error: user is not registered"}
	ErrUnknownCommand      = botError{msg: "error: unknown command"}
)
