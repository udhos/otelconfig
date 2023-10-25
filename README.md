# otelconfig

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
	time.Sleep(5 * time.Second)
}

```

# Configuration

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
