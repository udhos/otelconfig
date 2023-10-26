[![license](http://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/udhos/otelconfig/blob/main/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/udhos/otelconfig)](https://goreportcard.com/report/github.com/udhos/otelconfig)
[![Go Reference](https://pkg.go.dev/badge/github.com/udhos/otelconfig.svg)](https://pkg.go.dev/github.com/udhos/otelconfig)

# otelconfig

[otelconfig](https://github.com/udhos/otelconfig) provides a helper to quickly initialize OpenTelemetry tracing for a Go application. Then one can use standard OTEL_ env vars to customize tracing behavior.

# Example

See [examples/oteltrace-example/main.go](examples/oteltrace-example/main.go).

# Usage

```go
import "github.com/udhos/otelconfig/oteltrace"
import "go.opentelemetry.io/otel/trace"

//
// initialize tracing
//

var tracer trace.Tracer

{
    options := oteltrace.TraceOptions{
        DefaultService:     "my-program",
        NoopTracerProvider: false,
        Debug:              true,
    }

    tr, cancel, errTracer := oteltrace.TraceStart(options)

    if errTracer != nil {
        log.Fatalf("tracer: %v", errTracer)
    }

    defer cancel()

    tracer = tr
}

// use tracer to create spans

work(context.TODO(), tracer)

// ...

func work(ctx context.Context, tracer trace.Tracer) {
	_, span := tracer.Start(ctx, "work")
	defer span.End()
}

```

# Configuration

General configuration: https://opentelemetry.io/docs/concepts/sdk-configuration/general-sdk-configuration/

Exporter configuration: https://opentelemetry.io/docs/concepts/sdk-configuration/otlp-exporter-configuration/

```bash
export OTELCONFIG_EXPORTER=jaeger|grpc|http|stdout  ;#     Protocol    default: grpc
export OTEL_TRACES_EXPORTER=jaeger|otlp             ;#     Data Format default: otlp
export OTEL_PROPAGATORS=b3multi                     ;# [1] Propagator  default: tracecontext,baggage
export OTEL_EXPORTER_OTLP_ENDPOINT=http://host:port ;#     Endpoint    default: [2]

# [1] Propagators: tracecontext,baggage,b3,b3multi,jaeger,xray,ottrace,none
#
# [2] Default endpoint: http://localhost:4317 for grpc
#                       http://localhost:4318 for http
#
# Service name precedence from higher to lower:
# 1. OTEL_SERVICE_NAME=mysrv
# 2. OTEL_RESOURCE_ATTRIBUTES=service.name=mysrv
# 3. TraceOptions.DefaultService="mysrv"
```

Examples:

```bash
export OTELCONFIG_EXPORTER=jaeger
export OTEL_TRACES_EXPORTER=jaeger
export OTEL_PROPAGATORS=b3multi
export OTEL_EXPORTER_OTLP_ENDPOINT=http://jaeger-collector:14268
oteltrace-example

export OTELCONFIG_EXPORTER=grpc
export OTEL_TRACES_EXPORTER=otlp
export OTEL_PROPAGATORS=b3multi
export OTEL_EXPORTER_OTLP_ENDPOINT=http://jaeger-collector:4317
oteltrace-example

export OTELCONFIG_EXPORTER=http
export OTEL_TRACES_EXPORTER=otlp
export OTEL_PROPAGATORS=b3multi
export OTEL_EXPORTER_OTLP_ENDPOINT=http://jaeger-collector:4318
oteltrace-example
```
