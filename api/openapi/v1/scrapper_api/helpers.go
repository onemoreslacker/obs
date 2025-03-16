package scrapperapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.Info(err.Error())
	}
}

func respondWithError(w http.ResponseWriter, code int, msg, description string) {
	err := ApiErrorResponse{
		Description:  &description,
		Code:         &code,
		ErrorMessage: &msg,
	}

	respondWithJSON(w, code, err)
}
