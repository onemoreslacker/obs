package botapi

import (
	"encoding/json"
	"fmt"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type API struct {
	tc UpdateSender
}

type UpdateSender interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
}

func New(client *tgbotapi.BotAPI) *API {
	return &API{
		tc: client,
	}
}

func (a *API) PostUpdates(w http.ResponseWriter, r *http.Request) {
	var params LinkUpdate
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), ErrBotUpdates.Error())
		return
	}

	if _, err := a.tc.Send(tgbotapi.NewMessage(
		params.TgChatId, fmt.Sprintf("%s: %s", params.Description, params.Url))); err != nil {
	}

	respondWithJSON(w, http.StatusOK, http.NoBody)
}
