# opentelemetry-backstage-processor

| Status        |           |
| ------------- |-----------|
| Stability     | [alpha]: traces, logs, metrics   |

[alpha]: https://github.com/open-telemetry/opentelemetry-collector#alpha

The Backstage processor enriches telemetry data (traces, logs, and metrics) with organizational metadata from [Backstage](https://backstage.io/). It automatically adds `backstage.org` and `backstage.division` attributes based on the `service.name` resource attribute, enabling better observability and organization of telemetry data.

## How It Works

The processor:
1. Fetches repository metadata from Backstage API on startup
2. Matches `service.name` from telemetry against Backstage entities
3. Adds organizational attributes (`backstage.org`, `backstage.division`) to all telemetry signals
4. Optionally refreshes metadata periodically in the background

For detailed information about the implementation and potential issues, see [BACKGROUND_REFRESH.md](docs/BACKGROUND_REFRESH.md).

## Configuration

```yaml
processors:
  backstageprocessor:
    # The Backstage API endpoint URL
    # Required
    endpoint: "https://backstage.example.com"

    # Authentication token for Backstage API
    # Required. Supports environment variable expansion: ${env:BACKSTAGE_TOKEN}
    token: "your-api-token"

    # Interval for automatic background refresh of Backstage metadata
    # Optional. If not specified or set to 0, metadata is fetched only once at startup.
    # Recommended: 5m to 15m for most use cases.
    # Examples: 30s, 5m, 1h
    # default = 0 (disabled)
    refresh_interval: 1h
```

### Complete Example

```yaml
receivers:
  otlp:
    protocols:
      grpc:
      http:

processors:
  batch:
  backstageprocessor:
    endpoint: "https://api.backstage.example.com"
    token: "${env:BACKSTAGE_API_TOKEN}"
    refresh_interval: 5m

exporters:
  otlp:
    endpoint: "https://observability.example.com:4317"

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch, backstageprocessor]
      exporters: [otlp]
    metrics:
      receivers: [otlp]
      processors: [backstageprocessor]
      exporters: [otlp]
    logs:
      receivers: [otlp]
      processors: [backstageprocessor]
      exporters: [otlp]
```

## Attributes Added

The processor adds the following attributes to all telemetry signals:

| Attribute | Description | Example |
|-----------|-------------|---------|
| `backstage.org` | Organization/team owning the service | `platform-team` |
| `backstage.division` | Business division or department | `engineering` |

If a service is not found in Backstage, the attributes are set to `"unknown"`.
