package backstageprocessor

import (
	"go.opentelemetry.io/collector/component"
)

// Config defines configuration for Resource processor.
type Config struct {
}

var _ component.Config = (*Config)(nil)
