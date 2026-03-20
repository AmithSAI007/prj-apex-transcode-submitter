// Package constants defines typed context keys used across the application
// to store and retrieve request-scoped values such as user identity and
// trace correlation IDs. Using a dedicated type (CtxKey) prevents collisions
// with context keys from other packages.
package constants

// CtxKey is a private string type that prevents context-key collisions with
// other packages that might use plain string keys.
type CtxKey string

const (
	// CtxUserIDKey stores the authenticated user's ID in the request context.
	// Set by the auth middleware after successful JWT validation.
	CtxUserIDKey CtxKey = "user_id"

	// CtxTenantIDKey stores the tenant ID extracted from the request context.
	// Used for multi-tenant data isolation in Firestore queries.
	CtxTenantIDKey CtxKey = "tenant_id"

	// CtxTraceIDKey stores the application-level correlation ID (distinct from
	// OpenTelemetry trace IDs) for request logging and error responses.
	CtxTraceIDKey CtxKey = "trace_id"
)

// Instrumentation name used when creating OTel tracers. A single constant
// ensures every span produced by this service is grouped under the same
// instrumentation scope in the telemetry backend.
const InstrumentationName = "github.com/AmithSAI007/prj-apex-transcode-submitter"

// Standard log field keys used across all layers to ensure consistency
// in structured log output destined for BigQuery via GCP log sink.
const (
	LogKeyTraceID  = "trace_id"
	LogKeySpanID   = "span_id"
	LogKeyVideoID  = "video_id"
	LogKeyUserID   = "user_id"
	LogKeyBucket   = "bucket"
	LogKeyObject   = "object"
	LogKeyJobName  = "job_name"
	LogKeyStatus   = "status"
	LogKeyDuration = "duration_ms"
	LogKeyLayer    = "layer"
	LogKeyMethod   = "method"
)
