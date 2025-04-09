package scrapper

import (
	"context"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/entities"
	botclient "github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/clients/bot"
)

func (s *Scrapper) scrapeUpdates() error {
	chatIDs, err := s.repository.GetChatIDs()
	if err != nil {
		return err
	}

	updates := make(map[string][]int64)

	for _, chatID := range chatIDs {
		links, err := s.repository.GetLinks(chatID)
		if err != nil {
			return err
		}

		slog.Info(
			"starting updates collection",
			slog.Int("total_chats", len(chatIDs)),
			slog.String("service", "scrapper"),
		)

		if err := s.collectUpdates(links, updates, chatID); err != nil {
			return err
		}
	}

	for link, chats := range updates {
		update := botclient.PostUpdatesJSONRequestBody{
			Description: &Updates,
			TgChatIds:   &chats,
			Url:         &link,
		}
		if _, err := s.botClient.PostUpdates(context.Background(), update); err != nil {
			return err
		}
	}

	return nil
}

func (s *Scrapper) collectUpdates(links []entities.Link, updates map[string][]int64, chatID int64) error {
	for _, link := range links {
		l := *link.Url

		service, err := s.identifyService(l)
		if err != nil {
			return err
		}

		if _, exists := updates[l]; exists {
			updates[l] = append(updates[l], chatID)
		}

		switch service {
		case "github":
			updated, err := s.checkForUpdatesGithub(l)
			if err != nil {
				return err
			}

			if !updated {
				continue
			}
		case "stackoverflow":
			updated, err := s.checkForUpdatesStackOverflow(l)
			if err != nil {
				return err
			}

			if !updated {
				continue
			}
		}

		updates[l] = append(updates[l], chatID)

		slog.Info(
			"Found updates for link",
			"link", *link.Url,
			"service", service,
		)
	}

	return nil
}

func (s *Scrapper) checkForUpdatesStackOverflow(link string) (bool, error) {
	updates, err := s.external.RetrieveStackOverflowUpdates(link)
	if err != nil {
		return false, err
	}

	if len(updates) == 0 {
		return false, nil
	}

	createdAt := time.Unix(updates[0].CreatedAt, 0)

	return createdAt.After(getCutoff()), nil
}

func (s *Scrapper) checkForUpdatesGithub(link string) (bool, error) {
	updates, err := s.external.RetrieveGithubUpdates(link)
	if err != nil {
		return false, err
	}

	if len(updates) == 0 {
		return false, nil
	}

	createdAt, err := time.Parse(time.RFC3339, updates[0].CreatedAt)
	if err != nil {
		return false, err
	}

	return createdAt.After(getCutoff()), nil
}

func (s *Scrapper) identifyService(link string) (string, error) {
	u, err := url.Parse(link)
	if err != nil {
		return "", err
	}

	if strings.Contains(u.Host, "github") {
		return "github", nil
	}

	if strings.Contains(u.Host, "stackoverflow") {
		return "stackoverflow", nil
	}

	return "", ErrUnknownService
}

func getCutoff() time.Time {
	yesterday := time.Now().AddDate(0, 0, -1)
	cutoff := time.Date(
		yesterday.Year(),
		yesterday.Month(),
		yesterday.Day(),
		10,
		0,
		0,
		0,
		yesterday.Location(),
	)

	return cutoff
}

var (
	Updates = "🆕 New link activity!"
)
