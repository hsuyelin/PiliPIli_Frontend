// Package stream handles processing of media streams.
package stream

import (
	"PiliPili_Frontend/api"
	"PiliPili_Frontend/config"
	"PiliPili_Frontend/logger"
	"PiliPili_Frontend/util"
	"bytes"
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

type RequestParameters struct {
	EmbyApiKey    string // The API key for authenticating with the Emby server.
	ItemId        string // The unique identifier of the media item.
	MediaSourceID string // The identifier of the specific media source.
	MediaPath     string // The file path to the media.
	IsSpecialDate bool   // A flag indicating whether the request is for a special date or occasion.
}

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
	requestParameters := fetchRequestParameters(c)

	if requestParameters.EmbyApiKey == "" ||
		requestParameters.ItemId == "" ||
		requestParameters.MediaSourceID == "" {

		return // Early exit if parameters are missing.
	}

	// Handle cache: Check if a valid streaming URL exists in the cache.
	if _, found := handleCache(c, requestParameters); found {
		return
	}

	// Fetch media path if it is not a special date.
	var err error
	var mediaPath string
	mediaPath, err = fetchMediaPathIfNeeded(requestParameters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Generate and cache the streaming URL.
	streamingURL, err := generateAndCacheURL(mediaPath, requestParameters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Redirect the client to the generated streaming URL.
	logger.Info("Redirecting to streaming URL: %s", streamingURL)
	c.Header("Location", streamingURL)
	c.Status(http.StatusFound)
}

// fetchRequestParameters retrieves parameters from the request or special date configuration.
func fetchRequestParameters(c *gin.Context) RequestParameters {
	currentTime := time.Now()

	apiKey := c.Query("api_key")
	if apiKey == "" {
		apiKey = config.GetConfig().EmbyAPIKey
	}

	if apiKey == "" {
		logger.Error("Missing emby api key")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing emby api key"})
		return RequestParameters{}
	}

	logger.Debug("Emby api key: %s", apiKey)

	// Check for special date configuration.
	specialConfig := getMediaForSpecialDate(currentTime)
	if specialConfig.IsValid() {
		logger.Info("Special date detected. Using special configuration.")
		return RequestParameters{
			apiKey,
			specialConfig.ItemId,
			specialConfig.MediaSourceID,
			specialConfig.MediaPath,
			true,
		}
	}

	// Retrieve parameters from the request.
	itemID := c.Param("itemID")
	mediaSourceID := c.Query("MediaSourceId")
	logger.Debug("ItemID: %s, mediaSourceID: %s", itemID, mediaSourceID)
	if itemID == "" || mediaSourceID == "" {
		logger.Error("Missing itemID or MediaSourceId")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing itemID or MediaSourceId"})
		return RequestParameters{
			apiKey,
			"",
			"",
			"",
			false,
		}
	}

	return RequestParameters{
		apiKey,
		itemID,
		mediaSourceID,
		"",
		false,
	}
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
func handleCache(c *gin.Context, parameters RequestParameters) (string, bool) {
	cacheKey := fmt.Sprintf("%s:%s", parameters.ItemId, parameters.MediaSourceID)
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
func fetchMediaPathIfNeeded(parameters RequestParameters) (string, error) {
	itemID := parameters.ItemId
	mediaSourceID := parameters.MediaSourceID
	mediaPath := parameters.MediaPath

	if !parameters.IsSpecialDate {
		var err error
		mediaPath, err = fetchMediaPath(parameters)
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
func generateAndCacheURL(mediaPath string, parameters RequestParameters) (string, error) {
	itemID := parameters.ItemId
	mediaSourceID := parameters.MediaSourceID

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
func fetchMediaPath(parameters RequestParameters) (string, error) {
	embyAPI := api.NewEmbyAPI()
	mediaPath, err := embyAPI.GetMediaPath(
		parameters.EmbyApiKey,
		parameters.ItemId,
		parameters.MediaSourceID,
	)
	if err != nil {
		logger.Error(
			"Failed to fetch media path for itemID: %s, MediaSourceId: %s. Error: %v",
			parameters.ItemId,
			parameters.MediaSourceID,
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
	// Construct the full URL using a more efficient approach
	protocol := "http"
	if c.Request.TLS != nil {
		protocol = "https"
	}
	fullURL := protocol + "://" + c.Request.Host + c.Request.RequestURI
	logger.Debug("Request URL: %s", fullURL)

	// Log request headers directly (no need to format if the logger handles objects efficiently)
	logger.Debug("Request Headers:", c.Request.Header)

	// Log request body (if available)
	if c.Request.Body != nil {
		// Use a limited buffer to avoid unnecessary allocations for very large bodies
		bodyBytes, err := io.ReadAll(io.LimitReader(c.Request.Body, 1024*1024)) // Limit to 1 MB
		if err == nil {
			// Reset the body to allow further handlers to read it
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			logger.Debug("Request Body:", string(bodyBytes)) // Let logger handle the string formatting
		} else {
			logger.Warn("Failed to read request body:", err)
		}
	}
}
