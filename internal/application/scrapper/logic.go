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
	chatIDs := s.repository.GetChatIDs()
	updates := make(map[string][]int64)

	for _, chatID := range chatIDs {
		links, err := s.repository.GetLinks(chatID)
		if err != nil {
			return err
		}

		slog.Info(
			"Starting updates collection",
			"total_chats", len(chatIDs),
		)

		if err := s.collectUpdates(links, updates, chatID); err != nil {
			return err
		}
	}

	for link, chats := range updates {
		update := botclient.PostUpdatesJSONRequestBody{
			Description: &s.cfg.Meta.Updates.Plug,
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
		case s.cfg.Meta.Services.GitHub:
			updated, err := s.checkForUpdatesGithub(l)
			if err != nil {
				return err
			}

			if !updated {
				continue
			}
		case s.cfg.Meta.Services.StackOverflow:
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
	resp, err := s.external.GetStackOverflowAnswers(link)
	if err != nil {
		return false, err
	}

	lastActivityDate := time.Unix(resp.Items[0].LastActivityDate, 0)

	return lastActivityDate.After(getCutoff()), nil
}

func (s *Scrapper) checkForUpdatesGithub(link string) (bool, error) {
	resp, err := s.external.GetGitHubRepository(link)
	if err != nil {
		return false, err
	}

	updatedAt := resp.LastUpdated

	return updatedAt.After(getCutoff()), nil
}

func (s *Scrapper) identifyService(link string) (string, error) {
	u, err := url.Parse(link)
	if err != nil {
		return "", err
	}

	if strings.Contains(u.Host, s.cfg.Meta.Services.GitHub) {
		return s.cfg.Meta.Services.GitHub, nil
	}

	if strings.Contains(u.Host, s.cfg.Meta.Services.StackOverflow) {
		return s.cfg.Meta.Services.StackOverflow, nil
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
