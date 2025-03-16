package config

type configError struct{ msg string }

func (e configError) Error() string { return e.msg }

var (
	ErrFailedToFindEnv = configError{msg: "error: failed to find .env"}
)
