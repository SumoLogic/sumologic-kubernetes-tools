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
	"os"
	"strconv"
	"time"

	"google.golang.org/grpc"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

type stressTestConfig struct {
	// How many spans per minute are targeted
	spansPerMinute int
	// What should be the size of the trace
	spansPerTrace int
	// How many spans until the stress-test finishes
	totalSpans int

	// How many of the spans should be created right away (the rest will be created after a delay)
	spansCreatedImmediately int
	// What delay for the spans created later
	lateTraceDelay int
	// Each n-th trace will have the delay applied
	lateTraceFrequency int
	// OTC Collector Host Name
	collectorHostName string
	// Exporter - grpc or http
	exporter string
}

const (
	EnvSpansPerMin             = "SPANS_PER_MIN"
	EnvSpansPerTrace           = "SPANS_PER_TRACE"
	EnvTotalSpans              = "TOTAL_SPANS"
	EnvLateTraceDelayS         = "LATE_TRACE_DELAY_S"
	EnvLateTraceFrequency      = "LATE_TRACE_FREQ"
	EnvSpansCreatedImmediately = "LATE_TRACE_SPANS_CREATED_IMM"
	EnvCollectorHostName       = "COLLECTOR_HOSTNAME"
	EnvExporter                = "EXPORTER"
)

func (cfg *stressTestConfig) printConfig() {
	log.Printf("%s = %d\n", EnvSpansPerMin, cfg.spansPerMinute)
	log.Printf("%s = %d\n", EnvSpansPerTrace, cfg.spansPerTrace)
	log.Printf("%s = %d\n", EnvTotalSpans, cfg.totalSpans)
	log.Printf("%s = %d\n", EnvSpansCreatedImmediately, cfg.spansCreatedImmediately)
	log.Printf("%s = %d\n", EnvLateTraceDelayS, cfg.lateTraceDelay)
	log.Printf("%s = %d\n", EnvLateTraceFrequency, cfg.lateTraceFrequency)
	log.Printf("%s = %s\n", EnvCollectorHostName, cfg.collectorHostName)
	log.Printf("%s = %s\n", EnvExporter, cfg.exporter)
}

func createStressTestConfig() stressTestConfig {
	spm, err := strconv.Atoi(os.Getenv(EnvSpansPerMin))
	handleErr("SPANS_PER_MIN env variable not provided", err)

	spansPerTrace, err := strconv.Atoi(os.Getenv(EnvSpansPerTrace))
	if err != nil {
		spansPerTrace = 100
	}

	totalSpans, err := strconv.Atoi(os.Getenv(EnvTotalSpans))
	if err != nil {
		totalSpans = 10000000
	}

	lateTraceDelay, err := strconv.Atoi(os.Getenv(EnvLateTraceDelayS))
	if err != nil {
		lateTraceDelay = 8
	}

	lateTraceFrequency, err := strconv.Atoi(os.Getenv(EnvLateTraceFrequency))
	if err != nil {
		lateTraceFrequency = 20
	}

	spansCreatedImmediately, err := strconv.Atoi(os.Getenv(EnvSpansCreatedImmediately))
	if err != nil {
		spansCreatedImmediately = 50
	}

	collectorHostName := os.Getenv(EnvCollectorHostName)
	if collectorHostName == "" {
		collectorHostName = "collection-sumologic-otelagent.sumologic"
	}

	exporter := os.Getenv(EnvExporter)
	if exporter == "" {
		exporter = "http"
	}

	return stressTestConfig{
		spansPerMinute:          spm,
		spansPerTrace:           spansPerTrace,
		spansCreatedImmediately: spansCreatedImmediately,
		totalSpans:              totalSpans,
		lateTraceDelay:          lateTraceDelay,
		lateTraceFrequency:      lateTraceFrequency,
		collectorHostName:       collectorHostName,
		exporter:                exporter,
	}
}

type traceToFinishLater struct {
	rootSpan      *trace.Span
	toFinishSpans []*trace.Span
}

func (ttfl *traceToFinishLater) finishAll() {
	for _, s := range ttfl.toFinishSpans {
		(*s).End()
	}
	(*ttfl.rootSpan).End()
}

func (ttfl *traceToFinishLater) setMagicTag() {
	for _, s := range ttfl.toFinishSpans {
		(*s).SetAttributes(attribute.String("magicTag", "late"))
	}
}

func buildChildSpan(parentCtx context.Context, tracer trace.Tracer, countNumber int, magicValue int, magicTag *string) (context.Context, trace.Span) {
	ctx, childSpan := tracer.Start(
		parentCtx,
		fmt.Sprintf("ancestor-%d", countNumber+1),
		trace.WithAttributes(
			attribute.String("tagKey", "tagValue"),
			attribute.Int("countNumber", countNumber),
			attribute.Int("magicValue", magicValue),
		))
	if magicTag != nil {
		childSpan.SetAttributes(attribute.String("magicTag", *magicTag))
	}

	return ctx, childSpan
}

