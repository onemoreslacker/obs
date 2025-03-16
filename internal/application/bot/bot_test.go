package bot_test

import (
	"log"
	"testing"

	"github.com/es-debug/backend-academy-2024-go-template/internal/application/bot"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/entities"
	"github.com/stretchr/testify/require"
)

// Так и не понял, как написать этот тест =( .
// func TestUnknownCommand(t *testing.T) {
// cfg, err := config.MustLoad()
// if err != nil {
// require.FailNow(t, err.Error())
// }

// mockBotAPI := &MockTeletramBotAPI{}

// mockBotAPI.On("Send", mock.AnythingOfType("tgbotapi.MessageConfig")).Return(tgbotapi.Message{}, nil)

// bt, err := bot.New(mockBotAPI, cfg)
// if err != nil {
// require.FailNow(t, err.Error())
// }

// err = bt.MessageHandler(123, "/start")
// require.NoError(t, err)

// mockBotAPI.AssertCalled(t, "Send", tgbotapi.MessageConfig{
// Text: "❌ Unknown command",
// })
// }

func TestListCommandConstruction(t *testing.T) {
	test := []entities.Link{
		entities.NewLink(1, "https://github.com/golang/go", []string{"tag"}, []string{"filter:value"}),
		entities.NewLink(2, "https://github.com/ziglang/zig", []string{"tag"}, []string{"filter:value"}),
		entities.NewLink(3, "https://github.com/rust-lang/rust", []string{"tag"}, []string{"filter:value"}),
	}

	actual := bot.ConstructListMessage(test)

	log.Println(actual)

	expected := `1. https://github.com/golang/go
2. https://github.com/ziglang/zig
3. https://github.com/rust-lang/rust
`

	require.Equal(t, expected, actual)
}
