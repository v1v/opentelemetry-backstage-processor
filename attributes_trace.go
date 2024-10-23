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
	labels map[string]RepoInfo
}

// newTracesProcessor returns a processor that adds attributes to all the spans.
// To construct the attributes processors, the use of the factory methods are required
// in order to validate the inputs.
func newSpanAttributesProcessor(logger *zap.Logger, config component.Config) *spanAttributesProcessor {
	cfg := config.(*Config)
	logger.Info("Fetching Backstage labels", zap.String("endpoint", cfg.Endpoint))
	labels, err := getRepositoryLabelsMap(cfg.Endpoint, string(cfg.Token))

	if err != nil {
		logger.Error("Failed to fetch labels", zap.Error(err))
	}
	logger.Info("Fetched GitHub repositories", zap.Int("number of repositories", len(labels)))

	// for demo purposes let's also create some fake labels
	labels["demo-devops-ingress-nginx"] = RepoInfo{
		Repo:     "ingress-nginx",
		Org:      "open-source",
		Division: "engineering",
	}
	labels["demo-devops-elasticsearch"] = RepoInfo{
		Repo:     "elasticsearch ",
		Org:      "platform",
		Division: "engineering",
	}
	labels["demo-devops-apm-server"] = RepoInfo{
		Repo:     "apm-server",
		Org:      "obs",
		Division: "engineering",
	}
	labels["demo-devops-opentelemetry-lambda"] = RepoInfo{
		Repo:     "opentelemetry-lambda",
		Org:      "open-source",
		Division: "engineering",
	}
	labels["demo-devops-elastic-otel-node"] = RepoInfo{
		Repo:     "elastic-otel-node",
		Org:      "obs",
		Division: "engineering",
	}
	labels["demo-devops-elastic-otel-java"] = RepoInfo{
		Repo:     "elastic-otel-java",
		Org:      "obs",
		Division: "engineering",
	}
	labels["demo-devops-setup-go"] = RepoInfo{
		Repo:     "setup-go",
		Org:      "open-source",
		Division: "engineering",
	}
	labels["demo-devops-demo"] = RepoInfo{
		Repo:     "demo",
		Org:      "platform",
		Division: "engineering",
	}
	labels["demo-devops-codeql-action"] = RepoInfo{
		Repo:     "codeql-action",
		Org:      "platform",
		Division: "engineering",
	}
	// end of fake labels

	return &spanAttributesProcessor{
		config: *cfg,
		logger: logger,
		labels: labels,
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
				if repo, found := attrs.Get("service.name"); found {
					a.logger.Info("Found service name", zap.String("service.name", repo.Str()))
					repoinfo, ok := a.labels[repo.Str()]
					if !ok {
						continue
					}
					attrs.PutStr("backstage.division", repoinfo.Division)
					attrs.PutStr("backstage.org", repoinfo.Org)
				}
			}
		}
	}
	return td, nil
}
