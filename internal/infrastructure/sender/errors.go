package sender

import "fmt"

type senderError struct{ msg string }

func (e senderError) Error() string { return fmt.Sprintf("error: %s", e.msg) }

var ErrUnknownTransportMode = senderError{msg: "unknown transport mode"}
