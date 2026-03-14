package dto

// ErrorCode represents standardized machine-readable error codes returned in
// API error responses. Clients can switch on these codes to handle specific
// error conditions programmatically.
type ErrorCode string

const (
	ErrorCodeInvalidRequest     ErrorCode = "invalid_request"
	ErrorCodeInternal           ErrorCode = "internal"
	ErrorCodeInvalidContentType ErrorCode = "invalid_content_type"
	ErrorCodeEmptyBody          ErrorCode = "empty_body"
)

// ErrorDetail provides field-level error information, typically for validation failures.
type ErrorDetail struct {
	// Field is the request field that caused the error (omitted when not field-specific).
	Field string `json:"field,omitempty" example:"fileName"`
	// Message is a human-readable description of the error.
	Message string `json:"message" example:"fileName is required"`
}

// ErrorResponse is the top-level envelope for all API error responses.
type ErrorResponse struct {
	Error ErrorPayload `json:"error"`
}

// ErrorPayload carries the structured error information returned to API clients.
type ErrorPayload struct {
	// Code is the machine-readable error classification.
	Code ErrorCode `json:"code" example:"invalid_argument"`
	// Message is a human-readable summary of the error.
	Message string `json:"message" example:"Validation failed"`
	// RequestID is the correlation ID for tracing this request in logs and observability tools.
	RequestID string `json:"requestId,omitempty" example:"req_6f1a2c9d5e7b3a1c"`
	// Details contains field-level error information when applicable.
	Details []ErrorDetail `json:"details,omitempty"`
}
