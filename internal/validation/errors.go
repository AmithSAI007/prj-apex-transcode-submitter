package validation

import "errors"

var (
	ErrInvalidUUID            = errors.New("invalid UUID format")
	ErrPathInjection          = errors.New("path contains invalid characters")
	ErrInvalidGCSURI          = errors.New("invalid GCS URI format, expected gs://bucket/object")
	ErrUnexpectedBucket       = errors.New("unexpected bucket name in GCS URI")
	ErrPathMismatch           = errors.New("object path does not match expected userId/videoId format")
	ErrInvalidFileSize        = errors.New("file size is out of allowed range")
	ErrFileTooLarge           = errors.New("file size exceeds maximum allowed")
	ErrUnsupportedContentType = errors.New("content type is not allowed")
	ErrFieldTooLong           = errors.New("field value is too long")
	ErrSizeMismatch           = errors.New("file size does not match expected value")
)
