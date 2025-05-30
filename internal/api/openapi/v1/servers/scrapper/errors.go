package scrapperapi

type scrapperError struct{ msg string }

func (e scrapperError) Error() string { return e.msg }

var (
	ErrInvalidBody               = scrapperError{msg: "Некорректные параметры запроса"}
	ErrChatAlreadyExists         = scrapperError{msg: "error: chat already exists"}
	ErrChatNotExists             = scrapperError{msg: "error: chat not exists"}
	ErrLinkAlreadyExists         = scrapperError{msg: "error: link already exists"}
	ErrLinkNotExists             = scrapperError{msg: "error: link does not exist"}
	ErrAddLinkInvalidLink        = scrapperError{msg: "error: link is invalid or missing"}
	ErrAddLinkFailed             = scrapperError{msg: "error: failed to add link to db"}
	ErrGetLinksFailed            = scrapperError{msg: "error: failed to get links"}
	ErrDeleteLinkInvalidLink     = scrapperError{msg: "error: link is invalid or missing"}
	ErrAddTagFailed              = scrapperError{msg: "error: failed to "}
	ErrSubscriptionAlreadyExists = scrapperError{msg: "error: link is already being tracked"}
	ErrSubscriptionsNotExists    = scrapperError{msg: "error: link is not yet being tracked"}
	ErrTagAlreadyExists          = scrapperError{msg: "error: tag already exists"}
	ErrFilterAlreadyExists       = scrapperError{msg: "error: filter already exists"}
)
