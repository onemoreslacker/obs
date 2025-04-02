package scrapperapi

type scrapperError struct{ msg string }

func (e scrapperError) Error() string { return e.msg }

var (
	ErrInvalidBody           = scrapperError{msg: "Некорректные параметры запроса"}
	ErrChatAlreadyExists     = scrapperError{msg: "error: chat already exists"}
	ErrChatNotFound          = scrapperError{msg: "error: chat was not found"}
	ErrLinkAlreadyExists     = scrapperError{msg: "error: link already exists"}
	ErrLinkNotFound          = scrapperError{msg: "error: link not found"}
	ErrAddLinkInvalidLink    = scrapperError{msg: "error: link is invalid or missing"}
	ErrAddLinkInvalidTags    = scrapperError{msg: "error: tags are invalid or missing"}
	ErrAddLinkInvalidFilters = scrapperError{msg: "error: filters are invalid or missing"}
	ErrAddLinkFailed         = scrapperError{msg: "error: failed to add link to db"}
	ErrGetLinksFailed        = scrapperError{msg: "error: failed to get links"}
	ErrDeleteLinkFailed      = scrapperError{msg: "error: failed to delete link"}
	ErrDeleteLinkInvalidLink = scrapperError{msg: "error: link is invalid or missing"}
)
