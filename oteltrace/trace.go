// Package oteltrace provides helpers for otel tracing.
package oteltrace

import (
	"context"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/propagators/autoprop"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

const lib = "github.com/udhos/otelconfig"

// TraceStart initializes tracing.
//
// export OTELCONFIG_EXPORTER=jaeger
// export OTEL_TRACES_EXPORTER=jaeger
// export OTEL_PROPAGATORS=b3multi
// export OTEL_EXPORTER_OTLP_ENDPOINT=http://jaeger-collector:14268
//
// export OTELCONFIG_EXPORTER=grpc
// export OTEL_TRACES_EXPORTER=otlp
// export OTEL_PROPAGATORS=b3multi
// export OTEL_EXPORTER_OTLP_ENDPOINT=http://jaeger-collector:4317
//
// export OTELCONFIG_EXPORTER=http
// export OTEL_TRACES_EXPORTER=otlp
// export OTEL_PROPAGATORS=b3multi
// export OTEL_EXPORTER_OTLP_ENDPOINT=http://jaeger-collector:4318
func TraceStart(defaultService string, debug bool) (trace.Tracer, func(), error) {

	const me = "TraceStart"

	exporter := os.Getenv("OTELCONFIG_EXPORTER")

	if debug {
		log.Printf("%s: OTELCONFIG_EXPORTER=[%s]", me, exporter)
	}

	tp, errTracer := tracerProvider(defaultService, exporter, debug)
	if errTracer != nil {
		return nil, func() {}, errTracer
	}

	// Register our TracerProvider as the global so any imported
	// instrumentation in the future will default to using it.
	otel.SetTracerProvider(tp)

	// Invoke clean to shutdown cleanly and flush telemetry when the application exits.
	clean := func() {
		ctx, cancel1 := context.WithCancel(context.Background())
		defer cancel1()
		// Do not make the application hang when it is shutdown.
		ctx2, cancel2 := context.WithTimeout(ctx, time.Second*5)
		defer cancel2()
		if err := tp.Shutdown(ctx2); err != nil {
			log.Fatalf("trace shutdown: %v", err)
		}
	}

	tracePropagation(debug)

	return tp.Tracer(lib), clean, nil
}

/*
Open Telemetry tracing with Gin:

1) Initialize the tracing (see main.go)
2) Enable trace propagation (see tracePropagation below)
3) Use handler middleware (see main.go)
   import "go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
   router.Use(otelgin.Middleware("virtual-service"))
4) For http client, create a Request from Context (see backend.go)
   newCtx, span := b.tracer.Start(ctx, "backendHTTP.fetch")
   req, errReq := http.NewRequestWithContext(newCtx, "GET", u, nil)
   client := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
   resp, errGet := client.Do(req)
*/

// tracerProvider creates a trace provider.
// Service name precedence from higher to lower:
// 1. OTEL_SERVICE_NAME=mysrv
// 2. OTEL_RESOURCE_ATTRIBUTES=service.name=mysrv
// 3. defaultService="mysrv"
func tracerProvider(defaultService, exporter string, debug bool) (*tracesdk.TracerProvider, error) {

	const me = "tracerProvider"

	if debug {
		log.Printf("%s: service=%s exporter=%s", me, defaultService, exporter)
	}

	// Create the Jaeger exporter
	exp, err := createExporter(exporter)
	if err != nil {
		return nil, err
	}

	var rsrc *resource.Resource

	if defaultService == "" || hasServiceEnvVar() {
		rsrc = resource.NewWithAttributes(
			semconv.SchemaURL,
			//attribute.String("environment", environment),
			//attribute.Int64("ID", id),
		)
	} else {
		rsrc = resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(defaultService),
			//attribute.String("environment", environment),
			//attribute.Int64("ID", id),
		)
	}

	// Register the trace exporter with a TracerProvider, using a batch
	// span processor to aggregate spans before export.
	bsp := tracesdk.NewBatchSpanProcessor(exp)

	tp := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in a Resource.
		tracesdk.WithResource(rsrc),
		tracesdk.WithSpanProcessor(bsp),
	)
	return tp, nil
}

func createExporter(exporter string) (tracesdk.SpanExporter, error) {
	const me = "createExporter"
	switch exporter {
	case "jaeger":
		return jaeger.New(jaeger.WithCollectorEndpoint())
	case "grpc":
		client := otlptracegrpc.NewClient(
			otlptracegrpc.WithInsecure(),
		)
		return otlptrace.New(context.Background(), client)
	case "http":
		client := otlptracehttp.NewClient(
			otlptracehttp.WithInsecure(),
		)
		return otlptrace.New(context.Background(), client)
	case "stdout":
		return stdouttrace.New()
	}
	return nil, fmt.Errorf("%s: unrecognized exporter type: '%s'",
		me, exporter)

}

func hasServiceEnvVar() bool {
	const me = "hasServiceEnvVar"

	svc := os.Getenv("OTEL_SERVICE_NAME")

	if strings.TrimSpace(svc) != "" {
		log.Printf("%s: found OTEL_SERVICE_NAME=%s", me, svc)
		return true
	}

	attrs := os.Getenv("OTEL_RESOURCE_ATTRIBUTES")
	fields := strings.FieldsFunc(attrs, func(c rune) bool { return c == ',' })
	for _, f := range fields {
		key, val, _ := strings.Cut(f, "=")
		if key == "service.name" {
			log.Printf("%s: found OTEL_RESOURCE_ATTRIBUTES: %s=%s",
				me, key, val)
			return true
		}
	}

	return false
}

// tracePropagation enables trace propagation.
func tracePropagation(debug bool) {
	/*
		// In order to propagate trace context over the wire, a propagator must be registered with the OpenTelemetry API.
		// https://opentelemetry.io/docs/instrumentation/go/manual/
		//otel.SetTextMapPropagator(propagation.TraceContext{})
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
			b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader)),
			//propagation.Baggage{},
			//propagation.TraceContext{},
			//ot.OT{},
		))
	*/

	const me = "tracePropagation"

	prop := autoprop.NewTextMapPropagator(propagation.TraceContext{})

	if debug {
		fields := prop.Fields()
		slices.Sort(fields)
		log.Printf("%s: OTEL_PROPAGATORS='%s' fields: %v", me, os.Getenv("OTEL_PROPAGATORS"), fields)
	}

	otel.SetTextMapPropagator(prop)
}
