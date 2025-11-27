package backstageprocessor

import (
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configopaque"
)

// Config defines configuration for Resource processor.
type Config struct {
	Token           configopaque.String `mapstructure:"token"`
	Endpoint        string              `mapstructure:"endpoint"`
	RefreshInterval time.Duration       `mapstructure:"refresh_interval"`
}

var _ component.Config = (*Config)(nil)