func buildTrace(ctx context.Context, tracer trace.Tracer, testCfg stressTestConfig, traceNumber int, isLate bool) traceToFinishLater {
	ctx, parentSpan := tracer.Start(ctx, "root-span")
	parentSpan.SetAttributes(attribute.Bool("late", isLate))

	toFinishSpans := make([]*trace.Span, 0)
	currentParent := &parentSpan

	for i := 0; i < testCfg.spansPerTrace-1; i++ {
		if i < testCfg.spansCreatedImmediately {
			(*currentParent).End()
		} else {
			toFinishSpans = append(toFinishSpans, currentParent)
		}

		var magicTag *string
		if traceNumber%11 == 0 {
			val := "true"
			magicTag = &val
		}
		_, childSpan := buildChildSpan(ctx, tracer, i, traceNumber%100, magicTag)
		childSpan.SetAttributes(attribute.Bool("late", isLate))
		currentParent = &childSpan
	}

	return traceToFinishLater{
		rootSpan:      &parentSpan,
		toFinishSpans: toFinishSpans,
	}
}

func runStressTest(testCfg stressTestConfig, tracer trace.Tracer) {
	totalCount := 0
	lateTracesSent := 0

	lateTraceDelayInSpans := testCfg.lateTraceDelay * testCfg.spansPerMinute / 60
	lateTraceDelayInTraces := lateTraceDelayInSpans / testCfg.spansPerTrace
	lateTraceDelayFinishQueueSize := lateTraceDelayInTraces / testCfg.lateTraceFrequency

	tracesCount := testCfg.totalSpans / testCfg.spansPerTrace
	start := time.Now()
	tracesToFinishLater := make([]traceToFinishLater, 0)
	ctx := context.Background()
	for i := 0; i < tracesCount; i++ {
		isLate := i%testCfg.lateTraceFrequency == 0
		trace := buildTrace(ctx, tracer, testCfg, i, isLate)
		if isLate {
			trace.setMagicTag()
			tracesToFinishLater = append(tracesToFinishLater, trace)
		} else {
			trace.finishAll()
		}

		if l := len(tracesToFinishLater); l > 0 && (l >= lateTraceDelayFinishQueueSize || i == tracesCount-1) {
			tracesToFinishLater[0].finishAll()
			tracesToFinishLater = tracesToFinishLater[1:]
			lateTracesSent += 1
		}

		totalCount += testCfg.spansPerTrace
		duration := time.Now().Sub(start)

		desiredDurationMicros := int64(float64(totalCount*60*1000*1000) / float64(testCfg.spansPerMinute))
		sleepDurationMicros := desiredDurationMicros - duration.Microseconds()
		if sleepDurationMicros > 0 {
			time.Sleep(time.Duration(sleepDurationMicros) * time.Microsecond)
		}

		if i%100 == 99 {
			// Calculate again to take sleep into account
			duration := time.Now().Sub(start)
			rpm := (60 * 1000 * 1000 * float64(totalCount)) / float64(duration.Microseconds())
			log.Printf("[Queue size: %d, Late traces queue: %d, Late traces sent: %d] ", sdktrace.DefaultMaxQueueSize, len(tracesToFinishLater), lateTracesSent)
			log.Printf("Created %d spans in %.3f seconds, or %.1f spans/minute\n", totalCount, float64(duration.Milliseconds())/1000.0, rpm)
		}
	}

	log.Println("Finishing late spans...")
	for _, tt := range tracesToFinishLater {
		tt.finishAll()
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

func initProvider(ctx context.Context, spanProcessor sdktrace.SpanProcessor) func() {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			// the service name used to display traces in backends
			semconv.ServiceNameKey.String("stress-tester"),
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

func configureOtlpHTTPExporter(ctx context.Context, collectorHostName string) sdktrace.SpanProcessor {
	endpoint := fmt.Sprintf("%s:4318", collectorHostName)
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

func main() {
	flag.Parse()
	if help {
		fmt.Println("Trace stress-testing")
		os.Exit(0)
	}

	testCfg := createStressTestConfig()
	testCfg.printConfig()

	var spanProcessor sdktrace.SpanProcessor
	if testCfg.exporter == "http" {
		spanProcessor = configureOtlpHTTPExporter(context.Background(), testCfg.collectorHostName)
	} else if testCfg.exporter == "grpc" {
		spanProcessor = configureOtlpGrpcExporter(context.Background(), testCfg.collectorHostName)
	} else {
		log.Fatalf(fmt.Sprintf("Unsupported exporter set %s", testCfg.exporter))
	}

	shutdown := initProvider(context.Background(), spanProcessor)
	defer shutdown()

	tracer := otel.Tracer(testCfg.exporter)
	runStressTest(testCfg, tracer)
}
