// Package oteltrace provides helpers for otel tracing.
package oteltrace

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
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

// TraceOptions provides options for TraceStart.
type TraceOptions struct {
	DefaultService     string
	NoopTracerProvider bool // Disable tracer
	NoopPropagator     bool // Disable propagator
	Debug              bool
}

// TraceStart initializes tracing.
//
// These env vars become available for customization at runtime:
//
//	# Example for Jaeger
//	export OTELCONFIG_EXPORTER=jaeger
//	export OTEL_TRACES_EXPORTER=jaeger
//	export OTEL_PROPAGATORS=b3multi
//	export OTEL_EXPORTER_OTLP_ENDPOINT=http://jaeger-collector:14268
//
//	# Example for gRPC and OTLP
//	export OTELCONFIG_EXPORTER=grpc
//	export OTEL_TRACES_EXPORTER=otlp
//	export OTEL_PROPAGATORS=b3multi
//	export OTEL_EXPORTER_OTLP_ENDPOINT=http://jaeger-collector:4317
//
//	# Example for HTTP and OTLP
//	export OTELCONFIG_EXPORTER=http
//	export OTEL_TRACES_EXPORTER=otlp
//	export OTEL_PROPAGATORS=b3multi
//	export OTEL_EXPORTER_OTLP_ENDPOINT=http://jaeger-collector:4318
func TraceStart(options TraceOptions) (trace.Tracer, func(), error) {

	const me = "TraceStart"

	exporter := getEnv(me, "OTELCONFIG_EXPORTER", options.Debug)

	otelEndpoint := getEnv(me, "OTEL_EXPORTER_OTLP_ENDPOINT", options.Debug)

	var tp trace.TracerProvider
	clean := func() {}

	if options.NoopTracerProvider {
		tp = trace.NewNoopTracerProvider()
	} else {
		p, errTracer := tracerProvider(options.DefaultService, exporter, otelEndpoint, options.Debug)
		if errTracer != nil {
			return nil, clean, errTracer
		}
		tp = p

		// Invoke clean to shutdown cleanly and flush telemetry when the application exits.
		clean = func() {
			ctx, cancel1 := context.WithCancel(context.Background())
			defer cancel1()
			// Do not make the application hang when it is shutdown.
			ctx2, cancel2 := context.WithTimeout(ctx, time.Second*5)
			defer cancel2()
			if err := p.Shutdown(ctx2); err != nil {
				log.Fatalf("trace shutdown: %v", err)
			}
		}
	}

	// Register our TracerProvider as the global so any imported
	// instrumentation in the future will default to using it.
	otel.SetTracerProvider(tp)

	if !options.NoopPropagator {
		tracePropagation(options.Debug)
	}

	return tp.Tracer(lib), clean, nil
}

func getEnv(caller, key string, debug bool) string {
	value := os.Getenv(key)
	if debug {
		log.Printf("%s: %s='%s'", caller, key, value)
	}
	return value
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
func tracerProvider(defaultService, exporter, otelEndpoint string, debug bool) (*tracesdk.TracerProvider, error) {

	const me = "tracerProvider"

	if debug {
		log.Printf("%s: service='%s' exporter='%s'", me, defaultService, exporter)
	}

	// Create the Jaeger exporter
	exp, err := createExporter(exporter, otelEndpoint, debug)
	if err != nil {
		return nil, err
	}

	var rsrc *resource.Resource

	if defaultService == "" || hasServiceEnvVar(debug) {
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

	tp := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in a Resource.
		tracesdk.WithResource(rsrc),
	)

	return tp, nil
}

func createExporter(exporter, otelEndpoint string, debug bool) (tracesdk.SpanExporter, error) {
	const me = "createExporter"
	switch exporter {
	case "jaeger":
		// JaegerURL:          env.String("JAEGER_URL", "http://jaeger-collector:14268/api/traces"),
		// exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
		if otelEndpoint == "" {
			return jaeger.New(jaeger.WithCollectorEndpoint())
		}
		jaegerEndpoint, errJoin := url.JoinPath(otelEndpoint, "/api/traces")
		if errJoin != nil {
			return nil, errJoin
		}
		if debug {
			log.Printf("%s: jaeger endpoint: %s", me, jaegerEndpoint)
		}
		return jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerEndpoint)))
	case "", "grpc":
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

func hasServiceEnvVar(debug bool) bool {
	const me = "hasServiceEnvVar"

	if svc := getEnv(me, "OTEL_SERVICE_NAME", debug); strings.TrimSpace(svc) != "" {
		if debug {
			log.Printf("%s: found OTEL_SERVICE_NAME='%s'", me, svc)
		}
		return true
	}

	attrs := getEnv(me, "OTEL_RESOURCE_ATTRIBUTES", debug)
	fields := strings.FieldsFunc(attrs, func(c rune) bool { return c == ',' })
	for _, f := range fields {
		key, val, _ := strings.Cut(f, "=")
		if key == "service.name" {
			if debug {
				log.Printf("%s: found OTEL_RESOURCE_ATTRIBUTES: %s='%s'",
					me, key, val)
			}
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
		getEnv(me, "OTEL_PROPAGATORS", debug) // debug only
		log.Printf("%s: propagator fields: %v", me, fields)
	}

	otel.SetTextMapPropagator(prop)
}
