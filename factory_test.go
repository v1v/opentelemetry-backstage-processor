package backstageprocessor

import (
	"context"
	"testing"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/processor/processortest"
)

func TestNewFactory(t *testing.T) {
	factory := NewFactory()

	if factory == nil {
		t.Fatal("Expected factory to be created")
	}

	if factory.Type() != component.MustNewType("backstageprocessor") {
		t.Errorf("Expected type to be 'backstageprocessor', got '%s'", factory.Type())
	}
}

func TestCreateDefaultConfig(t *testing.T) {
	cfg := createDefaultConfig()

	if cfg == nil {
		t.Fatal("Expected default config to be created")
	}

	backstageCfg, ok := cfg.(*Config)
	if !ok {
		t.Fatal("Expected config to be of type *Config")
	}

	if backstageCfg.Token != "" {
		t.Error("Expected default token to be empty")
	}
	if backstageCfg.Endpoint != "" {
		t.Error("Expected default endpoint to be empty")
	}
}

func TestCreateTracesProcessor(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	backstageCfg := cfg.(*Config)
	backstageCfg.Endpoint = "https://backstage.example.com"
	backstageCfg.Token = "test-token"

	set := processortest.NewNopCreateSettings()
	nextConsumer := consumertest.NewNop()

	tp, err := factory.CreateTracesProcessor(context.Background(), set, cfg, nextConsumer)
	if err != nil {
		t.Fatalf("CreateTracesProcessor failed: %v", err)
	}

	if tp == nil {
		t.Fatal("Expected traces processor to be created")
	}
}

func TestCreateLogsProcessor(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	backstageCfg := cfg.(*Config)
	backstageCfg.Endpoint = "https://backstage.example.com"
	backstageCfg.Token = "test-token"

	set := processortest.NewNopCreateSettings()
	nextConsumer := consumertest.NewNop()

	lp, err := factory.CreateLogsProcessor(context.Background(), set, cfg, nextConsumer)
	if err != nil {
		t.Fatalf("CreateLogsProcessor failed: %v", err)
	}

	if lp == nil {
		t.Fatal("Expected logs processor to be created")
	}
}

func TestCreateMetricsProcessor(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	backstageCfg := cfg.(*Config)
	backstageCfg.Endpoint = "https://backstage.example.com"
	backstageCfg.Token = "test-token"

	set := processortest.NewNopCreateSettings()
	nextConsumer := consumertest.NewNop()

	mp, err := factory.CreateMetricsProcessor(context.Background(), set, cfg, nextConsumer)
	if err != nil {
		t.Fatalf("CreateMetricsProcessor failed: %v", err)
	}

	if mp == nil {
		t.Fatal("Expected metrics processor to be created")
	}
}

func TestProcessorCapabilities(t *testing.T) {
	if !processorCapabilities.MutatesData {
		t.Error("Expected processor to mutate data")
	}
}

func TestStabilityLevels(t *testing.T) {
	if LogsStability != component.StabilityLevelAlpha {
		t.Errorf("Expected logs stability to be Alpha, got %s", LogsStability)
	}
	if MetricsStability != component.StabilityLevelAlpha {
		t.Errorf("Expected metrics stability to be Alpha, got %s", MetricsStability)
	}
	if TracesStability != component.StabilityLevelAlpha {
		t.Errorf("Expected traces stability to be Alpha, got %s", TracesStability)
	}
}
