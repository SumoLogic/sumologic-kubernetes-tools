// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"google.golang.org/grpc"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

type traceTestConfig struct {
	// OTC Collector Host Name
	collectorHostName string
	// What should be the size of the trace
	spansPerTrace int
	// How many Traces should be generated
	totalTraces int
}

const (
	// EnvCollectorHostName OpenTelemetry Collector Hostname
	EnvCollectorHostName = "COLLECTOR_HOSTNAME"
	// EnvSpansPerTrace Number of spans generated per trace
	EnvSpansPerTrace = "SPANS_PER_TRACE"
	// EnvTotalTraces Number of traces generated per exporter
	EnvTotalTraces = "TOTAL_TRACES"
)

func (cfg *traceTestConfig) printConfig() {
	log.Printf("%s = %s\n", EnvCollectorHostName, cfg.collectorHostName)
	log.Printf("%s = %d\n", EnvTotalTraces, cfg.totalTraces)
	log.Printf("%s = %d\n", EnvSpansPerTrace, cfg.spansPerTrace)
}

func createTraceTestConfig() traceTestConfig {

	collectorHostName := os.Getenv(EnvCollectorHostName)
	if collectorHostName == "" {
		collectorHostName = "collection-sumologic-otelcol.sumologic"
	}

	spansPerTrace, err := strconv.Atoi(os.Getenv(EnvSpansPerTrace))
	if err != nil {
		spansPerTrace = 10
	}

	totalTraces, err := strconv.Atoi(os.Getenv(EnvTotalTraces))
	if err != nil {
		totalTraces = 1
	}

	return traceTestConfig{
		collectorHostName: collectorHostName,
		spansPerTrace:     spansPerTrace,
		totalTraces:       totalTraces,
	}
}

func configureOtlpGrpcExporter(ctx context.Context, collectorHostName string) sdktrace.SpanProcessor {
	endpoint := fmt.Sprintf("%s:4317", collectorHostName)
	log.Printf("OTLP gRPC Exporter endpoint: %s\n", endpoint)

	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithDialOption(grpc.WithBlock()),
	)
	handleErr("Failed to create OTLP gRPC exporter", err)

	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)

	return bsp
}

func configureOtlpHTTPExporter(ctx context.Context, collectorHostName string) sdktrace.SpanProcessor {
	endpoint := fmt.Sprintf("%s:55681", collectorHostName)
	log.Printf("OTLP HTTP Exporter endpoint: %s\n", endpoint)

	opts := []otlptracehttp.Option{
		otlptracehttp.WithInsecure(),
		otlptracehttp.WithEndpoint(endpoint),
	}
	client := otlptracehttp.NewClient(opts...)
	traceExporter, err := otlptrace.New(ctx, client)
	handleErr("Failed to create OTLP HTTP exporter", err)

	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)

	return bsp
}

func configureZipkinExporter(collectorHostName string) sdktrace.SpanProcessor {
	url := fmt.Sprintf("http://%s:9411/api/v2/spans", collectorHostName)
	log.Printf("Zipkin Exporter url: %s\n", url)

	traceExporter, err := zipkin.New(
		url,
		zipkin.WithSDKOptions(sdktrace.WithSampler(sdktrace.AlwaysSample())),
	)
	handleErr("Failed to create Zipkin trace exporter", err)
	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)

	return bsp
}

func configureJaegerThriftHTTPExporter(collectorHostName string) sdktrace.SpanProcessor {
	url := fmt.Sprintf("http://%s:14268/api/traces", collectorHostName)
	log.Printf("Jaeger Thrift HTTP Exporter url: %s\n", url)

	traceExporter, err := jaeger.New(
		jaeger.WithCollectorEndpoint(
			jaeger.WithEndpoint(url),
		),
	)
	handleErr("Failed to create Zipkin trace exporter", err)

	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)

	return bsp
}

func initProvider(ctx context.Context, spanProcessor sdktrace.SpanProcessor) func() {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			// the service name used to display traces in backends
			semconv.ServiceNameKey.String("customer-trace-test-service"),
		),
	)
	handleErr("failed to create resource", err)

	// Register the trace exporter with a TracerProvider, using a batch
	// span processor to aggregate spans before export.
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(spanProcessor),
	)
	otel.SetTracerProvider(tracerProvider)

	// set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return func() {
		// Shutdown will flush any remaining spans and shut down the exporter.
		handleErr("failed to shutdown TracerProvider", tracerProvider.Shutdown(ctx))
	}
}
func waitForAWhile() {
	max := 1000
	min := 10
	time.Sleep(time.Duration(rand.Intn(max-min+1)+min) * time.Millisecond)
}

