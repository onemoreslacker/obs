package telebot_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"

	mocks "github.com/es-debug/backend-academy-2024-go-template/internal/application/bot/mocks"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/bot/telebot"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCommandRequest(t *testing.T) {
	cmd := mocks.NewMockCommand(t)
	defer cmd.AssertExpectations(t)

	cmd.On("Request").Once().Return("reply")

	b := &telebot.Bot{CurrentCommand: cmd}
	msg := &tgbotapi.Message{
		Chat: &tgbotapi.Chat{ID: 1},
	}

	actualMsg := b.CommandRequest(msg)
	expectedMsg := tgbotapi.NewMessage(1, "reply")

	require.Equal(t, expectedMsg, actualMsg)
}

func TestQueryHandler(t *testing.T) {
	var (
		chatID int64 = 1
		query        = &tgbotapi.CallbackQuery{
			ID:   "callback_id",
			Data: "Yes",
			Message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: chatID},
				Text: "Question?",
			},
		}
	)

	tests := map[string]struct {
		setupMocks     func(api *mocks.MockTgAPI, cmd *mocks.MockCommand)
		currentCommand telebot.Command
		expectedMsg    tgbotapi.MessageConfig
	}{
		"successful callback": {
			setupMocks: func(api *mocks.MockTgAPI, cmd *mocks.MockCommand) {
				api.On("Request", mock.AnythingOfType("tgbotapi.CallbackConfig")).Once().
					Return(&tgbotapi.APIResponse{Ok: true}, nil)

				cmd.On("Validate", "Yes").Once().
					Return(nil)
				cmd.On("Done").Once().
					Return(true)
				cmd.On("Request").Once().
					Return("Request completed")
			},
			currentCommand: mocks.NewMockCommand(t),
			expectedMsg:    tgbotapi.NewMessage(1, "Request completed"),
		},
		"failed telegram api request": {
			setupMocks: func(api *mocks.MockTgAPI, _ *mocks.MockCommand) {
				api.On("Request", mock.AnythingOfType("tgbotapi.CallbackConfig")).Once().
					Return(&tgbotapi.APIResponse{Ok: true}, tgbotapi.Error{Message: "failed"})
			},
			expectedMsg: tgbotapi.NewMessage(1, "💥 Telegram API request failed!"),
		},
		"confirmation failed": {
			setupMocks: func(api *mocks.MockTgAPI, _ *mocks.MockCommand) {
				api.On("Request", mock.AnythingOfType("tgbotapi.CallbackConfig")).Once().
					Return(&tgbotapi.APIResponse{Ok: false}, nil)
			},
			expectedMsg: tgbotapi.NewMessage(1, "💥 Failed to confirm!"),
		},
		"no active command": {
			setupMocks: func(api *mocks.MockTgAPI, _ *mocks.MockCommand) {
				api.On("Request", mock.AnythingOfType("tgbotapi.CallbackConfig")).Once().
					Return(&tgbotapi.APIResponse{Ok: true}, nil)
			},
			currentCommand: nil,
			expectedMsg:    tgbotapi.NewMessage(1, "⚡️ Nothing to confirm!"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			api := mocks.NewMockTgAPI(t)
			defer api.AssertExpectations(t)

			var cmd *mocks.MockCommand
			if test.currentCommand != nil {
				var ok bool
				cmd, ok = test.currentCommand.(*mocks.MockCommand)
				require.True(t, ok)

				defer cmd.AssertExpectations(t)
			}

			test.setupMocks(api, cmd)

			b := &telebot.Bot{Tgb: api}
			if test.currentCommand != nil {
				b.CurrentCommand = cmd
			}

			actualMsg := b.QueryHandler(query)
			require.Equal(t, test.expectedMsg, actualMsg)
		})
	}
}

