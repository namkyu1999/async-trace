package main

import (
	"context"
	"runtime"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/kelseyhightower/envconfig"
	"github.com/namkyu1999/async-trace/parent/pkg/handlers"
	"github.com/namkyu1999/async-trace/parent/pkg/utils"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func init() {
	err := envconfig.Process("", &utils.Config)
	log.SetReportCaller(true)

	log.Infof("go version: %s", runtime.Version())
	log.Infof("go os/arch: %s/%s", runtime.GOOS, runtime.GOARCH)

	if err != nil {
		log.Fatal(err)
	}
}

func initTracer() func() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	res, err := newResource(ctx)
	reportErr(err, "failed to create res")

	// Set up a trace exporter
	traceExporter, err := otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithEndpoint(utils.Config.OtelExporterOtlpEndpoint),
		),
	)
	reportErr(err, "failed to create trace exporter")

	// Register the trace exporter with a TracerProvider, using a batch
	// span processor to aggregate spans before export.
	batchSpanProcessor := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := newTraceProvider(res, batchSpanProcessor)
	otel.SetTracerProvider(tracerProvider)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return func() {
		// Shutdown will flush any remaining spans and shut down the exporter.
		reportErr(tracerProvider.Shutdown(ctx), "failed to shutdown TracerProvider")
		cancel()
	}
}

func newTraceProvider(res *resource.Resource, bsp sdktrace.SpanProcessor) *sdktrace.TracerProvider {
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	return tracerProvider
}

func newResource(ctx context.Context) (*resource.Resource, error) {
	return resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(utils.Config.OtelServiceName),
		),
	)
}

func reportErr(err error, message string) {
	if err != nil {
		log.Printf("%s: %v", message, err)
	}
}

func main() {
	// setup tracing
	shutdown := initTracer()
	defer shutdown()

	tracer := otel.Tracer("parentTracer")

	// setup gin
	gin.SetMode(gin.ReleaseMode)
	gin.EnableJsonDecoderDisallowUnknownFields()
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowHeaders:     []string{"*"},
		AllowMethods:     []string{"PUT", "GET", "POST", "DELETE"},
		AllowCredentials: true,
	}))

	router.GET("/status", handlers.StatusHandler)
	router.GET("/hello", handlers.GetLogicHandler(tracer))
	router.GET("/child-process", handlers.CreateChildProcessHandler(tracer))

	err := router.Run(":" + utils.Config.ParentServicePort)
	if err != nil {
		log.Fatal(err)
	}
}
