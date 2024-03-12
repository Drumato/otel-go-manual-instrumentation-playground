package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
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

loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		default:
			client := http.Client{
				Transport: otelhttp.NewTransport(http.DefaultTransport),
			}
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/", nil)
			if err != nil {
				fmt.Printf("failed to create request: %v\n", err)
				continue
			}

			resp, err := client.Do(req)
			if err != nil {
				fmt.Printf("request failed: %v\n", err)
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				fmt.Printf("got status: %v\n", resp.Status)
				continue
			}

			fmt.Printf("%v\n", resp.Status)
			time.Sleep(1 * time.Second)
		}
	}
}

func newResource() *resource.Resource {
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("chapter2-client"),
	)
}
