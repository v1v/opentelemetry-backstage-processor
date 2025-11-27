# opentelemetry-backstage-processor

| Status        |           |
| ------------- |-----------|
| Stability     | [alpha]: traces, logs   |

[alpha]: https://github.com/open-telemetry/opentelemetry-collector#alpha

## Local Development

### Prerequisites

- [Go](https://golang.org/dl/)
- [Gh](https://cli.github.com/)

### Install tools

#### ocb

The [OpenTelemetry Collector Builder (OCB)](https://opentelemetry.io/docs/collector/custom-collector/) can be installed by running:

```shell
make install-ocb
```

### Build the collector

```shell
make build
```

### Run the collector

```shell
make run
```

## Configuration

The backstage processor supports the following configuration options:

```yaml
processors:
  backstageprocessor:
    endpoint: "https://backstage.example.com"  # Backstage API endpoint
    token: "your-api-token"                    # Backstage API token
    refresh_interval: 2h                       # Optional: refresh labels periodically
```

### Configuration Options

- `endpoint` (required): The URL of your Backstage API
- `token` (required): API token for authenticating with Backstage
- `refresh_interval` (optional): Duration between automatic refreshes of Backstage labels
  - If not specified or set to `0`, labels are fetched only once at startup
  - Examples: `30s`, `5m`, `1h`
  - Recommended: 1h minutes for most use cases

### Background Refresh Feature

When `refresh_interval` is configured, the processor automatically refreshes Backstage labels in the background without requiring a collector restart. This feature:

- Uses thread-safe concurrent access patterns
- Continues processing telemetry during refreshes
- Handles API failures gracefully without affecting telemetry flow
- Shuts down cleanly with the collector

For detailed information about the implementation and potential issues, see [BACKGROUND_REFRESH.md](docs/BACKGROUND_REFRESH.md).
