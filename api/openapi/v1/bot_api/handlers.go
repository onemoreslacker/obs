package botapi

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type API struct {
	tc *tgbotapi.BotAPI
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

	if params.Description == nil {
		respondWithError(w, http.StatusBadRequest,
			ErrMissingDescription.Error(), ErrBotUpdates.Error())
	}

	if params.Url == nil {
		respondWithError(w, http.StatusBadRequest,
			ErrMissingURL.Error(), ErrBotUpdates.Error())
		return
	}

	if params.TgChatIds == nil {
		respondWithError(w, http.StatusBadRequest,
			ErrMissingChatIDs.Error(), ErrBotUpdates.Error())
	}

	for _, chat := range *params.TgChatIds {
		if _, err := a.tc.Send(tgbotapi.NewMessage(
			chat, fmt.Sprintf("%s: %s", *params.Description, *params.Url))); err != nil {
			slog.Warn("Endpoint: POST /updates", "ChatID=", chat, "Description=", *params.Description)
		}
	}

	respondWithJSON(w, http.StatusOK, http.NoBody)
}
