package updater

import "fmt"

type updaterError struct{ msg string }

func (e updaterError) Error() string { return fmt.Sprintf("error: %s", e.msg) }

var ErrUnknownTransportMode = updaterError{msg: "unknown transport mode"}
var ErrSendUpdate = updaterError{msg: "failed to send update to http server or kafka"}
var ErrHTTPSendUpdate = updaterError{msg: "failed to send update to http server"}
