package commands

type Command interface {
	Stage() (string, bool)
	Validate(input string) error
	Done() bool
	Request() (any, error)
	Name() string
}
