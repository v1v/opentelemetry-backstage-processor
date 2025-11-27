# Background Refresh Feature

## Overview

The backstage processor now supports automatic periodic refresh of Backstage labels without requiring a collector restart. This is implemented using a background goroutine that periodically queries the Backstage API.

## Configuration

To enable background refresh, add the `refresh_interval` field to your processor configuration:

```yaml
processors:
  backstageprocessor:
    endpoint: "https://backstage.example.com"
    token: "${env:BACKSTAGE_TOKEN}"
    refresh_interval: 5m  # Refresh every 5 minutes
```

If `refresh_interval` is not specified or set to `0`, the processor will only fetch labels once during initialization (original behavior).

## Implementation Details

### Thread Safety

The implementation uses `sync.RWMutex` to ensure thread-safe access to the backstage labels map:

- **Read operations** (during telemetry processing): Use `RLock()` for concurrent reads
- **Write operations** (during refresh): Use `Lock()` for exclusive write access

This ensures that telemetry processing can continue uninterrupted while background refreshes occur.

### Lifecycle Management

The background goroutine is properly managed through the processor lifecycle:

1. **Start**: Goroutine starts automatically if `refresh_interval > 0`
2. **Refresh Loop**: Uses `time.Ticker` to trigger periodic refreshes
3. **Graceful Shutdown**: 
   - Respects context cancellation
   - Closes the `done` channel when finished
   - Factory calls `Shutdown()` with timeout context

### Error Handling

Refresh errors are logged but do not terminate the goroutine:

```go
if err != nil {
    b.logger.Error("Failed to refresh backstage labels", zap.Error(err))
    continue  // Skip this iteration, try again next tick
}
```

This ensures temporary network issues or API failures don't break the processor.

## Potential Issues and Safeguards

### 1. Goroutine Leaks

**Issue**: If `Shutdown()` is not called, the background goroutine will continue running indefinitely.

**Safeguards**:
- Factory automatically registers `Shutdown()` via `processorhelper.WithShutdown()`
- OpenTelemetry Collector calls `Shutdown()` on all processors during graceful shutdown
- Goroutine monitors context cancellation and terminates when signaled

### 2. Memory Usage

**Issue**: The backstage labels map is duplicated during refresh (old + new map temporarily in memory).

**Impact**:
- For typical deployments with hundreds of services: negligible (~few KB)
- For large deployments with thousands of services: could be significant (~few MB)

**Mitigation**:
- Map replacement is atomic (no partial updates)
- Old map is garbage collected immediately after replacement
- No unbounded growth - map size is bounded by number of Backstage entities

### 3. Race Conditions

**Issue**: Concurrent access to the map without proper synchronization could cause data races.

**Safeguards**:
- `sync.RWMutex` protects all map access
- Read lock during telemetry processing (`processAttrs`)
- Write lock during refresh (`refreshLoop`)
- Tested with `go test -race` to detect race conditions

### 4. Refresh Interval Considerations

**Issue**: Very short refresh intervals could cause excessive API load.

**Recommendations**:
- Minimum recommended: 1 minute
- Typical use case: 5-15 minutes
- Consider Backstage API rate limits
- Balance between freshness and API load

### 5. Network Failures

**Issue**: Temporary network failures during refresh could clear the labels map if not handled properly.

**Safeguards**:
- Refresh errors are logged but don't modify the existing map
- Map is only updated on successful API response
- Previous labels remain valid until successful refresh

### 6. Startup Behavior

**Issue**: Initial labels fetch happens synchronously during processor creation.

**Impact**:
- Collector startup may be delayed if Backstage API is slow
- Startup failure if API is unreachable and no retry logic

**Note**: This is existing behavior not changed by background refresh feature. Background refresh only affects post-startup behavior.

## Testing

### Unit Tests

All existing tests pass, validating:
- Telemetry processing with backstage attributes
- Processor creation and configuration
- Thread-safe map access patterns

### Race Detection

Run tests with race detector to verify thread safety:

```bash
go test -race ./...
```

### Manual Testing

1. Start collector with `refresh_interval` configured
2. Verify logs show periodic "Refreshing backstage labels" messages
3. Update Backstage labels via API
4. Verify telemetry reflects new labels after next refresh
5. Shutdown collector and verify "Stopping refresh loop" message

## Performance Impact

- **Telemetry Processing**: Minimal overhead (RLock is very fast for read-heavy workloads)
- **Background Refresh**: 
  - CPU: One HTTP request per refresh interval
  - Memory: Temporary duplication during map replacement
  - Network: One API call per refresh interval

## Migration from Previous Version

No configuration changes required for existing deployments:

- **Without `refresh_interval`**: Behavior unchanged (fetch once at startup)
- **With `refresh_interval`**: Background refresh enabled automatically

## Future Improvements

Potential enhancements for future versions:

1. **Incremental Updates**: Query only changed entities instead of full refresh
2. **Backpressure**: Skip refresh if previous one is still in progress
3. **Metrics**: Expose refresh success/failure metrics
4. **Jitter**: Add random jitter to prevent thundering herd with multiple collectors
5. **Configurable Timeout**: Add separate timeout for refresh API calls
