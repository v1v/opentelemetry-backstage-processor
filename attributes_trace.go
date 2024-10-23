package backstageprocessor

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	conventions "go.opentelemetry.io/collector/semconv/v1.7.0"
	"go.uber.org/zap"
)

// span keys
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

// newTracesProcessor returns a processor that adds attributes to all the spans.
// To construct the attributes processors, the use of the factory methods are required
// in order to validate the inputs.
func newBackstageProcessor(logger *zap.Logger, config component.Config) *backstageprocessor {
	cfg := config.(*Config)
	logger.Info("Fetching Backstage labels", zap.String("endpoint", cfg.Endpoint))
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
	if repo, found := attributes.Get(conventions.AttributeServiceName); found {
		b.logger.Debug("Found service name", zap.String(conventions.AttributeServiceName, repo.Str()))
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

// createDemoLabels only for illustrating how it works in the demo environment.
func createDemoLabels(original map[string]RepoInfo) {
	labels := make(map[string]RepoInfo)

	labels["demo-devops-bcn-ingress-nginx"] = RepoInfo{
		Repo:     "ingress-nginx",
		Org:      "open-source",
		Division: "engineering",
	}
	labels["demo-devops-bcn-elasticsearch"] = RepoInfo{
		Repo:     "elasticsearch",
		Org:      "platform",
		Division: "engineering",
	}
	labels["demo-devops-bcn-apm-server"] = RepoInfo{
		Repo:     "apm-server",
		Org:      "obs",
		Division: "engineering",
	}
	labels["demo-devops-bcn-opentelemetry-lambda"] = RepoInfo{
		Repo:     "opentelemetry-lambda",
		Org:      "open-source",
		Division: "engineering",
	}
	labels["demo-devops-bcn-elastic-otel-node"] = RepoInfo{
		Repo:     "elastic-otel-node",
		Org:      "obs",
		Division: "engineering",
	}
	labels["demo-devops-bcn-elastic-otel-java"] = RepoInfo{
		Repo:     "elastic-otel-java",
		Org:      "obs",
		Division: "engineering",
	}
	labels["demo-devops-bcn-setup-go"] = RepoInfo{
		Repo:     "setup-go",
		Org:      "open-source",
		Division: "engineering",
	}
	labels["demo-devops-bcn-demo"] = RepoInfo{
		Repo:     "demo",
		Org:      "platform",
		Division: "engineering",
	}
	labels["demo-devops-bcn-codeql-action"] = RepoInfo{
		Repo:     "codeql-action",
		Org:      "platform",
		Division: "engineering",
	}

	for key, value := range labels {
		original[key] = value
	}

}
