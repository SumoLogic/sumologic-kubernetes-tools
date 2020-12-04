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

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

type stressTestConfig struct {
	// Endpoint address
	address string
	// Auth token
	token string
	// How many spans per minute are targeted
	spansPerMinute int
	// What should be the size of the trace
	spansPerTrace int
	// How many spans until the stress-test finishes
	totalSpans int
}

const (
	DefaultAddress             = "localhost:55680"
	TokenKey                   = "Auth-Token"
	EnvAddress                 = "OTLP_ENDPOINT"
	EnvToken                   = "AUTH_TOKEN"
	EnvSpansPerMin             = "SPANS_PER_MIN"
	EnvSpansPerTrace           = "SPANS_PER_TRACE"
	EnvTotalSpans              = "TOTAL_SPANS"
)

func (cfg *stressTestConfig) printConfig() {
	log.Printf("%s = %s\n", EnvAddress, cfg.address)
	log.Printf("%s = %s\n", EnvToken, cfg.token)
	log.Printf("%s = %d\n", EnvSpansPerMin, cfg.spansPerMinute)
	log.Printf("%s = %d\n", EnvSpansPerTrace, cfg.spansPerTrace)
	log.Printf("%s = %d\n", EnvTotalSpans, cfg.totalSpans)
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

	address := os.Getenv(EnvAddress)
	if address == "" {
		address = DefaultAddress
	}

	token := os.Getenv(EnvToken)

	return stressTestConfig{
		address:                 address,
		token:                   token,
		spansPerMinute:          spm,
		spansPerTrace:           spansPerTrace,
		totalSpans:              totalSpans,
	}
}

func buildSpan(tracer trace.Tracer, parentCtx context.Context, countNumber int, magicValue int) (context.Context, trace.Span) {
	ctx, childSpan := tracer.Start(
		parentCtx,
		fmt.Sprintf("ancestor-%d", countNumber+1),
		trace.WithAttributes(
			label.String("tagKey", "tagValue"),
			label.Int("countNumber", countNumber),
			label.Int("magicValue", magicValue)),
	)

	return ctx, childSpan
}

func buildTrace(tracer trace.Tracer, testConfig stressTestConfig, traceNumber int) {
	ctx, parentSpan := tracer.Start(
		context.Background(),
		"parent",
		trace.WithNewRoot(),
		trace.WithAttributes(label.String("foo", "bar")))

	currentParent := parentSpan
	currentCtx := ctx

	for i := 0; i < testConfig.spansPerTrace-1; i++ {
		currentParent.End()

		newCtx, childSpan := buildSpan(tracer, currentCtx, i, traceNumber%100)
		currentParent = childSpan
		currentCtx = newCtx
	}
	currentParent.End()
}

func runIteration(bsp *sdktrace.BatchSpanProcessor, testCfg stressTestConfig, i int) int {
	tracer := otel.Tracer("stress-test-tracer")
	buildTrace(tracer, testCfg, i)
	bsp.ForceFlush()

	return testCfg.spansPerTrace
}

func runStressTest(testCfg stressTestConfig) {
	bsp, shutdown := initTracing(testCfg)
	defer shutdown()

	totalCount := 0

	tracesCount := testCfg.totalSpans / testCfg.spansPerTrace
	start := time.Now()
	for i := 0; i < tracesCount; i++ {
		totalCount += runIteration(bsp, testCfg, i)

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
			log.Printf("Created %d spans in %.3f seconds, or %.1f spans/minute\n", totalCount, float64(duration.Milliseconds())/1000.0, rpm)
		}
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

func initTracing(testCfg stressTestConfig) (*sdktrace.BatchSpanProcessor, func()) {
	ctx := context.Background()

	headers := map[string]string{}
	if testCfg.token != "" {
		headers[TokenKey] = testCfg.token
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			// the service name used to display traces in backends
			semconv.ServiceNameKey.String("stress-tester"),
		),
	)
	handleErr("Could not create resource", err)

	exp, err := otlp.NewExporter(otlp.WithInsecure(),
		otlp.WithHeaders(headers),
		otlp.WithAddress(testCfg.address),
		otlp.WithGRPCDialOption(grpc.WithBlock()), // useful for testing
	)

	handleErr("Could not create exporter", err)

	// No sampling
	bsp := sdktrace.NewBatchSpanProcessor(exp)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithConfig(sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	// Set global propagator to tracecontext (the default is no-op).
	// Doesn't matter here much but good to keep it like if the code is reused
	otel.SetTextMapPropagator(propagation.TraceContext{})
	otel.SetTracerProvider(tracerProvider)

	return bsp, func() {
		handleErr("Failed to shutdown provider", tracerProvider.Shutdown(ctx))
		handleErr("Failed to stop exporter", exp.Shutdown(ctx))
	}
}

func main() {
	flag.Parse()
	if help {
		fmt.Println("Trace stress-testing")
		os.Exit(0)
	}

	testCfg := createStressTestConfig()
	testCfg.printConfig()

	runStressTest(testCfg)
}
