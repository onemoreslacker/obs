package commands

type commandsError struct{ msg string }

func (e commandsError) Error() string { return e.msg }

var (
	ErrInvalidLinkFormat    = commandsError{msg: "error: provided link does not satisfy either format"}
	ErrInvalidTagsFormat    = commandsError{msg: "error: provided tags do not satisfy specified format"}
	ErrInvalidFiltersFormat = commandsError{msg: "error: provided fileters do not satisfy specified format"}
	ErrInvalidAck           = commandsError{msg: "error: your acknowledgment should be either yes or no"}
)
