package updater

import (
	"context"
	"fmt"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/pkg"
)

func (upd *Updater) checkActivity(ctx context.Context, url string) (bool, error) {
	var (
		updated bool
		err     error
	)

	service, err := pkg.ServiceFromURL(url)
	if err != nil {
		return updated, err
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	switch service {
	case "github":
		updated, err = upd.checkForUpdatesGithub(ctx, url)
	case "stackoverflow":
		updated, err = upd.checkForUpdatesStackOverflow(ctx, url)
	default:
		return false, fmt.Errorf("failed to check for updates: %w", ErrUnknownService)
	}

	return updated, err
}

func (upd *Updater) checkForUpdatesGithub(ctx context.Context, link string) (bool, error) {
	updates, err := upd.external.RetrieveGitHubUpdates(ctx, link)
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

func (upd *Updater) checkForUpdatesStackOverflow(ctx context.Context, link string) (bool, error) {
	updates, err := upd.external.RetrieveStackOverflowUpdates(ctx, link)
	if err != nil {
		return false, err
	}

	if len(updates) == 0 {
		return false, nil
	}

	createdAt := time.Unix(updates[0].CreatedAt, 0)

	return createdAt.After(getCutoff()), nil
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
