package updater

import (
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/pkg"
)

func (upd *Updater) checkActivity(url string) (bool, error) {
	var (
		updated bool
		err     error
	)

	service, err := pkg.ServiceFromURL(url)
	if err != nil {
		return updated, err
	}

	switch service {
	case "github":
		updated, err = upd.checkForUpdatesGithub(url)

	case "stackoverflow":
		updated, err = upd.checkForUpdatesStackOverflow(url)
	}

	return updated, err
}

func (upd *Updater) checkForUpdatesStackOverflow(link string) (bool, error) {
	updates, err := upd.external.RetrieveStackOverflowUpdates(link)
	if err != nil {
		return false, err
	}

	if len(updates) == 0 {
		return false, nil
	}

	createdAt := time.Unix(updates[0].CreatedAt, 0)

	return createdAt.After(getCutoff()), nil
}

func (upd *Updater) checkForUpdatesGithub(link string) (bool, error) {
	updates, err := upd.external.RetrieveGitHubUpdates(link)
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
