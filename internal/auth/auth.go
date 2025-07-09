package auth

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"gemini-cli-go/internal/constants"
	"gemini-cli-go/internal/types"
)

// AuthManager handles OAuth2 authentication and Google Code Assist API communication
type AuthManager struct {
	config      *types.Environment
	accessToken string
	mu          sync.RWMutex
	cache       map[string]*types.CachedTokenData
	cacheMu     sync.RWMutex
}

// NewAuthManager creates a new AuthManager instance
func NewAuthManager(config *types.Environment) *AuthManager {
	return &AuthManager{
		config: config,
		cache:  make(map[string]*types.CachedTokenData),
	}
}

// InitializeAuth initializes authentication using OAuth2 credentials with caching
func (a *AuthManager) InitializeAuth() error {
	if a.config.GCPServiceAccount == "" {
		return fmt.Errorf(constants.ErrMissingGCPServiceAccount)
	}

	// First, try to get a cached token
	if cachedToken := a.getCachedToken(); cachedToken != nil {
		timeUntilExpiry := time.Until(cachedToken.ExpiryDate)
		if timeUntilExpiry > constants.TokenBufferTime {
			a.mu.Lock()
			a.accessToken = cachedToken.AccessToken
			a.mu.Unlock()
			return nil
		}
	}

	// Parse original credentials from environment
	var oauth2Creds types.OAuth2Credentials
	if err := json.Unmarshal([]byte(a.config.GCPServiceAccount), &oauth2Creds); err != nil {
		return fmt.Errorf("%s: %w", constants.ErrInvalidOAuth2Credentials, err)
	}

	// Check if the original token is still valid
	expiryTime := time.Unix(oauth2Creds.ExpiryDate/1000, 0)
	timeUntilExpiry := time.Until(expiryTime)

	if timeUntilExpiry > constants.TokenBufferTime {
		// Original token is still valid, cache it and use it
		a.mu.Lock()
		a.accessToken = oauth2Creds.AccessToken
		a.mu.Unlock()

		// Cache the token
		a.cacheToken(oauth2Creds.AccessToken, expiryTime)
		return nil
	}

	// Both original and cached tokens are expired, refresh the token
	return a.refreshAndCacheToken(oauth2Creds.RefreshToken)
}

