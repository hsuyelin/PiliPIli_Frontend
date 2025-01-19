// Package util provides utility functions for URL handling.
package util

import (
	"fmt"
	"net/url"
	"strings"
)

// BuildFullURL constructs a complete URL with an optional port.
// It ensures the URL has a valid scheme, removes extra slashes, and appends the port if provided.
func BuildFullURL(baseURL string, port int) string {
	if baseURL == "" {
		return ""
	}

	// Parse the base URL
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		// If the baseURL is invalid, return it as-is (caller should handle validation errors)
		return baseURL
	}

	// Ensure the URL has a scheme (default to http)
	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "http"
	}

	// Ensure the URL does not end with an extra slash
	parsedURL.Path = strings.TrimRight(parsedURL.Path, "/")

	// Add the port if provided and not already included in the Host
	if port > 0 && parsedURL.Port() == "" {
		parsedURL.Host = fmt.Sprintf("%s:%d", parsedURL.Hostname(), port)
	}

	return parsedURL.String()
}
