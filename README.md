# otelconfig

General configuration: https://opentelemetry.io/docs/concepts/sdk-configuration/general-sdk-configuration/

Exporter configuration: https://opentelemetry.io/docs/concepts/sdk-configuration/otlp-exporter-configuration/

```bash
export OTELCONFIG_EXPORTER=jaeger|grpc|http|stdout
export OTEL_TRACES_EXPORTER=jaeger|otlp
export OTEL_PROPAGATORS=b3multi
```

Examples:

```bash
export OTELCONFIG_EXPORTER=jaeger
export OTEL_TRACES_EXPORTER=jaeger
export OTEL_PROPAGATORS=b3multi
export OTEL_EXPORTER_OTLP_ENDPOINT=http://jaeger-collector:14268

export OTELCONFIG_EXPORTER=grpc
export OTEL_TRACES_EXPORTER=otlp
export OTEL_PROPAGATORS=b3multi
export OTEL_EXPORTER_OTLP_ENDPOINT=http://jaeger-collector:4317

export OTELCONFIG_EXPORTER=http
export OTEL_TRACES_EXPORTER=otlp
export OTEL_PROPAGATORS=b3multi
export OTEL_EXPORTER_OTLP_ENDPOINT=http://jaeger-collector:4318
```
