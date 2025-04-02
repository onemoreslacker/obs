package bot

type botError struct{ msg string }

func (e botError) Error() string { return e.msg }

var (
	ErrUserNotRegistered = botError{msg: "error: user is not registered"}
)
