package receiver

import "fmt"

type receiverError struct{ msg string }

func (e receiverError) Error() string { return fmt.Sprintf("error: %s", e.msg) }

var ErrUnknownTransportMode = receiverError{msg: "unknown transport mode"}
