// Package stream handles processing of media streams.
package stream

import (
	"PiliPili_Frontend/api"
	"PiliPili_Frontend/config"
	"PiliPili_Frontend/logger"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// cache instance for avoiding repeated processing.
var cache *Cache

func init() {
	var err error
	cache, err = NewCache(30 * time.Minute)
	if err != nil {
		logger.Error("Failed to initialize cache: %v", err)
	}
}

// HandleStreamRequest processes requests and redirects to a streaming URL.
func HandleStreamRequest(c *gin.Context) {
	logger.Info("Handling stream request...")
	logRequestDetails(c)

	itemID := c.Param("itemID")
	mediaSourceID := c.Query("MediaSourceId")

	if itemID == "" || mediaSourceID == "" {
		logger.Warn("Missing itemID or MediaSourceId")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing itemID or MediaSourceId"})
		return
	}

	logger.Debug("itemID: %s, MediaSourceId: %s", itemID, mediaSourceID)

	cacheKey := fmt.Sprintf("%s:%s", itemID, mediaSourceID)
	if cachedURL, found := cache.Get(cacheKey); found {
		logger.Info("Cache hit for key: %s", cacheKey)
		if validateSignature(cachedURL) {
			logger.Debug("Signature is valid. Redirecting to cached URL: %s", cachedURL)
			c.Header("Location", cachedURL)
			c.Status(http.StatusFound)
			return
		}
		logger.Warn("Signature expired or invalid. Regenerating URL.")
	}

	mediaPath, err := fetchMediaPath(itemID, mediaSourceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	streamingURL, err := generateStreamingURL(mediaPath, mediaSourceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := cache.Set(cacheKey, streamingURL); err != nil {
		logger.Error("Failed to set cache for key %s: %v", cacheKey, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cache streaming URL"})
		return
	}

	logger.Info("Redirecting to streaming URL: %s", streamingURL)
	c.Header("Location", streamingURL)
	c.Status(http.StatusFound)
}

// validateSignature checks if a signature is valid and not expired.
func validateSignature(cachedURL string) bool {
	signatureStart := "signature="
	index := strings.Index(cachedURL, signatureStart)
	if index == -1 {
		return false
	}

	signature := cachedURL[index+len(signatureStart):]
	signatureInstance, _ := GetSignatureInstance()
	decoded, err := signatureInstance.Decrypt(signature)
	if err != nil {
		logger.Warn("Failed to decrypt signature: %v", err)
		return false
	}

	expireAt, ok := decoded["expireAt"].(float64)
	if !ok || int64(expireAt) <= time.Now().Unix() {
		logger.Warn("Signature expired")
		return false
	}

	return true
}

// fetchMediaPath retrieves the media path from Emby.
func fetchMediaPath(itemID, mediaSourceID string) (string, error) {
	embyAPI := api.NewEmbyAPI()
	mediaPath, err := embyAPI.GetMediaPath(itemID, mediaSourceID)
	if err != nil {
		logger.Error("Failed to fetch media path for itemID: %s, MediaSourceId: %s. Error: %v", itemID, mediaSourceID, err)
		return "", fmt.Errorf("failed to fetch media path")
	}
	logger.Debug("Fetched original media path: %s", mediaPath)

	backendStorageBasePath := config.GetConfig().BackendStorageBasePath
	if backendStorageBasePath != "" && strings.HasPrefix(mediaPath, backendStorageBasePath) {
		mediaPath = strings.TrimPrefix(mediaPath, backendStorageBasePath)
		mediaPath = strings.TrimPrefix(mediaPath, "/")
	}

	logger.Debug("Processed media path: %s", mediaPath)
	return mediaPath, nil
}

// generateStreamingURL creates a signed streaming URL.
func generateStreamingURL(mediaPath, mediaSourceID string) (string, error) {
	cfg := config.GetConfig()
	signatureInstance, _ := GetSignatureInstance()
	expireAt := time.Now().Unix() + int64(cfg.PlayURLMaxAliveTime)
	signature, err := signatureInstance.Encrypt(mediaSourceID, expireAt)
	logger.Debug(
		"Generated signature: mediaSourceID %s, expireAt %s, signature %s, mediaPath: %s",
		mediaSourceID,
		expireAt,
		signature,
		mediaPath,
	)

	if err != nil {
		logger.Error("Failed to generate signed URL for MediaSourceId: %s. Error: %v", mediaSourceID, err)
		return "", fmt.Errorf("failed to generate signed URL")
	}
	streamingURL := fmt.Sprintf(
		"%s?path=%s&signature=%s",
		config.GetFullBackendURL(),
		url.QueryEscape(mediaPath),
		signature,
	)
	logger.Debug("Generated streaming URL: %s", streamingURL)
	return streamingURL, nil
}

// logRequestDetails logs the details of the incoming request, including headers and body.
func logRequestDetails(c *gin.Context) {
	logger.Info("Request Headers: %v", c.Request.Header)
	if c.Request.Body != nil {
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err == nil {
			c.Request.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))
			logger.Debug("Request Body: %s", string(bodyBytes))
		} else {
			logger.Warn("Failed to read request body: %v", err)
		}
	}
}
