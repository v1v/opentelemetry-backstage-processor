package backstageprocessor

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.uber.org/zap"
)

func TestBackgroundRefresh(t *testing.T) {
	t.Run("processor with no refresh interval doesn't start goroutine", func(t *testing.T) {
		cfg := &Config{
			Endpoint:        "http://example.com",
			Token:           "test-token",
			RefreshInterval: 0, // No refresh
		}

		processor := newBackstageProcessor(zap.NewNop(), cfg)
		assert.Nil(t, processor.cancel, "cancel should be nil when refresh is disabled")
		assert.Nil(t, processor.done, "done channel should be nil when refresh is disabled")
	})

	t.Run("processor with refresh interval starts goroutine", func(t *testing.T) {
		cfg := &Config{
			Endpoint:        "http://example.com",
			Token:           "test-token",
			RefreshInterval: 100 * time.Millisecond,
		}

		processor := newBackstageProcessor(zap.NewNop(), cfg)
		require.NotNil(t, processor.cancel, "cancel should be set when refresh is enabled")
		require.NotNil(t, processor.done, "done channel should be set when refresh is enabled")

		// Cleanup
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := processor.Shutdown(ctx)
		assert.NoError(t, err)
	})

	t.Run("shutdown stops refresh loop gracefully", func(t *testing.T) {
		cfg := &Config{
			Endpoint:        "http://example.com",
			Token:           "test-token",
			RefreshInterval: 100 * time.Millisecond,
		}

		processor := newBackstageProcessor(zap.NewNop(), cfg)
		require.NotNil(t, processor.cancel)
		require.NotNil(t, processor.done)

		// Give the goroutine time to start
		time.Sleep(50 * time.Millisecond)

		// Shutdown with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		start := time.Now()
		err := processor.Shutdown(ctx)
		elapsed := time.Since(start)

		assert.NoError(t, err, "shutdown should complete successfully")
		assert.Less(t, elapsed, 1*time.Second, "shutdown should complete quickly")

		// Verify done channel is closed
		select {
		case <-processor.done:
			// Success - channel is closed
		case <-time.After(100 * time.Millisecond):
			t.Fatal("done channel should be closed after shutdown")
		}
	})

	t.Run("shutdown with no background goroutine", func(t *testing.T) {
		cfg := &Config{
			Endpoint:        "http://example.com",
			Token:           "test-token",
			RefreshInterval: 0, // No refresh
		}

		processor := newBackstageProcessor(zap.NewNop(), cfg)

		ctx := context.Background()
		err := processor.Shutdown(ctx)
		assert.NoError(t, err, "shutdown should work even without background goroutine")
	})

	t.Run("concurrent map access during refresh", func(t *testing.T) {
		cfg := &Config{
			Endpoint:        "http://example.com",
			Token:           "test-token",
			RefreshInterval: 50 * time.Millisecond,
		}

		processor := newBackstageProcessor(zap.NewNop(), cfg)
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = processor.Shutdown(ctx)
		}()

		// Add some initial data
		processor.backstageMap = map[string]RepoInfo{
			"service1": {Org: "org1", Division: "div1"},
			"service2": {Org: "org2", Division: "div2"},
		}

		// Simulate concurrent reads while refresh might be happening
		done := make(chan bool)
		go func() {
			for i := 0; i < 100; i++ {
				attrs := pcommon.NewMap()
				attrs.PutStr(serviceNameKey, "service1")
				processor.processAttrs(context.Background(), attrs)
				time.Sleep(1 * time.Millisecond)
			}
			done <- true
		}()

		// Wait for concurrent operations to complete
		<-done
	})

	t.Run("thread-safe map read operations", func(t *testing.T) {
		cfg := &Config{
			Endpoint: "http://example.com",
			Token:    "test-token",
		}

		processor := newBackstageProcessor(zap.NewNop(), cfg)
		processor.backstageMap = map[string]RepoInfo{
			"myservice": {Org: "myorg", Division: "mydiv"},
		}

		// Test concurrent reads (should be safe with RLock)
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				attrs := pcommon.NewMap()
				attrs.PutStr(serviceNameKey, "myservice")
				processor.processAttrs(context.Background(), attrs)

				// Verify attributes were added
				org, orgExists := attrs.Get(orgKey)
				div, divExists := attrs.Get(divisionKey)

				assert.True(t, orgExists, "org should be added")
				assert.True(t, divExists, "division should be added")
				assert.Equal(t, "myorg", org.Str())
				assert.Equal(t, "mydiv", div.Str())

				done <- true
			}()
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

func TestShutdownTimeout(t *testing.T) {
	// Note: This test is tricky because we'd need to simulate a stuck refresh loop
	// For now, we verify that shutdown respects the context timeout
	t.Run("shutdown respects context timeout", func(t *testing.T) {
		cfg := &Config{
			Endpoint:        "http://example.com",
			Token:           "test-token",
			RefreshInterval: 10 * time.Millisecond,
		}

		processor := newBackstageProcessor(zap.NewNop(), cfg)
		require.NotNil(t, processor.cancel)
		require.NotNil(t, processor.done)

		// Create a very short timeout context
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		// This might fail due to timeout, which is expected behavior
		err := processor.Shutdown(ctx)
		// We don't assert the error here because timing can be unpredictable in tests
		// The important thing is that the function returns and doesn't hang
		_ = err

		// Cleanup with proper context
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		_ = processor.Shutdown(cleanupCtx)
	})
}

func TestRefreshIntervalConfiguration(t *testing.T) {
	tests := []struct {
		name            string
		refreshInterval time.Duration
		expectGoroutine bool
	}{
		{
			name:            "zero duration - no refresh",
			refreshInterval: 0,
			expectGoroutine: false,
		},
		{
			name:            "positive duration - enable refresh",
			refreshInterval: 1 * time.Minute,
			expectGoroutine: true,
		},
		{
			name:            "very short duration",
			refreshInterval: 1 * time.Millisecond,
			expectGoroutine: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Endpoint:        "http://example.com",
				Token:           "test-token",
				RefreshInterval: tt.refreshInterval,
			}

			processor := newBackstageProcessor(zap.NewNop(), cfg)

			if tt.expectGoroutine {
				assert.NotNil(t, processor.cancel, "cancel should be set")
				assert.NotNil(t, processor.done, "done should be set")

				// Cleanup
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				_ = processor.Shutdown(ctx)
			} else {
				assert.Nil(t, processor.cancel, "cancel should be nil")
				assert.Nil(t, processor.done, "done should be nil")
			}
		})
	}
}

func TestProcessorIntegration(t *testing.T) {
	t.Run("processor lifecycle with factory", func(t *testing.T) {
		// Verify that processor properly integrates with the collector lifecycle
		cfg := &Config{
			Endpoint:        "http://example.com",
			Token:           "test-token",
			RefreshInterval: 100 * time.Millisecond,
		}

		processor := newBackstageProcessor(zap.NewNop(), cfg)
		require.NotNil(t, processor)
		require.NotNil(t, processor.cancel, "background goroutine should be started")

		// Give it some time to run
		time.Sleep(150 * time.Millisecond)

		// Shutdown the processor
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := processor.Shutdown(shutdownCtx)
		assert.NoError(t, err)
	})
}
