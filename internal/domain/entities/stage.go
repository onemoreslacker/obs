package entities

// Stage stores related to the command data.
type Stage struct {
	Prompt   string
	Manual   string
	Validate func(msg string) error
}

// NewStage instantiates a new Stage entity.
func NewStage(
	prompt, manual string,
	validate func(msg string) error,
) *Stage {
	return &Stage{
		Prompt:   prompt,
		Manual:   manual,
		Validate: validate,
	}
}
