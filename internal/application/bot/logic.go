package bot

import (
	"context"
	"net/http"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/commands"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/entities"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) identifyCommand(chatID int64, input string) error {
	var err error

	switch input {
	case b.cfg.Meta.Commands.Track:
		b.currentCommand, err = b.authorizationMiddleware(chatID, func() (commands.Command, error) {
			return commands.NewCommandTrack(
				chatID,
				b.scrapperClient,
				b.cfg,
			), nil
		})()

	case b.cfg.Meta.Commands.Untrack:
		b.currentCommand, err = b.authorizationMiddleware(chatID, func() (commands.Command, error) {
			return commands.NewCommandUntrack(
				chatID,
				b.scrapperClient,
				b.cfg,
			), nil
		})()

	case b.cfg.Meta.Commands.List:
		b.currentCommand, err = b.authorizationMiddleware(chatID, func() (commands.Command, error) {
			return commands.NewCommandList(
				chatID,
				b.scrapperClient,
				b.cfg,
			), nil
		})()

	default:
		err = ErrUnknownCommand
	}

	return err
}

func (b *Bot) commandTermination(chatID int64) error {
	defer func() { b.currentCommand = nil }()

	var handler entities.Handler

	switch b.currentCommand.Name() {
	case b.cfg.Meta.Commands.Track:
		handler = entities.Handler{
			FailMsg:    b.cfg.Meta.Fails.Track,
			SuccessMsg: b.cfg.Meta.Replies.Tracking,
		}

	case b.cfg.Meta.Commands.Untrack:
		handler = entities.Handler{
			FailMsg:    b.cfg.Meta.Fails.Untrack,
			SuccessMsg: b.cfg.Meta.Replies.Untracked,
		}

	case b.cfg.Meta.Commands.List:
		handler = entities.Handler{
			FailMsg: b.cfg.Meta.Fails.List,
			Processor: func(res any) (string, error) {
				list, ok := res.([]entities.Link)
				if !ok {
					return "", ErrFailedToFormatLinks
				}

				return ConstructListMessage(list), nil
			},
		}

	default:
		return ErrUnknownCommand
	}

	return b.executeCommand(chatID, handler)
}

func (b *Bot) executeCommand(chatID int64, handler entities.Handler) error {
	res, err := b.currentCommand.Request()

	if err != nil {
		if _, err := b.tgb.Send(tgbotapi.NewMessage(chatID, handler.FailMsg)); err != nil {
			return err
		}

		return nil
	}

	var msg string

	if handler.Processor != nil {
		processed, err := handler.Processor(res)
		if err != nil {
			return err
		}

		msg = processed
	} else {
		msg = handler.SuccessMsg
	}

	if _, err := b.tgb.Send(tgbotapi.NewMessage(chatID, msg)); err != nil {
		return err
	}

	return nil
}
func (b *Bot) registerChat(id int64) (string, error) {
	resp, err := b.scrapperClient.PostTgChatId(context.Background(), id)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest {
		return b.cfg.Meta.Fails.Registration, nil
	}

	return b.cfg.Meta.Replies.Registration, nil
}

func (b *Bot) isRegistered(id int64) error {
	resp, err := b.scrapperClient.GetTgChatId(context.Background(), id)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest {
		return ErrUserNotRegistered
	}

	return nil
}
