// Package api provides functions to interact with the Emby API.
package api

import (
	"PiliPili_Frontend/config"
	"PiliPili_Frontend/logger"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// EmbyAPI provides methods to interact with the Emby API.
type EmbyAPI struct {
	EmbyURL string
	APIKey  string
	Client  *http.Client
}

// NewEmbyAPI initializes a new EmbyAPI instance.
func NewEmbyAPI() *EmbyAPI {
	cfg := config.GetConfig()
	return &EmbyAPI{
		EmbyURL: config.GetFullEmbyURL(),
		APIKey:  cfg.EmbyAPIKey,
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetMediaPath fetches the media file path from Emby using the provided item ID and MediaSourceID.
func (api *EmbyAPI) GetMediaPath(itemID, mediaSourceID string) (string, error) {
	url := fmt.Sprintf("%s/Items/%s/PlaybackInfo?MediaSourceId=%s&api_key=%s",
		api.EmbyURL, itemID, mediaSourceID, api.APIKey)

	logger.Info("Fetching media path from Emby: %s", url)

	resp, err := api.Client.Get(url)
	if err != nil {
		logger.Error("Failed to fetch media path: %v", err)
		return "", err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Error("Failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		logger.Error("Received non-200 response from Emby: %d", resp.StatusCode)
		return "", errors.New("failed to fetch media path")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Error reading response body: %v", err)
		return "", err
	}

	var result struct {
		MediaSources []struct {
			ID   string `json:"Id"`
			Path string `json:"Path"`
		} `json:"MediaSources"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		logger.Error("Error parsing JSON response: %v", err)
		return "", err
	}

	for _, source := range result.MediaSources {
		if source.ID == mediaSourceID {
			logger.Info("Found media path: %s", source.Path)
			return source.Path, nil
		}
	}

	logger.Warn("MediaSourceId not found in response")
	return "", errors.New("media source not found")
}
