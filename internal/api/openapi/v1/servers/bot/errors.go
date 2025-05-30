package botapi

type botapiError struct{ msg string }

func (e botapiError) Error() string { return e.msg }

var (
	ErrMissingDescription = botapiError{msg: "error: description parameter is missing"}
	ErrMissingChatIDs     = botapiError{msg: "error: chat ids parameter is missing"}
	ErrMissingURL         = botapiError{msg: "error: url parameter is missing"}
	ErrUnknownURL         = botapiError{msg: "error: unknown url"}
	ErrBotUpdates         = botapiError{msg: "Некорректные параметры запроса"}
)
