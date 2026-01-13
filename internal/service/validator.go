package service

import (
	"fmt"
	"net/url"
)

func ValidLink(originalURL string) error {

	if len(originalURL) > 1000 || originalURL == "" {
		return ErrLinkBadRequest
	}
	_, err := url.ParseRequestURI(originalURL)

	if err != nil {
		return fmt.Errorf("Error checking the url: %w", err)
	}

	return nil
}
