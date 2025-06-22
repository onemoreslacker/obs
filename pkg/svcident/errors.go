package svcident

import "fmt"

type svcidentErr struct{ msg string }

func (s svcidentErr) Error() string { return fmt.Sprintf("error: %s", s.msg) }

var ErrUnknownService = svcidentErr{msg: "unknown service"}
