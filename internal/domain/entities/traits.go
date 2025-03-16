package entities

type Traits struct {
	Stage     int
	Span      int
	ChatID    int64
	Malformed bool
	Name      string
}

func NewTraits(span int, chatID int64, name string) *Traits {
	return &Traits{
		Span:   span,
		ChatID: chatID,
		Name:   name,
	}
}