// refreshAndCacheToken refreshes the OAuth token and caches it
func (a *AuthManager) refreshAndCacheToken(refreshToken string) error {
	data := url.Values{
		"client_id":     {constants.OAuthClientID},
		"client_secret": {constants.OAuthClientSecret},
		"refresh_token": {refreshToken},
		"grant_type":    {"refresh_token"},
	}
	// Create HTTP client with TLS configuration
	client := a.CreateHTTPClient()

	resp, err := client.PostForm(constants.OAuthRefreshURL, data)
	if err != nil {
		return fmt.Errorf("%s: %w", constants.ErrTokenRefreshFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: HTTP %d", constants.ErrTokenRefreshFailed, resp.StatusCode)
	}

	var refreshData types.TokenRefreshResponse
	if err := json.NewDecoder(resp.Body).Decode(&refreshData); err != nil {
		return fmt.Errorf("%s: %w", constants.ErrTokenRefreshFailed, err)
	}

	a.mu.Lock()
	a.accessToken = refreshData.AccessToken
	a.mu.Unlock()

	// Calculate expiry time (typically 1 hour from now)
	expiryTime := time.Now().Add(time.Duration(refreshData.ExpiresIn) * time.Second)

	// Cache the new token
	a.cacheToken(refreshData.AccessToken, expiryTime)

	return nil
}

// cacheToken caches the access token with expiry time
func (a *AuthManager) cacheToken(accessToken string, expiryTime time.Time) {
	a.cacheMu.Lock()
	defer a.cacheMu.Unlock()

	a.cache["oauth_token"] = &types.CachedTokenData{
		AccessToken: accessToken,
		ExpiryDate:  expiryTime,
		CachedAt:    time.Now(),
	}
}

// getCachedToken retrieves the cached token if it exists and is not expired
func (a *AuthManager) getCachedToken() *types.CachedTokenData {
	a.cacheMu.RLock()
	defer a.cacheMu.RUnlock()

	if token, exists := a.cache["oauth_token"]; exists {
		if time.Now().Before(token.ExpiryDate) {
			return token
		}
		// Token is expired, remove it from cache
		delete(a.cache, "oauth_token")
	}
	return nil
}

// ClearTokenCache clears the cached token
func (a *AuthManager) ClearTokenCache() {
	a.cacheMu.Lock()
	defer a.cacheMu.Unlock()
	delete(a.cache, "oauth_token")
}

// GetCachedTokenInfo returns information about the cached token
func (a *AuthManager) GetCachedTokenInfo() *types.TokenCacheInfo {
	a.cacheMu.RLock()
	defer a.cacheMu.RUnlock()

	if token, exists := a.cache["oauth_token"]; exists {
		timeUntilExpiry := time.Until(token.ExpiryDate)
		return &types.TokenCacheInfo{
			Cached:                 true,
			CachedAt:               token.CachedAt,
			ExpiresAt:              token.ExpiryDate,
			TimeUntilExpirySeconds: int(timeUntilExpiry.Seconds()),
			IsExpired:              timeUntilExpiry <= 0,
		}
	}

	return &types.TokenCacheInfo{
		Cached:  false,
		Message: "No token found in cache",
	}
}

// CallEndpoint makes a generic API call to a Code Assist endpoint
func (a *AuthManager) CallEndpoint(method string, body interface{}) (interface{}, error) {
	return a.callEndpointWithRetry(method, body, false)
}

// callEndpointWithRetry makes an API call with retry logic for 401 errors
func (a *AuthManager) callEndpointWithRetry(method string, body interface{}, isRetry bool) (interface{}, error) {
	if err := a.InitializeAuth(); err != nil {
		return nil, err
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	url := fmt.Sprintf("%s/%s:%s", constants.CodeAssistEndpoint, constants.CodeAssistAPIVersion, method)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	a.mu.RLock()
	token := a.accessToken
	a.mu.RUnlock()

	req.Header.Set("Content-Type", constants.ContentTypeJSON)
	req.Header.Set("Authorization", constants.BearerPrefix+token)

	client := a.CreateHTTPClient()

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized && !isRetry {
		// Clear token cache and retry once
		a.ClearTokenCache()
		a.mu.Lock()
		a.accessToken = ""
		a.mu.Unlock()

		if err := a.InitializeAuth(); err != nil {
			return nil, err
		}
		return a.callEndpointWithRetry(method, body, true)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API call failed with status %d", resp.StatusCode)
	}

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// GetAccessToken returns the current access token
func (a *AuthManager) GetAccessToken() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.accessToken
}

// IsAuthenticated checks if the user is authenticated
func (a *AuthManager) IsAuthenticated() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.accessToken != ""
}

// TestAuthentication tests if the authentication is working
func (a *AuthManager) TestAuthentication() error {
	if err := a.InitializeAuth(); err != nil {
		return err
	}

	// Try to make a simple API call to test authentication
	_, err := a.CallEndpoint("loadCodeAssist", map[string]interface{}{
		"cloudaicompanionProject": "test-project",
		"metadata": map[string]interface{}{
			"duetProject": "test-project",
		},
	})

	return err
}

// GetAuthenticationStatus returns detailed authentication status
func (a *AuthManager) GetAuthenticationStatus() map[string]interface{} {
	status := map[string]interface{}{
		"authenticated": a.IsAuthenticated(),
		"token_cache":   a.GetCachedTokenInfo(),
	}

	if a.IsAuthenticated() {
		status["message"] = constants.MsgAuthenticationOK
	} else {
		status["message"] = constants.ErrAuthenticationFailed
	}

	return status
}

// CreateHTTPClient creates an HTTP client with appropriate TLS configuration
func (a *AuthManager) CreateHTTPClient() *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{},
	}

	// Check for insecure TLS environment variable or configuration
	if a.shouldSkipTLSVerify() {
		transport.TLSClientConfig.InsecureSkipVerify = true
	}

	return &http.Client{
		Transport: transport,
		Timeout:   time.Duration(a.config.RequestTimeout) * time.Second,
	}
}

// shouldSkipTLSVerify determines if TLS verification should be skipped
func (a *AuthManager) shouldSkipTLSVerify() bool {
	// Check configuration first
	if a.config.SkipTLSVerify != "" {
		return strings.ToLower(a.config.SkipTLSVerify) == "true" || a.config.SkipTLSVerify == "1"
	}

	// Check environment variable as fallback
	if insecure := os.Getenv("SKIP_TLS_VERIFY"); insecure != "" {
		return strings.ToLower(insecure) == "true" || insecure == "1"
	}

	return false
}
