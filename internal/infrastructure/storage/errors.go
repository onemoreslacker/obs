package storage

type storageError struct{ msg string }

func (e storageError) Error() string { return e.msg }

var (
	ErrUnknownDBAccessType = storageError{msg: "error: unknown database access type"}
)
