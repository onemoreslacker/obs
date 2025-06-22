package svcident

import (
	"net/url"
	"strings"

	"github.com/es-debug/backend-academy-2024-go-template/config"
)

func FromLink(link string) (string, error) {
	u, err := url.Parse(link)
	if err != nil {
		return "", err
	}

	switch {
	case strings.Contains(u.Host, config.GitHub):
		return config.GitHub, nil
	case strings.Contains(u.Host, config.StackOverflow):
		return config.StackOverflow, nil
	}

	return "", ErrUnknownService
}
