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

	//
	// initialize tracing
	//

	var tracer trace.Tracer

	{
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

	for i := 0; i < 10; i++ {
		work(ctx, i, tracer)
	}
}

func work(ctx context.Context, i int, tracer trace.Tracer) {
	me := fmt.Sprintf("work %d", i)
	_, span := tracer.Start(ctx, me)
	defer span.End()
	log.Printf("%s: working", me)
	time.Sleep(200 * time.Millisecond)
}
