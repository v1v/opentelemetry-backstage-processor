package backstageprocessor

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

type spanAttributesProcessor struct {
	logger *zap.Logger
	config Config
}

// newTracesProcessor returns a processor that adds attributes to all the spans.
// To construct the attributes processors, the use of the factory methods are required
// in order to validate the inputs.
func newSpanAttributesProcessor(logger *zap.Logger, config component.Config) *spanAttributesProcessor {
	cfg := config.(*Config)
	return &spanAttributesProcessor{
		config: *cfg,
		logger: logger,
	}
}

func (a *spanAttributesProcessor) processTraces(ctx context.Context, td ptrace.Traces) (ptrace.Traces, error) {
	rss := td.ResourceSpans()
	for i := 0; i < rss.Len(); i++ {
		rs := rss.At(i)
		ilss := rs.ScopeSpans()
		for j := 0; j < ilss.Len(); j++ {
			ils := ilss.At(j)
			spans := ils.Spans()
			for k := 0; k < spans.Len(); k++ {
				span := spans.At(k)
				attrs := span.Attributes()
				if _, found := attrs.Get("MyKey"); found {
					continue
				}
				attrs.PutStr("MyKey", "MyValue")
			}
		}
	}
	return td, nil
}
