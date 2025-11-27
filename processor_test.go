package backstageprocessor

import (
	"context"
	"testing"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

func TestProcessAttrs(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		Endpoint: "https://backstage.example.com",
		Token:    "test-token",
	}

	backstageMap := map[string]RepoInfo{
		"test-service": {
			Repo:     "test-service",
			Org:      "test-org",
			Division: "test-division",
		},
	}

	processor := &backstageprocessor{
		logger:       logger,
		config:       *config,
		backstageMap: backstageMap,
	}

	t.Run("with known service name", func(t *testing.T) {
		attrs := pcommon.NewMap()
		attrs.PutStr(serviceNameKey, "test-service")

		processor.processAttrs(context.Background(), attrs)

		org, orgExists := attrs.Get(orgKey)
		division, divExists := attrs.Get(divisionKey)

		if !orgExists {
			t.Errorf("Expected org attribute to exist")
		}
		if !divExists {
			t.Errorf("Expected division attribute to exist")
		}
		if org.Str() != "test-org" {
			t.Errorf("Expected org to be 'test-org', got '%s'", org.Str())
		}
		if division.Str() != "test-division" {
			t.Errorf("Expected division to be 'test-division', got '%s'", division.Str())
		}
	})

	t.Run("with unknown service name", func(t *testing.T) {
		attrs := pcommon.NewMap()
		attrs.PutStr(serviceNameKey, "unknown-service")

		processor.processAttrs(context.Background(), attrs)

		org, orgExists := attrs.Get(orgKey)
		division, divExists := attrs.Get(divisionKey)

		if !orgExists {
			t.Errorf("Expected org attribute to exist")
		}
		if !divExists {
			t.Errorf("Expected division attribute to exist")
		}
		if org.Str() != unknown {
			t.Errorf("Expected org to be '%s', got '%s'", unknown, org.Str())
		}
		if division.Str() != unknown {
			t.Errorf("Expected division to be '%s', got '%s'", unknown, division.Str())
		}
	})

	t.Run("without service name", func(t *testing.T) {
		attrs := pcommon.NewMap()

		processor.processAttrs(context.Background(), attrs)

		_, orgExists := attrs.Get(orgKey)
		_, divExists := attrs.Get(divisionKey)

		if orgExists {
			t.Errorf("Expected org attribute not to exist")
		}
		if divExists {
			t.Errorf("Expected division attribute not to exist")
		}
	})
}

func TestProcessTraces(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		Endpoint: "https://backstage.example.com",
		Token:    "test-token",
	}

	backstageMap := map[string]RepoInfo{
		"trace-service": {
			Repo:     "trace-service",
			Org:      "trace-org",
			Division: "trace-division",
		},
	}

	processor := &backstageprocessor{
		logger:       logger,
		config:       *config,
		backstageMap: backstageMap,
	}

	t.Run("adds backstage attributes to traces", func(t *testing.T) {
		traces := ptrace.NewTraces()
		rs := traces.ResourceSpans().AppendEmpty()
		rs.Resource().Attributes().PutStr(serviceNameKey, "trace-service")

		span := rs.ScopeSpans().AppendEmpty().Spans().AppendEmpty()
		span.SetName("test-span")

		result, err := processor.processTraces(context.Background(), traces)
		if err != nil {
			t.Fatalf("processTraces failed: %v", err)
		}

		attrs := result.ResourceSpans().At(0).Resource().Attributes()
		org, _ := attrs.Get(orgKey)
		division, _ := attrs.Get(divisionKey)

		if org.Str() != "trace-org" {
			t.Errorf("Expected org to be 'trace-org', got '%s'", org.Str())
		}
		if division.Str() != "trace-division" {
			t.Errorf("Expected division to be 'trace-division', got '%s'", division.Str())
		}
	})
}

func TestProcessLogs(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		Endpoint: "https://backstage.example.com",
		Token:    "test-token",
	}

	backstageMap := map[string]RepoInfo{
		"log-service": {
			Repo:     "log-service",
			Org:      "log-org",
			Division: "log-division",
		},
	}

	processor := &backstageprocessor{
		logger:       logger,
		config:       *config,
		backstageMap: backstageMap,
	}

	t.Run("adds backstage attributes to logs", func(t *testing.T) {
		logs := plog.NewLogs()
		rl := logs.ResourceLogs().AppendEmpty()
		rl.Resource().Attributes().PutStr(serviceNameKey, "log-service")

		logRecord := rl.ScopeLogs().AppendEmpty().LogRecords().AppendEmpty()
		logRecord.Body().SetStr("test log message")

		result, err := processor.processLogs(context.Background(), logs)
		if err != nil {
			t.Fatalf("processLogs failed: %v", err)
		}

		attrs := result.ResourceLogs().At(0).Resource().Attributes()
		org, _ := attrs.Get(orgKey)
		division, _ := attrs.Get(divisionKey)

		if org.Str() != "log-org" {
			t.Errorf("Expected org to be 'log-org', got '%s'", org.Str())
		}
		if division.Str() != "log-division" {
			t.Errorf("Expected division to be 'log-division', got '%s'", division.Str())
		}
	})
}

