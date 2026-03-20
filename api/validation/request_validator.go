package validation

import "errors"

const (
	MaxBodySize                 = 10 * 1024 // 10KB - adjust as needed for your application
	applicationJSON             = "application/json"
	XCloudTasksTaskNameHeader   = "X-CloudTasks-TaskName"
	XCloudTasksQueueNameHeader  = "X-CloudTasks-QueueName"
	XCloudTasksRetryCountHeader = "X-CloudTasks-RetryCount"
)

var (
	ErrInvalidContentType         = errors.New("invalid content type, expected application/json")
	ErrEmptyBody                  = errors.New("request body is empty")
	ErrPayloadTooLarge            = errors.New("payload too large, exceeds maximum allowed size")
	ErrMissingCloudTaskHeader     = errors.New("missing required Cloud Tasks header")
	ErrInvalidCloudTasksQueueName = errors.New("invalid Cloud Tasks queue name")
)

func ValidateContentType(contentType string) error {
	if contentType != applicationJSON {
		return ErrInvalidContentType
	}
	return nil
}

func ValidateContentLength(contentLength int64) error {
	if contentLength == 0 {
		return ErrEmptyBody
	}
	if contentLength > MaxBodySize {
		return ErrPayloadTooLarge
	}
	return nil
}

func ValidateCloudTaskHeaders(headers map[string]string, queue string) error {
	if headers[XCloudTasksTaskNameHeader] == "" {
		return ErrMissingCloudTaskHeader
	}

	if headers[XCloudTasksQueueNameHeader] == "" || headers[XCloudTasksQueueNameHeader] != queue {
		return ErrInvalidCloudTasksQueueName
	}

	return nil
}
