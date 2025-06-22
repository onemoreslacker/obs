package scrapperapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/es-debug/backend-academy-2024-go-template/config"
)

type Storage interface {
	AddChat(ctx context.Context, chatID int64) error
	ExistsChat(ctx context.Context, chatID int64) error
	DeleteChat(ctx context.Context, chatID int64) error
	AddLink(ctx context.Context, link AddLinkRequest, chatID int64) (int64, error)
	GetLinksWithChat(ctx context.Context, chatID int64) ([]LinkResponse, error)
	DeleteLink(ctx context.Context, link RemoveLinkRequest, chatID int64) error
}
type API struct {
	storage Storage
}

func New(storage Storage) *API {
	return &API{
		storage: storage,
	}
}

//nolint:revive,stylecheck // Generated code cannot be edited.
func (a *API) PostTgChatId(w http.ResponseWriter, r *http.Request, chatID int64) {
	ctx := r.Context()

	if err := a.storage.AddChat(ctx, chatID); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), ErrInvalidBody.Error())
		return
	}

	respondWithJSON(w, http.StatusNoContent, http.NoBody)
}

//nolint:revive,stylecheck // Generated code cannot be edited.
func (a *API) GetTgChatId(w http.ResponseWriter, r *http.Request, chatID int64) {
	ctx := r.Context()

	if err := a.storage.ExistsChat(ctx, chatID); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), ErrInvalidBody.Error())
		return
	}

	respondWithJSON(w, http.StatusNoContent, http.NoBody)
}

//nolint:revive,stylecheck // Generated code cannot be edited.
func (a *API) DeleteTgChatId(w http.ResponseWriter, r *http.Request, id int64) {
	ctx := r.Context()

	if err := a.storage.DeleteChat(ctx, id); err != nil {
		var status int

		if errors.Is(err, ErrChatNotExists) {
			status = http.StatusNotFound
		} else {
			status = http.StatusBadRequest
		}

		respondWithError(w, status, err.Error(), ErrInvalidBody.Error())
		return
	}

	respondWithJSON(w, http.StatusNoContent, http.NoBody)
}

func (a *API) PostLinks(w http.ResponseWriter, r *http.Request, params PostLinksParams) {
	ctx := r.Context()

	chatID := params.TgChatId

	var model AddLinkRequest

	if err := json.NewDecoder(r.Body).Decode(&model); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), ErrInvalidBody.Error())
		return
	}

	u, err := url.Parse(model.Link)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, ErrAddLinkInvalidLink.Error(), ErrInvalidBody.Error())
		return
	}

	u.Scheme = config.SchemeSecure

	if !isAvailable(u.String()) {
		respondWithError(w, http.StatusBadRequest, ErrAddLinkInvalidLink.Error(), ErrInvalidBody.Error())
		return
	}

	model.Link = u.String()

	linkID, err := a.storage.AddLink(ctx, model, chatID)
	if err != nil {
		var status int

		if errors.Is(err, ErrLinkAlreadyExists) {
			status = http.StatusConflict
		} else {
			status = http.StatusBadRequest
		}

		respondWithError(w, status, err.Error(),
			ErrAddLinkFailed.Error())

		return
	}

	respondWithJSON(w, http.StatusOK, LinkResponse{
		Id:      linkID,
		Url:     model.Link,
		Tags:    model.Tags,
		Filters: model.Filters,
	})
}

func (a *API) GetLinks(w http.ResponseWriter, r *http.Request, params GetLinksParams) {
	ctx := r.Context()

	chatID := params.TgChatId

	links, err := a.storage.GetLinksWithChat(ctx, chatID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, ErrGetLinksFailed.Error(), ErrInvalidBody.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, ListLinksResponse{
		Links: links,
		Size:  len(links),
	})
}

func (a *API) DeleteLinks(w http.ResponseWriter, r *http.Request, params DeleteLinksParams) {
	ctx := r.Context()

	chatID := params.TgChatId

	var model RemoveLinkRequest

	if err := json.NewDecoder(r.Body).Decode(&model); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), ErrInvalidBody.Error())
		return
	}

	u, err := url.Parse(model.Link)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, ErrDeleteLinkInvalidLink.Error(), ErrInvalidBody.Error())
		return
	}

	u.Scheme = config.SchemeSecure

	model.Link = u.String()

	if err := a.storage.DeleteLink(ctx, model, chatID); err != nil {
		var status int

		if errors.Is(err, ErrLinkAlreadyExists) {
			status = http.StatusConflict
		} else {
			status = http.StatusBadRequest
		}

		respondWithError(w, status, err.Error(), ErrAddLinkFailed.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, model.Link)
}
