package backstageprocessor

import (
	"testing"

	"go.opentelemetry.io/collector/component"
)

func TestConfigValidation(t *testing.T) {
	t.Run("config implements component.Config interface", func(t *testing.T) {
		var _ component.Config = (*Config)(nil)
	})

	t.Run("config can be created with values", func(t *testing.T) {
		config := &Config{
			Token:    "test-token",
			Endpoint: "https://backstage.example.com",
		}

		if config.Token != "test-token" {
			t.Errorf("Expected token to be 'test-token', got '%s'", config.Token)
		}
		if config.Endpoint != "https://backstage.example.com" {
			t.Errorf("Expected endpoint to be 'https://backstage.example.com', got '%s'", config.Endpoint)
		}
	})
}
