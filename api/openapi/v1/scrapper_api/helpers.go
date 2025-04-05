package scrapperapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
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

func checkResourceAvailability(link string) bool {
	u, err := url.Parse(link)
	if err != nil {
		return false
	}

	if u.Scheme == "" {
		u.Scheme = "https"
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), http.NoBody)
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
