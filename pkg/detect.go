package pkg

import (
	"errors"
	"net/url"
	"strings"
)

var ErrUnknownService = errors.New("unknown service")

func ServiceFromURL(link string) (string, error) {
	u, err := url.Parse(link)
	if err != nil {
		return "", err
	}

	switch {
	case strings.Contains(u.Host, "github"):
		return "github", nil

	case strings.Contains(u.Host, "stackoverflow"):
		return "stackoverflow", nil
	}

	return "", ErrUnknownService
}
