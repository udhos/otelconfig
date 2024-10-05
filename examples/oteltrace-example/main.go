// Package main implements the program.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/udhos/otelconfig/oteltrace"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	me := filepath.Base(os.Args[0])

	noop := envBool("NOOP", false) // NOOP env var disables tracing
	interval := envDuration("INTERVAL", 200*time.Millisecond)
	repeat := envInt("REPEAT", 10)

	//
	// initialize tracing
	//

	var tracer trace.Tracer

	if noop {
		tracer = oteltrace.NewNoopTracer()
	} else {
		options := oteltrace.TraceOptions{
			DefaultService:     me,
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

	//
	// do the work, create spans to record it
	//

	ctx, span := tracer.Start(context.TODO(), "main")
	defer span.End()

	for i := 0; i < repeat; i++ {
		work(ctx, i+1, repeat, tracer, interval)
	}
}

func work(ctx context.Context, i, repeat int, tracer trace.Tracer, interval time.Duration) {
	me := fmt.Sprintf("work %d/%d", i, repeat)
	_, span := tracer.Start(ctx, me)
	defer span.End()
	log.Printf("%s: working", me)
	time.Sleep(interval)
}
