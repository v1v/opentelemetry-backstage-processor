package backstageprocessor

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

const serviceNameKey = "service.name"

// atribute keys
const (
	orgKey      = "backstage.org"
	divisionKey = "backstage.division"
	unknown     = "unknown"
)

type backstageprocessor struct {
	logger       *zap.Logger
	config       Config
	backstageMap map[string]RepoInfo
}

// newBackstageProcessor returns a processor that adds attributes to all the spans, logs and metrics.
// To construct the attributes processors, the use of the factory methods are required
// in order to validate the inputs.
func newBackstageProcessor(logger *zap.Logger, config component.Config) *backstageprocessor {
	cfg := config.(*Config)
	logger.Info("Fetching Backstage labels", zap.String("endpoint", cfg.Endpoint))

	// eventually this should support some dynamism in the labels
	// so new labels can be picked up without a restart
	labels, err := getRepositoryLabelsMap(cfg.Endpoint, string(cfg.Token))

	if err != nil {
		logger.Error("Failed to fetch the Backstage labels", zap.Error(err))
		return &backstageprocessor{
			config:       *cfg,
			logger:       logger,
			backstageMap: map[string]RepoInfo{},
		}
	}
	logger.Info("Fetched GitHub repositories", zap.Int("number of repositories", len(labels)))

	// Append demo labels to the existing labels
	createDemoLabels(labels)

	return &backstageprocessor{
		config:       *cfg,
		logger:       logger,
		backstageMap: labels,
	}
}

// processTraces processes the incoming data
// and returns the data to be sent to the next component
func (b *backstageprocessor) processTraces(ctx context.Context, batch ptrace.Traces) (ptrace.Traces, error) {
	for i := 0; i < batch.ResourceSpans().Len(); i++ {
		rs := batch.ResourceSpans().At(i)
		b.processResourceSpan(ctx, rs)
	}
	return batch, nil
}

// processResourceSpan processes the RS and all of its spans
func (b *backstageprocessor) processResourceSpan(ctx context.Context, rs ptrace.ResourceSpans) {
	rsAttrs := rs.Resource().Attributes()

	// Attributes can be part of a resource span
	b.processAttrs(ctx, rsAttrs)

	for j := 0; j < rs.ScopeSpans().Len(); j++ {
		ils := rs.ScopeSpans().At(j)
		for k := 0; k < ils.Spans().Len(); k++ {
			span := ils.Spans().At(k)
			spanAttrs := span.Attributes()

			// Attributes can also be part of span
			b.processAttrs(ctx, spanAttrs)
		}
	}
}

// processAttrs adds backstage metadata tags to resource based on service.name map
func (b *backstageprocessor) processAttrs(_ context.Context, attributes pcommon.Map) {
	if repo, found := attributes.Get(serviceNameKey); found {
		b.logger.Debug("Found service name", zap.String(serviceNameKey, repo.Str()))
		org := unknown
		division := unknown
		repoinfo, ok := b.backstageMap[repo.Str()]
		if ok {
			org = repoinfo.Org
			division = repoinfo.Division
		}
		attributes.PutStr(divisionKey, division)
		attributes.PutStr(orgKey, org)
	} else {
		b.logger.Debug("Not found service name", zap.Any("attributes", attributes))
	}
}

// processLogs processes the incoming data
// and returns the data to be sent to the next component
func (b *backstageprocessor) processLogs(ctx context.Context, logs plog.Logs) (plog.Logs, error) {
	for i := 0; i < logs.ResourceLogs().Len(); i++ {
		rl := logs.ResourceLogs().At(i)
		b.processResourceLog(ctx, rl)
	}
	return logs, nil
}

// processResourceLog processes the log resource and all of its logs and then returns the last
// view metric context. The context can be used for tests
func (b *backstageprocessor) processResourceLog(ctx context.Context, rl plog.ResourceLogs) {
	rsAttrs := rl.Resource().Attributes()

	b.processAttrs(ctx, rsAttrs)

	for j := 0; j < rl.ScopeLogs().Len(); j++ {
		ils := rl.ScopeLogs().At(j)
		for k := 0; k < ils.LogRecords().Len(); k++ {
			log := ils.LogRecords().At(k)
			b.processAttrs(ctx, log.Attributes())
		}
	}
}

// processMetrics process metrics and add the backstage lable metadata.
func (b *backstageprocessor) processMetrics(ctx context.Context, metrics pmetric.Metrics) (pmetric.Metrics, error) {
	for i := 0; i < metrics.ResourceMetrics().Len(); i++ {
		rm := metrics.ResourceMetrics().At(i)
		b.processResourceMetric(ctx, rm)
	}
	return metrics, nil
}

func (b *backstageprocessor) processResourceMetric(ctx context.Context, rm pmetric.ResourceMetrics) {
	rsAttrs := rm.Resource().Attributes()

	b.processAttrs(ctx, rsAttrs)

	for j := 0; j < rm.ScopeMetrics().Len(); j++ {
		ils := rm.ScopeMetrics().At(j)
		for k := 0; k < ils.Metrics().Len(); k++ {
			metric := ils.Metrics().At(k)
			b.processMetricAttributes(ctx, metric)
		}
	}
}

// processMetricAttributes Attributes are provided for each log and trace, but not at the metric level
// Need to process attributes for every data point within a metric.
func (b *backstageprocessor) processMetricAttributes(ctx context.Context, metric pmetric.Metric) {
	switch metric.Type() {
	case pmetric.MetricTypeGauge:
		dps := metric.Gauge().DataPoints()
		for i := 0; i < dps.Len(); i++ {
			b.processAttrs(ctx, dps.At(i).Attributes())
		}
	case pmetric.MetricTypeSum:
		dps := metric.Sum().DataPoints()
		for i := 0; i < dps.Len(); i++ {
			b.processAttrs(ctx, dps.At(i).Attributes())
		}
	case pmetric.MetricTypeHistogram:
		dps := metric.Histogram().DataPoints()
		for i := 0; i < dps.Len(); i++ {
			b.processAttrs(ctx, dps.At(i).Attributes())
		}
	case pmetric.MetricTypeExponentialHistogram:
		dps := metric.ExponentialHistogram().DataPoints()
		for i := 0; i < dps.Len(); i++ {
			b.processAttrs(ctx, dps.At(i).Attributes())
		}
	case pmetric.MetricTypeSummary:
		dps := metric.Summary().DataPoints()
		for i := 0; i < dps.Len(); i++ {
			b.processAttrs(ctx, dps.At(i).Attributes())
		}
	case pmetric.MetricTypeEmpty:
	}
}
