package updater

type scrapperError struct{ msg string }

func (e scrapperError) Error() string { return e.msg }

var (
	ErrUnknownService = scrapperError{msg: "error: unknown service"}
)
