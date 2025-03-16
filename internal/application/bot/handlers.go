package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) singleStepHandler(chatID int64, input string) string {
	var (
		reply string
		err   error
	)

	switch input {
	case b.cfg.Meta.Commands.Start:
		reply, err = b.registerChat(chatID)

	case b.cfg.Meta.Commands.Help:
		reply = b.withAuthorization(chatID, b.constructHelpMessage)

	case b.cfg.Meta.Commands.Cancel:
		reply = b.withAuthorization(chatID, func() string {
			return b.cfg.Meta.Fails.Cancel
		})

	default:
		reply = b.withAuthorization(chatID, func() string {
			return b.cfg.Meta.Fails.Unknown
		})
	}

	if err != nil {
		return err.Error()
	}

	return reply
}

func (b *Bot) commandFlowHandler(chatID int64, input string) error {
	var reply tgbotapi.MessageConfig

	if err := b.currentCommand.Validate(input); err != nil {
		reply = b.configureReply(chatID)
		if _, err := b.tgb.Send(reply); err != nil {
			return err
		}

		return nil
	}

	if b.currentCommand.Done() {
		if err := b.commandTermination(chatID); err != nil {
			return err
		}

		return nil
	}

	reply = b.configureReply(chatID)
	if _, err := b.tgb.Send(reply); err != nil {
		return err
	}

	return nil
}

func (b *Bot) QueryHandler(query *tgbotapi.CallbackQuery) error {
	chatID, input := query.Message.Chat.ID, query.Data

	callback := tgbotapi.NewCallback(query.ID, input)
	if _, err := b.tgb.Request(callback); err != nil {
		return err
	}

	if b.currentCommand == nil {
		if _, err := b.tgb.Send(
			tgbotapi.NewMessage(chatID, b.cfg.Meta.Fails.Ack),
		); err != nil {
			return err
		}

		return nil
	}

	return b.commandFlowHandler(chatID, input)
}

func (b *Bot) MessageHandler(chatID int64, input string) error {
	var reply tgbotapi.MessageConfig

	if b.currentCommand == nil {
		if err := b.identifyCommand(chatID, input); err != nil {
			reply = tgbotapi.NewMessage(chatID, b.singleStepHandler(chatID, input))
		} else {
			reply = b.configureReply(chatID)
		}

		if _, err := b.tgb.Send(reply); err != nil {
			return err
		}

		return nil
	}

	if input == b.cfg.Meta.Commands.Cancel {
		reply = tgbotapi.NewMessage(chatID, b.cfg.Meta.Replies.Cancel)
		if _, err := b.tgb.Send(reply); err != nil {
			return err
		}

		b.currentCommand = nil

		return nil
	}

	return b.commandFlowHandler(chatID, input)
}
