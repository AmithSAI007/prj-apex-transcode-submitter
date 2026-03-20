// Package config provides application configuration loading, logger setup,
// and OpenTelemetry tracer initialization for the Apex Upload Platform.
package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all application configuration values loaded from config.yaml
// and/or environment variables. Environment variables take precedence over
// file-based values when both are present.
type Config struct {
	// AppEnv is the deployment environment ("development", "production", etc.).
	AppEnv string `mapstructure:"APP_ENV"`
	// ProjectRegion is the GCP region where resources are located (e.g., "us-central1").
	ProjectRegion string `mapstructure:"GCP_PROJECT_REGION"`
	// HttpPort is the TCP port the HTTP server listens on (default "8080").
	HttpPort string `mapstructure:"HTTP_PORT"`
	// GCPProjectID is the Google Cloud project ID used for GCS, Firestore, and Secret Manager.
	GCPProjectID string `mapstructure:"GCP_PROJECT_ID"`
	// FirestoreDatabaseID is the named Firestore database to connect to.
	FirestoreDatabaseID string `mapstructure:"FIRESTORE_DATABASE_ID"`
	// GCSBucket is the GCS bucket where uploaded objects are stored.
	GCSBucket string `mapstructure:"GCS_BUCKET"`
	// OutputGCSBucket is the GCS bucket where transcoded outputs will be stored.
	OutputGCSBucket string `mapstructure:"OUTPUT_GCS_BUCKET"`
	// OtelServiceName is the name of the service for OpenTelemetry tracing.
	OtelServiceName string `mapstructure:"OTEL_SERVICE_NAME"`
	// OtelExporterOtlpHeaders are the headers to include when exporting traces to an OTLP endpoint.
	OtelExporterOtlpHeaders string `mapstructure:"OTEL_EXPORTER_OTLP_HEADERS"`
	// MaxFileSizeBytes is the maximum allowed file size for uploads, in bytes.
	MaxFileSizeBytes int64 `mapstructure:"MAX_FILE_SIZE_BYTES"`
	//MinFileSizeBytes is the minimum allowed file size for uploads, in bytes.
	MinFileSizeBytes int64 `mapstructure:"MIN_FILE_SIZE_BYTES"`
	// AllowedContentTypes is a comma-separated list of allowed MIME types for uploaded files.
	AllowedContentTypes string `mapstructure:"ALLOWED_CONTENT_TYPES"`
	// TranscodeTaskQueue is the name of the Cloud Tasks queue to which transcode tasks will be submitted.
	TranscodeTaskQueue string `mapstructure:"TRANSCODE_TASK_QUEUE"`
	// FirestoreCollection is the Firestore collection used to store video document state.
	FirestoreCollection string `mapstructure:"FIRESTORE_COLLECTION"`
	// TranscoderTemplateID is the Transcoder API job template to use for transcoding jobs.
	TranscoderTemplateID string `mapstructure:"TRANSCODER_TEMPLATE_ID"`
	// OtelResourceAttributes is a comma-separated list of key=value pairs to set as resource attributes on all traces.
	OtelResourceAttributes string `mapstructure:"OTEL_RESOURCE_ATTRIBUTES"`
}

// LoadConfig reads configuration from a YAML file at the given path and merges
// it with environment variables. Sensible defaults are provided for local
// development. Returns an error if the config file exists but cannot be parsed,
// or if the values cannot be unmarshalled into the Config struct.
func LoadConfig(path string) (*Config, error) {

	// Set sensible defaults for local development.
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("GCP_PROJECT_REGION", "us-central1")
	viper.SetDefault("HTTP_PORT", ":8080")
	viper.SetDefault("GCP_PROJECT_ID", "amith-testing")
	viper.SetDefault("FIRESTORE_DATABASE_ID", "apex-firestore-db")
	viper.SetDefault("GCS_BUCKET", "")
	viper.SetDefault("OUTPUT_GCS_BUCKET", "")
	viper.SetDefault("OTEL_SERVICE_NAME", "prj-apex-upload-platform")
	viper.SetDefault("OTEL_EXPORTER_OTLP_HEADERS", "x-goog-user-project=amith-testing")
	viper.SetDefault("MAX_FILE_SIZE_BYTES", 10*1024*1024*1024) // 10 GB
	viper.SetDefault("MIN_FILE_SIZE_BYTES", 1)                 // 1 byte
	viper.SetDefault("ALLOWED_CONTENT_TYPES", "video/mp4,video/mkv,video/avi")
	viper.SetDefault("TRANSCODE_TASK_QUEUE", "transcode-tasks")
	viper.SetDefault("FIRESTORE_COLLECTION", "videos")
	viper.SetDefault("TRANSCODER_TEMPLATE_ID", "preset/web-hd")

	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Bind environment variables; dots in keys become underscores (e.g., GCS.BUCKET -> GCS_BUCKET).
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read the config file; ignore "file not found" since env vars may supply all values.
	err := viper.ReadInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); err != nil && !ok {
		return nil, fmt.Errorf("fatal error config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode into struct, %w", err)
	}

	return &config, nil
}
