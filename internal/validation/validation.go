package validation

import (
	"regexp"
	"slices"
	"strings"
)

var (
	dangerousPathPatterns = []string{
		"..",
		"/",
		"\\",
		"\x00",
		"%2f",
		"%2F",
		"%00",
		"%2e%2e",
	}
	uuidRegex = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)
)

func ValidateUUID(value string) bool {
	return uuidRegex.MatchString(value)
}

func ValidateNoPathInjection(name string) error {
	for _, pattern := range dangerousPathPatterns {
		if strings.Contains(name, pattern) {
			return ErrPathInjection
		}
	}
	return nil
}

func ValidateGCSURI(path string) (string, string, error) {

	// Check starts with "gs://"
	if !strings.HasPrefix(path, "gs://") {
		return "", "", ErrInvalidGCSURI
	}

	// Remove "gs://" prefix and split into bucket and object
	// sample path gs://my-bucket/userId/videoId/file.mp4
	parts := strings.SplitN(path[5:], "/", 2)
	if len(parts) != 2 {
		return "", "", ErrInvalidGCSURI
	}

	bucket := parts[0]
	object := parts[1]

	// Validate bucket name (basic check, can be enhanced)
	if bucket == "" || object == "" {
		return "", "", ErrInvalidGCSURI
	}

	return bucket, object, nil
}

func ValidateGCSBucket(bucket, expectedBucket string) error {
	if bucket != expectedBucket {
		return ErrUnexpectedBucket
	}
	return nil
}

func ValidatePathConsistency(objectPath, userID, videoID string) error {
	// Check objectPath contains userID and videoID in the expected format
	expectedPath := userID + "/" + videoID
	if !strings.HasPrefix(objectPath, expectedPath) {
		return ErrPathMismatch
	}
	return nil
}

func ValidateFileSize(size int64, minSize, maxSize int64) error {
	if size <= 0 {
		return ErrInvalidFileSize
	}
	if size > maxSize {
		return ErrFileTooLarge
	}
	return nil
}

func ValidateContentType(contentType string, allowedTypes []string) error {

	isValid := slices.Contains(allowedTypes, contentType)
	if isValid {
		return nil
	}
	return ErrUnsupportedContentType
}

func ValidateStringLength(field, value string, maxLength int) error {
	if len(value) > maxLength {
		return ErrFieldTooLong
	}
	return nil
}
