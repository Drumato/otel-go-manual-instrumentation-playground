package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	exporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		panic(err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(newResource()),
		sdktrace.WithSyncer(exporter),
	)
	defer tp.Shutdown(ctx)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	tracer := otel.Tracer("chapter1")

	count := 0
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		default:
			ctx, span := tracer.Start(ctx, fmt.Sprintf("chapter1.%d", count))
			f1(ctx, tracer, count)
			span.End()

			count++
			time.Sleep(1 * time.Second)
		}
	}
}

func f1(ctx context.Context, tracer trace.Tracer, count int) {
	ctx, span := tracer.Start(ctx, fmt.Sprintf("chapter1.%d.f1", count))
	defer span.End()

	time.Sleep(1 * time.Second)
}

func newResource() *resource.Resource {
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("chapter1-main"),
	)
}
