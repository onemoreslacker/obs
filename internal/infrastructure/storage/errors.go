package storage

type storageError struct{ msg string }

func (e storageError) Error() string { return e.msg }

var (
	ErrChatAlreadyExists = storageError{msg: "error: chat already exists"}
	ErrChatNotFound      = storageError{msg: "error: chat was not found"}
	ErrLinkAlreadyExists = storageError{msg: "error: link already exists"}
	ErrLinkNotFound      = storageError{msg: "error: link not found"}
)
