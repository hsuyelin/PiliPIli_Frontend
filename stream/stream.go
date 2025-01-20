// Package stream handles processing of media streams.
package stream

import (
	"PiliPili_Frontend/api"
	"PiliPili_Frontend/config"
	"PiliPili_Frontend/logger"
	"PiliPili_Frontend/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// Cache instance for avoiding repeated processing.
var cache *Cache
var globalTimeChecker util.TimeChecker

// init initializes global variables such as cache and time checker.
func init() {
	var err error
	cache, err = NewCache(30 * time.Minute)
	if err != nil {
		logger.Error("Failed to initialize cache: %v", err)
		os.Exit(1)
	}

	globalTimeChecker = util.TimeChecker{}
	logger.Info("TimeChecker initialized successfully")
}

// HandleStreamRequest processes client requests and redirects them to a generated streaming URL.
func HandleStreamRequest(c *gin.Context) {
	logger.Info("Handling stream request...")
	logRequestDetails(c)

	// Fetch necessary parameters for processing the request.
	itemID, mediaSourceID, mediaPath, isSpecialDate := fetchParameters(c)
	if itemID == "" || mediaSourceID == "" {
		return // Early exit if parameters are missing.
	}

	// Handle cache: Check if a valid streaming URL exists in the cache.
	if _, found := handleCache(c, itemID, mediaSourceID); found {
		return
	}

	// Fetch media path if it is not a special date.
	var err error
	mediaPath, err = fetchMediaPathIfNeeded(itemID, mediaSourceID, mediaPath, isSpecialDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Generate and cache the streaming URL.
	streamingURL, err := generateAndCacheURL(itemID, mediaSourceID, mediaPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Redirect the client to the generated streaming URL.
	logger.Info("Redirecting to streaming URL: %s", streamingURL)
	c.Header("Location", streamingURL)
	c.Status(http.StatusFound)
}

// fetchParameters retrieves parameters from the request or special date configuration.
func fetchParameters(c *gin.Context) (string, string, string, bool) {
	currentTime := time.Now()

	// Check for special date configuration.
	specialConfig := getMediaForSpecialDate(currentTime)
	if specialConfig.IsValid() {
		logger.Info("Special date detected. Using special configuration.")
		return specialConfig.ItemId, specialConfig.MediaSourceID, specialConfig.MediaPath, true
	}

	// Retrieve parameters from the request.
	itemID := c.Param("itemID")
	mediaSourceID := c.Query("MediaSourceId")
	if itemID == "" || mediaSourceID == "" {
		logger.Warn("Missing itemID or MediaSourceId")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing itemID or MediaSourceId"})
		return "", "", "", false
	}

	return itemID, mediaSourceID, "", false
}

// getMediaForSpecialDate returns the special media configuration for the current date.
func getMediaForSpecialDate(t time.Time) config.SpecialMediaConfig {
	specialMedias := config.GetConfig().SpecialMedias

	// Iterate through special media configurations and match with the current date.
	for _, media := range specialMedias {
		switch media.Key {
		case "ChineseNewYearEve":
			if globalTimeChecker.IsChineseNewYearEve(t) {
				return media
			}
		case "October1":
			if globalTimeChecker.IsOctober1Morning(t) {
				return media
			}
		case "December13":
			if globalTimeChecker.IsDecember13Morning(t) {
				return media
			}
		case "September18":
			if globalTimeChecker.IsSeptember18Morning(t) {
				return media
			}
		}
	}

	return config.SpecialMediaConfig{}
}

// getMediaForMissingMedia returns the default media configuration for missing cases.
func getMediaForMissingMedia() config.SpecialMediaConfig {
	specialMedias := config.GetConfig().SpecialMedias

	// Find the configuration with the key "MediaMissing".
	for _, media := range specialMedias {
		if media.Key == "MediaMissing" {
			return media
		}
	}

	return config.SpecialMediaConfig{}
}

// handleCache checks the cache for an existing streaming URL.
func handleCache(c *gin.Context, itemID, mediaSourceID string) (string, bool) {
	cacheKey := fmt.Sprintf("%s:%s", itemID, mediaSourceID)
	if cachedURL, found := cache.Get(cacheKey); found {
		logger.Info("Cache hit for key: %s", cacheKey)
		if validateSignature(cachedURL) {
			logger.Debug("Signature is valid. Redirecting to cached URL: %s", cachedURL)
			c.Header("Location", cachedURL)
			c.Status(http.StatusFound)
			return cachedURL, true
		}
		logger.Warn("Signature expired or invalid. Regenerating URL.")
	}
	return "", false
}

// fetchMediaPathIfNeeded fetches the media path if the date is not a special date.
func fetchMediaPathIfNeeded(itemID, mediaSourceID, mediaPath string, isSpecialDate bool) (string, error) {
	if !isSpecialDate {
		var err error
		mediaPath, err = fetchMediaPath(itemID, mediaSourceID)
		if err != nil {
			missingMediaConfig := getMediaForMissingMedia()
			itemID = missingMediaConfig.ItemId
			mediaSourceID = missingMediaConfig.MediaSourceID
			if itemID == "" || mediaSourceID == "" {
				logger.Error("Missing itemID or MediaSourceId")
				return "", err
			} else {
				mediaPath = missingMediaConfig.MediaPath
			}
		}
	}
	return mediaPath, nil
}

// generateAndCacheURL generates a streaming URL and caches it.
func generateAndCacheURL(itemID, mediaSourceID, mediaPath string) (string, error) {
	streamingURL, err := generateStreamingURL(mediaPath, itemID, mediaSourceID)
	if err != nil {
		return "", err
	}

	cacheKey := fmt.Sprintf("%s:%s", itemID, mediaSourceID)
	if err := cache.Set(cacheKey, streamingURL); err != nil {
		logger.Error("Failed to set cache for key %s: %v", cacheKey, err)
		return "", err
	}

	return streamingURL, nil
}

// validateSignature checks if a cached URL's signature is valid and not expired.
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

// fetchMediaPath retrieves the media path from the Emby server.
func fetchMediaPath(itemID, mediaSourceID string) (string, error) {
	embyAPI := api.NewEmbyAPI()
	mediaPath, err := embyAPI.GetMediaPath(itemID, mediaSourceID)
	if err != nil {
		logger.Error(
			"Failed to fetch media path for itemID: %s, MediaSourceId: %s. Error: %v",
			itemID,
			mediaSourceID,
			err,
		)
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

// generateStreamingURL creates a signed streaming URL with a signature.
func generateStreamingURL(mediaPath, itemID, mediaSourceID string) (string, error) {
	cfg := config.GetConfig()
	signatureInstance, _ := GetSignatureInstance()
	expireAt := time.Now().Unix() + int64(cfg.PlayURLMaxAliveTime)
	signature, err := signatureInstance.Encrypt(itemID, mediaSourceID, expireAt)
	logger.Debug(
		"Generated signature: itemID: %s, mediaSourceID %s, expireAt %d, signature %s, mediaPath: %s",
		itemID,
		mediaSourceID,
		expireAt,
		signature,
		mediaPath,
	)

	if err != nil {
		logger.Error(
			"Failed to generate signed URL for itemID: %s, MediaSourceId: %s. Error: %v",
			itemID,
			mediaSourceID,
			err,
		)
		return "", fmt.Errorf("failed to generate signed URL")
	}
	streamingURL := fmt.Sprintf(
		"%s?path=%s&signature=%s",
		config.GetFullBackendURL(),
		url.QueryEscape(mediaPath),
		signature,
	)
	logger.Info("Generated streaming URL: %s", streamingURL)
	return streamingURL, nil
}

// logRequestDetails logs request headers and body for debugging purposes.
func logRequestDetails(c *gin.Context) {
	logger.Debug("Request Headers: %v", c.Request.Header)
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