func TestHandleStart(t *testing.T) {
	var (
		chatID int64 = 1
		msg          = &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: chatID},
			Text: "/start",
		}
	)

	tests := map[string]struct {
		setupMock   func(client *mocks.MockScrapperClient)
		expectedMsg tgbotapi.MessageConfig
	}{
		"successful registration": {
			setupMock: func(client *mocks.MockScrapperClient) {
				client.On("PostTgChatId", mock.Anything, mock.AnythingOfType("int64")).
					Once().Return(&http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(nil)),
				}, nil)
			},
			expectedMsg: tgbotapi.NewMessage(chatID,
				"✨ You are successfully registered! Use /help to get a list of available commands."),
		},
		"failed registration": {
			setupMock: func(client *mocks.MockScrapperClient) {
				client.On("PostTgChatId", mock.Anything, mock.AnythingOfType("int64")).
					Once().Return(&http.Response{}, http.ErrHandlerTimeout)
			},
			expectedMsg: tgbotapi.NewMessage(chatID, "💥 Failed to register!"),
		},
		"repeated registration": {
			setupMock: func(client *mocks.MockScrapperClient) {
				client.On("PostTgChatId", mock.Anything, mock.AnythingOfType("int64")).
					Once().Return(&http.Response{
					StatusCode: http.StatusBadRequest,
					Body:       io.NopCloser(bytes.NewReader(nil)),
				}, nil)
			},
			expectedMsg: tgbotapi.NewMessage(chatID, "⚡️ You are already registered!"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			client := mocks.NewMockScrapperClient(t)
			defer client.AssertExpectations(t)

			test.setupMock(client)

			b := &telebot.Bot{
				Client: client,
			}

			actualMsg := b.HandleStart(msg)
			require.Equal(t, test.expectedMsg, actualMsg)
		})
	}
}

func TestHandleCancel(t *testing.T) {
	var (
		chatID int64 = 1
		msg          = &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: chatID},
			Text: "/cancel",
		}
	)

	tests := map[string]struct {
		currentCommand telebot.Command
		expectedMsg    tgbotapi.MessageConfig
	}{
		"successful cancellation": {
			currentCommand: mocks.NewMockCommand(t),
			expectedMsg:    tgbotapi.NewMessage(chatID, "↩️ Command cancelled!"),
		},
		"nothing to cancel": {
			currentCommand: nil,
			expectedMsg:    tgbotapi.NewMessage(chatID, "⚡️ Nothing to cancel!"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			b := &telebot.Bot{
				CurrentCommand: test.currentCommand,
			}

			actualMsg := b.HandleCancel(msg)
			require.Equal(t, test.expectedMsg, actualMsg)
		})
	}
}

func TestHandleState(t *testing.T) {
	var chatID int64 = 1

	tests := map[string]struct {
		msg            *tgbotapi.Message
		setupMock      func(cmd *mocks.MockCommand)
		currentCommand telebot.Command
		expectedMsg    tgbotapi.MessageConfig
	}{
		"no current command": {
			msg: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: chatID},
				Text: "/track",
				Entities: []tgbotapi.MessageEntity{
					{
						Type:   "bot_command",
						Offset: 0,
						Length: len("/track"),
					},
				},
			},
			setupMock:      func(_ *mocks.MockCommand) {},
			currentCommand: nil,
			expectedMsg: tgbotapi.NewMessage(chatID,
				"✨ Please, enter the Link you want to track! (press /cancel to quit)"),
		},
		"validation fails": {
			msg: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: chatID},
				Text: "invalid input",
			},
			setupMock: func(cmd *mocks.MockCommand) {
				cmd.On("Validate", "invalid input").
					Return(errors.New("validation error"))
				cmd.On("Stage").Once().
					Return("Please provide valid input", false)
			},
			currentCommand: mocks.NewMockCommand(t),
			expectedMsg:    tgbotapi.NewMessage(chatID, "Please provide valid input"),
		},
		"command done": {
			msg: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: chatID},
				Text: "valid input",
			},
			setupMock: func(cmd *mocks.MockCommand) {
				cmd.On("Validate", "valid input").
					Return(nil)
				cmd.On("Done").Once().
					Return(true)
				cmd.On("Request").Once().
					Return("Request completed successfully")
			},
			currentCommand: mocks.NewMockCommand(t),
			expectedMsg:    tgbotapi.NewMessage(chatID, "Request completed successfully"),
		},
		"command continues": {
			msg: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: chatID},
				Text: "valid input",
			},
			setupMock: func(cmd *mocks.MockCommand) {
				cmd.On("Validate", "valid input").
					Return(nil)
				cmd.On("Done").Once().
					Return(false)
				cmd.On("Stage").Once().
					Return("What's next?", false)
			},
			currentCommand: mocks.NewMockCommand(t),
			expectedMsg:    tgbotapi.NewMessage(chatID, "What's next?"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var cmd *mocks.MockCommand
			if test.currentCommand != nil {
				var ok bool
				cmd, ok = test.currentCommand.(*mocks.MockCommand)
				require.True(t, ok)

				defer cmd.AssertExpectations(t)
			}

			test.setupMock(cmd)

			b := &telebot.Bot{}
			if test.currentCommand != nil {
				b.CurrentCommand = cmd
			}

			actualMsg := b.HandleState(test.msg)
			require.Equal(t, test.expectedMsg, actualMsg)
		})
	}
}
