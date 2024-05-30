package main

import (
	"context"
	"encoding/json"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// reference: https://github.com/wavefrontHQ/opentelemetry-examples/tree/master/go-example/manual-instrumentation
func initTracer() (*sdktrace.TracerProvider, error) {
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	// Set up a trace exporter
	traceExporter, err := otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithEndpoint(endpoint),
		),
	)
	reportErr(err, "failed to create trace exporter")

	// Register the trace exporter with a TracerProvider, using a batch
	// span processor to aggregate spans before export.
	batchSpanProcessor := sdktrace.NewBatchSpanProcessor(traceExporter)
	tp := newTraceProvider(resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceNameKey.String("childservice")), batchSpanProcessor)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return tp, nil
}

func newTraceProvider(res *resource.Resource, bsp sdktrace.SpanProcessor) *sdktrace.TracerProvider {
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	return tracerProvider
}

func reportErr(err error, message string) {
	if err != nil {
		log.Printf("%s: %v", message, err)
	}
}

func main() {
	tp, err := initTracer()
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			panic(err)
		}
	}()

	traceParent := os.Getenv("TRACE_PARENT")

	pro := otel.GetTextMapPropagator()
	carrier := make(map[string]string)
	if err := json.Unmarshal([]byte(traceParent), &carrier); err != nil {
		log.Fatal(err)
	}

	parentContext := pro.Extract(context.Background(), propagation.MapCarrier(carrier))

	// work begins
	parentFunction(parentContext)

	log.Info("Child process finished")
}

func parentFunction(parentContext context.Context) {
	ctx, parentSpan := otel.Tracer("childTracer").Start(parentContext, "childservice-span1")

	log.Printf("In parent span, before calling a child function.")
	defer parentSpan.End()
	time.Sleep(2 * time.Second)

	childFunction(ctx)

	log.Printf("In parent span, after calling a child function. When this function ends, parentSpan will complete.")
}

func childFunction(parentContext context.Context) {
	_, childSpan := otel.Tracer("childTracer2").Start(parentContext, "childservice-span2")
	defer childSpan.End()
	time.Sleep(2 * time.Second)

	log.Printf("In child span, when this function returns, childSpan will complete.")
}
