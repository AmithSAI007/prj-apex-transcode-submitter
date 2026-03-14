package validation

import "errors"

var (
	ErrInvalidContentType = errors.New("invalid content type in request header")
	ErrEmptyBody          = errors.New("request body is empty")
	ErrPayloadTooLarge    = errors.New("request body exceeds maximum allowed size")
	ErrMalformedJSON      = errors.New("malformed JSON in request body")
)
