package commands

import "fmt"

type commandsError struct{ msg string }

func (e commandsError) Error() string { return fmt.Sprintf("error: %s", e.msg) }

var (
	ErrInvalidLinkFormat    = commandsError{msg: "link does not satisfy either format"}
	ErrInvalidTagsFormat    = commandsError{msg: "tags do not satisfy specified format"}
	ErrInvalidFiltersFormat = commandsError{msg: "filters do not satisfy specified format"}
	ErrInvalidAck           = commandsError{msg: "your acknowledgment should be either yes or no"}
	ErrLinkAlreadyExists    = commandsError{msg: "link is already being tracked"}
	ErrLinkNotExists        = commandsError{msg: "link is not yet begin tracked"}
)