func buildSpan(parentCtx context.Context, tracer trace.Tracer, countNumber int) (context.Context, trace.Span) {
	ctx, childSpan := tracer.Start(
		parentCtx,
		fmt.Sprintf("child-%d", countNumber+1),
		trace.WithAttributes(
			attribute.Int("countNumber", countNumber+1),
		),
	)

	childSpan.SetName(fmt.Sprintf("ancestor-%d", countNumber+1))
	return ctx, childSpan
}

func buildTrace(ctx context.Context, tracer trace.Tracer, spansPerTrace int, name string) {
	ctx, parentSpan := tracer.Start(ctx, "parent")
	parentSpan.SetName(fmt.Sprintf("root-span-%s", name))
	waitForAWhile()
	parentSpan.End()

	currentCtx := ctx
	for i := 0; i < spansPerTrace-1; i++ {
		newCtx, childSpan := buildSpan(currentCtx, tracer, i)
		currentCtx = newCtx
		waitForAWhile()
		childSpan.End()
	}
}

func runCustomerTraceTest(testCfg traceTestConfig, name string) {
	tracer := otel.Tracer(name)

	tracesCount := testCfg.totalTraces
	spansPerTrace := testCfg.spansPerTrace
	for i := 0; i < tracesCount; i++ {
		buildTrace(context.Background(), tracer, spansPerTrace, name)
	}
}

func handleErr(message string, err error) {
	if err != nil {
		log.Fatalf("%s: %v\n", message, err)
	}
}

var help bool

func init() {
	flag.BoolVar(&help, "help", false, "show help")
}

func main() {
	flag.Parse()
	if help {
		fmt.Println("Customer Trace Tester")
		fmt.Println("Simple application sending traces using various exporters")
		fmt.Println("Configuration environment variables:")
		fmt.Printf("- %s (default=collection-sumologic-otelcol.sumologic) - OT Collector hostname\n", EnvCollectorHostName)
		fmt.Printf("- %s (default=1) - Number of traces generated by exporter\n", EnvTotalTraces)
		fmt.Printf("- %s (default=10) - Number of spans per trace\n", EnvSpansPerTrace)
		os.Exit(0)
	}
	testCfg := createTraceTestConfig()

	otlpGrpcExporter := configureOtlpGrpcExporter(context.Background(), testCfg.collectorHostName)
	otlpHTTPExporter := configureOtlpHTTPExporter(context.Background(), testCfg.collectorHostName)
	zipkinExporter := configureZipkinExporter(testCfg.collectorHostName)
	jaegerThriftHTTPExporter := configureJaegerThriftHTTPExporter(testCfg.collectorHostName)
	spanProcessors := map[string]sdktrace.SpanProcessor{
		"otlpHttp":         otlpHTTPExporter,
		"otlpGrpc":         otlpGrpcExporter,
		"zipkin":           zipkinExporter,
		"jaegerThriftHttp": jaegerThriftHTTPExporter,
	}

	for name, spanProcessor := range spanProcessors {
		log.Printf("*******************************\n")
		log.Printf("Sending traces thru %s exporter\n", name)
		shutdown := initProvider(context.Background(), spanProcessor)
		defer shutdown()

		testCfg.printConfig()

		runCustomerTraceTest(testCfg, name)

	}

	log.Printf("*******************************\n")
	expectedNoOfTracesTotal := len(spanProcessors) * testCfg.totalTraces
	expectedNoOfSpansPerTrace := testCfg.totalTraces * testCfg.spansPerTrace
	expectedNoOfAllSpans := expectedNoOfTracesTotal * testCfg.spansPerTrace
	log.Printf("Expected number of all traces: %d\n", expectedNoOfTracesTotal)
	log.Printf("Expected number of spans in single trace: %d\n", expectedNoOfSpansPerTrace)
	log.Printf("Expected number of spans for all traces: %d\n", expectedNoOfAllSpans)
}
