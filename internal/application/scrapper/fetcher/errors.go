package fetcher

import (
	"fmt"
)

type fetcherErr struct{ msg string }

func (e fetcherErr) Error() string { return fmt.Sprintf("error: %s", e.msg) }

var (
	ErrInvalidRepoPath = fetcherErr{msg: "invalid repo path"}
)