func TestProcessMetrics(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		Endpoint: "https://backstage.example.com",
		Token:    "test-token",
	}

	backstageMap := map[string]RepoInfo{
		"metric-service": {
			Repo:     "metric-service",
			Org:      "metric-org",
			Division: "metric-division",
		},
	}

	processor := &backstageprocessor{
		logger:       logger,
		config:       *config,
		backstageMap: backstageMap,
	}

	t.Run("adds backstage attributes to gauge metrics", func(t *testing.T) {
		metrics := pmetric.NewMetrics()
		rm := metrics.ResourceMetrics().AppendEmpty()
		rm.Resource().Attributes().PutStr(serviceNameKey, "metric-service")

		metric := rm.ScopeMetrics().AppendEmpty().Metrics().AppendEmpty()
		metric.SetName("test.gauge")
		gauge := metric.SetEmptyGauge()
		dp := gauge.DataPoints().AppendEmpty()
		dp.SetDoubleValue(42.0)
		dp.Attributes().PutStr(serviceNameKey, "metric-service")

		result, err := processor.processMetrics(context.Background(), metrics)
		if err != nil {
			t.Fatalf("processMetrics failed: %v", err)
		}

		attrs := result.ResourceMetrics().At(0).Resource().Attributes()
		org, _ := attrs.Get(orgKey)
		division, _ := attrs.Get(divisionKey)

		if org.Str() != "metric-org" {
			t.Errorf("Expected org to be 'metric-org', got '%s'", org.Str())
		}
		if division.Str() != "metric-division" {
			t.Errorf("Expected division to be 'metric-division', got '%s'", division.Str())
		}

		// Check data point attributes
		dpAttrs := result.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Gauge().DataPoints().At(0).Attributes()
		dpOrg, _ := dpAttrs.Get(orgKey)
		dpDivision, _ := dpAttrs.Get(divisionKey)

		if dpOrg.Str() != "metric-org" {
			t.Errorf("Expected data point org to be 'metric-org', got '%s'", dpOrg.Str())
		}
		if dpDivision.Str() != "metric-division" {
			t.Errorf("Expected data point division to be 'metric-division', got '%s'", dpDivision.Str())
		}
	})

	t.Run("adds backstage attributes to sum metrics", func(t *testing.T) {
		metrics := pmetric.NewMetrics()
		rm := metrics.ResourceMetrics().AppendEmpty()
		rm.Resource().Attributes().PutStr(serviceNameKey, "metric-service")

		metric := rm.ScopeMetrics().AppendEmpty().Metrics().AppendEmpty()
		metric.SetName("test.sum")
		sum := metric.SetEmptySum()
		dp := sum.DataPoints().AppendEmpty()
		dp.SetDoubleValue(100.0)
		dp.Attributes().PutStr(serviceNameKey, "metric-service")

		result, err := processor.processMetrics(context.Background(), metrics)
		if err != nil {
			t.Fatalf("processMetrics failed: %v", err)
		}

		// Check data point attributes
		dpAttrs := result.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Sum().DataPoints().At(0).Attributes()
		dpOrg, _ := dpAttrs.Get(orgKey)
		dpDivision, _ := dpAttrs.Get(divisionKey)

		if dpOrg.Str() != "metric-org" {
			t.Errorf("Expected data point org to be 'metric-org', got '%s'", dpOrg.Str())
		}
		if dpDivision.Str() != "metric-division" {
			t.Errorf("Expected data point division to be 'metric-division', got '%s'", dpDivision.Str())
		}
	})
}

func TestNewBackstageProcessor(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{
		Endpoint: "https://invalid-endpoint.example.com",
		Token:    "test-token",
	}

	processor := newBackstageProcessor(logger, config)

	if processor == nil {
		t.Fatal("Expected processor to be created even with invalid endpoint")
	}

	if processor.logger != logger {
		t.Error("Expected logger to be set correctly")
	}

	if processor.backstageMap == nil {
		t.Error("Expected backstageMap to be initialized")
	}

	if len(processor.backstageMap) != 0 {
		t.Error("Expected backstageMap to be empty when endpoint is invalid")
	}
}
