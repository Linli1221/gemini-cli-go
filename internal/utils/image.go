package utils

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"

	"gemini-cli-go/internal/constants"
)

// ImageValidation represents the result of image validation
type ImageValidation struct {
	IsValid  bool   `json:"is_valid"`
	MimeType string `json:"mime_type,omitempty"`
	Error    string `json:"error,omitempty"`
}

// ValidateImageURL validates an image URL and returns validation information
func ValidateImageURL(imageURL string) *ImageValidation {
	if imageURL == "" {
		return &ImageValidation{
			IsValid: false,
			Error:   "Image URL cannot be empty",
		}
	}

	// Check if it's a data URL (base64 encoded)
	if strings.HasPrefix(imageURL, "data:") {
		return validateDataURL(imageURL)
	}

	// Check if it's a regular URL
	if strings.HasPrefix(imageURL, "http://") || strings.HasPrefix(imageURL, "https://") {
		return validateHTTPURL(imageURL)
	}

	return &ImageValidation{
		IsValid: false,
		Error:   "Invalid image URL format. Must be a data URL or HTTP(S) URL",
	}
}

// validateDataURL validates a data URL
func validateDataURL(dataURL string) *ImageValidation {
	// Format: data:image/jpeg;base64,<base64-data>
	parts := strings.Split(dataURL, ",")
	if len(parts) != 2 {
		return &ImageValidation{
			IsValid: false,
			Error:   "Invalid data URL format",
		}
	}

	// Parse the media type
	mediaTypePart := parts[0]
	if !strings.HasPrefix(mediaTypePart, "data:") {
		return &ImageValidation{
			IsValid: false,
			Error:   "Invalid data URL prefix",
		}
	}

	// Extract MIME type
	mediaTypeInfo := strings.TrimPrefix(mediaTypePart, "data:")
	mimeTypeParts := strings.Split(mediaTypeInfo, ";")
	if len(mimeTypeParts) == 0 {
		return &ImageValidation{
			IsValid: false,
			Error:   "Missing MIME type in data URL",
		}
	}

	mimeType := mimeTypeParts[0]
	if !constants.SupportedImageMimeTypes[mimeType] {
		return &ImageValidation{
			IsValid: false,
			Error:   fmt.Sprintf("Unsupported image MIME type: %s", mimeType),
		}
	}

	// Validate base64 data
	base64Data := parts[1]
	if _, err := base64.StdEncoding.DecodeString(base64Data); err != nil {
		return &ImageValidation{
			IsValid: false,
			Error:   "Invalid base64 data",
		}
	}

	return &ImageValidation{
		IsValid:  true,
		MimeType: mimeType,
	}
}

// validateHTTPURL validates an HTTP(S) URL
func validateHTTPURL(httpURL string) *ImageValidation {
	parsedURL, err := url.Parse(httpURL)
	if err != nil {
		return &ImageValidation{
			IsValid: false,
			Error:   fmt.Sprintf("Invalid URL: %s", err.Error()),
		}
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return &ImageValidation{
			IsValid: false,
			Error:   "URL must use HTTP or HTTPS scheme",
		}
	}

	if parsedURL.Host == "" {
		return &ImageValidation{
			IsValid: false,
			Error:   "URL must have a valid host",
		}
	}

	// Try to infer MIME type from file extension
	path := parsedURL.Path
	mimeType := inferMimeTypeFromPath(path)

	return &ImageValidation{
		IsValid:  true,
		MimeType: mimeType,
	}
}

// inferMimeTypeFromPath infers MIME type from file path
func inferMimeTypeFromPath(path string) string {
	if path == "" {
		return "image/jpeg" // Default
	}

	// Get file extension
	lastDot := strings.LastIndex(path, ".")
	if lastDot == -1 {
		return "image/jpeg" // Default
	}

	ext := strings.ToLower(path[lastDot+1:])
	
	switch ext {
	case "jpg", "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	case "webp":
		return "image/webp"
	default:
		return "image/jpeg" // Default
	}
}

// IsValidImageMimeType checks if a MIME type is supported for images
func IsValidImageMimeType(mimeType string) bool {
	return constants.SupportedImageMimeTypes[mimeType]
}

// ExtractBase64Data extracts base64 data from a data URL
func ExtractBase64Data(dataURL string) (string, string, error) {
	if !strings.HasPrefix(dataURL, "data:") {
		return "", "", fmt.Errorf("not a data URL")
	}

	parts := strings.Split(dataURL, ",")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid data URL format")
	}

	// Parse the media type
	mediaTypePart := parts[0]
	mediaTypeInfo := strings.TrimPrefix(mediaTypePart, "data:")
	mimeTypeParts := strings.Split(mediaTypeInfo, ";")
	if len(mimeTypeParts) == 0 {
		return "", "", fmt.Errorf("missing MIME type")
	}

	mimeType := mimeTypeParts[0]
	base64Data := parts[1]

	// Validate base64 data
	if _, err := base64.StdEncoding.DecodeString(base64Data); err != nil {
		return "", "", fmt.Errorf("invalid base64 data: %w", err)
	}

	return mimeType, base64Data, nil
}

// CreateDataURL creates a data URL from MIME type and base64 data
func CreateDataURL(mimeType, base64Data string) string {
	return fmt.Sprintf("data:%s;base64,%s", mimeType, base64Data)
}

// GetImageSizeEstimate estimates image size in bytes from base64 data
func GetImageSizeEstimate(base64Data string) int {
	// Base64 encoding increases size by approximately 33%
	// So decoded size is roughly 3/4 of encoded size
	return (len(base64Data) * 3) / 4
}

// ValidateImageSize validates image size against limits
func ValidateImageSize(base64Data string, maxSizeBytes int) error {
	estimatedSize := GetImageSizeEstimate(base64Data)
	if estimatedSize > maxSizeBytes {
		return fmt.Errorf("image size (%d bytes) exceeds maximum allowed size (%d bytes)", estimatedSize, maxSizeBytes)
	}
	return nil
}