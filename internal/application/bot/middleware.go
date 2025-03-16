package bot

import "github.com/es-debug/backend-academy-2024-go-template/internal/domain/commands"

func (b *Bot) authorizationMiddleware(id int64, constructor func() (
	commands.Command, error)) func() (commands.Command, error) {
	return func() (commands.Command, error) {
		if err := b.isRegistered(id); err != nil {
			return nil, err
		}

		return constructor()
	}
}
