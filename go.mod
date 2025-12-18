module github.com/udhos/otelconfig

go 1.24.11 // minimum

toolchain go1.25.5 // preferred

require (
	go.opentelemetry.io/contrib/propagators/autoprop v0.64.0
	go.opentelemetry.io/otel v1.39.0
	go.opentelemetry.io/otel/exporters/jaeger v1.17.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.39.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.39.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.39.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.39.0
	go.opentelemetry.io/otel/sdk v1.39.0
	go.opentelemetry.io/otel/trace v1.39.0
)

require (
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.3 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/propagators/aws v1.39.0 // indirect
	go.opentelemetry.io/contrib/propagators/b3 v1.39.0 // indirect
	go.opentelemetry.io/contrib/propagators/jaeger v1.39.0 // indirect
	go.opentelemetry.io/contrib/propagators/ot v1.39.0 // indirect
	go.opentelemetry.io/otel/metric v1.39.0 // indirect
	go.opentelemetry.io/proto/otlp v1.9.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20251213004720-97cd9d5aeac2 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251213004720-97cd9d5aeac2 // indirect
	google.golang.org/grpc v1.77.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)
