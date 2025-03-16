package scrapperapi

import (
	"encoding/json"
	"errors"
	"math/rand/v2"
	"net/http"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/entities"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/repositories"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/storage"
)

type API struct {
	links repositories.LinksRepository
}

func New(links repositories.LinksRepository) *API {
	return &API{
		links: links,
	}
}

//nolint:revive,stylecheck // Generated code cannot be edited.
func (a *API) PostTgChatId(w http.ResponseWriter, _ *http.Request, id int64) {
	if err := a.links.AddChat(id); err != nil {
		respondWithError(w, http.StatusBadRequest,
			err.Error(), ErrInvalidBody.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, http.NoBody)
}

//nolint:revive,stylecheck // Generated code cannot be edited.
func (a *API) GetTgChatId(w http.ResponseWriter, _ *http.Request, id int64) {
	if _, err := a.links.GetLinks(id); err != nil {
		if errors.Is(err, storage.ErrChatNotFound) {
			respondWithError(w, http.StatusBadRequest,
				err.Error(), ErrInvalidBody.Error())
			return
		}
	}

	respondWithJSON(w, http.StatusOK, http.NoBody)
}

//nolint:revive,stylecheck // Generated code cannot be edited.
func (a *API) DeleteTgChatId(w http.ResponseWriter, _ *http.Request, id int64) {
	if err := a.links.DeleteChat(id); err != nil {
		respondWithError(w, http.StatusBadRequest,
			err.Error(), ErrInvalidBody.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, http.NoBody)
}

func (a *API) PostLinks(w http.ResponseWriter, r *http.Request, params PostLinksParams) {
	id := params.TgChatId

	var model AddLinkRequest

	if err := json.NewDecoder(r.Body).Decode(&model); err != nil {
		respondWithError(w, http.StatusBadRequest,
			err.Error(), ErrInvalidBody.Error())
		return
	}

	if model.Link == nil {
		respondWithError(w, http.StatusBadRequest,
			ErrAddLinkInvalidLink.Error(), ErrInvalidBody.Error())
		return
	}

	url, tags, filters := *model.Link, *model.Tags, *model.Filters

	req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	if err != nil {
		respondWithError(w, http.StatusBadRequest,
			ErrAddLinkInvalidLink.Error(), ErrInvalidBody.Error())
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		respondWithError(w, http.StatusBadRequest,
			ErrAddLinkInvalidLink.Error(), ErrInvalidBody.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		respondWithError(w, http.StatusBadRequest,
			ErrAddLinkInvalidLink.Error(), ErrInvalidBody.Error())
		return
	}

	link := entities.NewLink(rand.Int64(), url, tags, filters) //nolint:gosec // Temporary solution

	// NOTE: adjuct oapi config? (only 200 and 400
	// responses are expected currently)
	if err := a.links.AddLink(id, link); err != nil {
		respondWithError(w, http.StatusInternalServerError,
			ErrAddLinkFailed.msg, ErrAddLinkFailed.msg)
		return
	}

	respondWithJSON(w, http.StatusOK, LinkResponse(link))
}

func (a *API) GetLinks(w http.ResponseWriter, _ *http.Request, params GetLinksParams) {
	id := params.TgChatId

	links, err := a.links.GetLinks(id)
	if err != nil {
		respondWithError(w, http.StatusBadRequest,
			ErrGetLinksFailed.Error(), ErrInvalidBody.Error())
		return
	}

	linksResponse := make([]LinkResponse, len(links))
	for i := range len(links) {
		linksResponse[i] = LinkResponse(links[i])
	}

	sz := int32(len(linksResponse)) //nolint:gosec // Generated code cannot be edited.

	respondWithJSON(w, http.StatusOK, ListLinksResponse{
		Links: &linksResponse,
		Size:  &sz,
	})
}

func (a *API) DeleteLinks(w http.ResponseWriter, r *http.Request, params DeleteLinksParams) {
	id := params.TgChatId

	var model RemoveLinkRequest

	if err := json.NewDecoder(r.Body).Decode(&model); err != nil {
		respondWithError(w, http.StatusBadRequest,
			err.Error(), ErrInvalidBody.Error())
		return
	}

	if model.Link == nil {
		respondWithError(w, http.StatusBadRequest,
			ErrDeleteLinkInvalidLink.Error(), ErrInvalidBody.Error())
		return
	}

	link, err := a.links.DeleteLink(id, *model.Link)
	if err != nil {
		respondWithError(w, http.StatusBadRequest,
			ErrDeleteLinkFailed.Error(), ErrDeleteLinkFailed.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, LinkResponse(link))
}
