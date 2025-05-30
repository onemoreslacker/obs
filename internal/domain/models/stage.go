package models

type Stage struct {
	Prompt   string
	Manual   string
	Validate func(msg string) error
}

func NewStage(prompt, manual string, validate func(msg string) error) *Stage {
	return &Stage{
		Prompt:   prompt,
		Manual:   manual,
		Validate: validate,
	}
}
