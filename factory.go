package backstageprocessor

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/processorhelper"
)

const (
	TracesStability = component.StabilityLevelAlpha
	LogsStability   = component.StabilityLevelAlpha
)

var processorCapabilities = consumer.Capabilities{MutatesData: true}

// Note: This isn't a valid configuration because the processor would do no work.
func createDefaultConfig() component.Config {
	return &Config{}
}

// NewFactory returns a new factory for the Attributes processor.
func NewFactory() processor.Factory {
	return processor.NewFactory(
		component.MustNewType("backstageprocessor"),
		createDefaultConfig,
		processor.WithTraces(createTracesProcessor, TracesStability),
		processor.WithLogs(createLogsProcessor, LogsStability))
}

func createTracesProcessor(
	ctx context.Context,
	set processor.CreateSettings,
	cfg component.Config,
	nextConsumer consumer.Traces,
) (processor.Traces, error) {

	// need to find out how we can create the maps with the labels once
	// rather than one per type of processor as it is done in the createLogsProcessor
	// function below.
	processor := newBackstageProcessor(set.Logger, cfg)
	return processorhelper.NewTracesProcessor(
		ctx,
		set,
		cfg,
		nextConsumer,
		processor.processTraces,
		processorhelper.WithCapabilities(processorCapabilities))
}

func createLogsProcessor(
	ctx context.Context,
	set processor.CreateSettings,
	cfg component.Config,
	nextLogsConsumer consumer.Logs,
) (processor.Logs, error) {

	processor := newBackstageProcessor(set.Logger, cfg)
	return processorhelper.NewLogsProcessor(
		ctx,
		set,
		cfg,
		nextLogsConsumer,
		processor.processLogs,
		processorhelper.WithCapabilities(processorCapabilities))
}
