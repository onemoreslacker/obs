package scrapperapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if code == http.StatusNoContent {
		return
	}

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.Info(err.Error())
	}
}

func respondWithError(w http.ResponseWriter, code int, msg, description string) {
	err := ApiErrorResponse{
		Description:  description,
		Code:         code,
		ErrorMessage: msg,
	}

	respondWithJSON(w, code, err)
}

func isAvailable(url string) bool {
	req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	if err != nil {
		return false
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}
