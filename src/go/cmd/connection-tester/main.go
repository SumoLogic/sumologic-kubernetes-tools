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
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/trace/zipkin"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv"
	"go.opentelemetry.io/otel/trace"
)

type TestConfig struct {
	// Endpoint Address
	Address string
	// Auth Token
	Token string
	// How many spans per minute are targeted
	SpansPerMinute int
	// What should be the size of the trace
	SpansPerTrace int
	// How many spans until the stress-test finishes
	TotalSpans int
	//service name
	ServiceName string
}

type Config struct {
	SpansPerTrace int `yaml:"spansPerTrace"`
	SpansPerMinute int `yaml:"spansPerMinute"`
	Address string `yaml:"address"`
	Tokens []Token `yaml:"tokens"`
}

type Token struct {
	Token string `yaml:"token"`
	ServiceName string `yaml:"serviceName"`
	TotalSpans int `yaml:"totalSpans"`
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

func (cfg *TestConfig) printConfig() {
	log.Printf("%s = %s\n", EnvAddress, cfg.Address)
	log.Printf("%s = %s\n", EnvToken, cfg.Token)
	log.Printf("%s = %d\n", EnvSpansPerMin, cfg.SpansPerMinute)
	log.Printf("%s = %d\n", EnvSpansPerTrace, cfg.SpansPerTrace)
	log.Printf("%s = %d\n", EnvTotalSpans, cfg.TotalSpans)
}

func readConfigFromFile(configPath string) (*Config, error) {
	config := &Config{}
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}
	return config, nil
}

func createConfigs(config *Config) []TestConfig {
	var configs []TestConfig
	for i:=0; i<len(config.Tokens); i++ {
		testConfig := TestConfig{
			Address:        config.Address,
			Token:          config.Tokens[i].Token,
			SpansPerMinute: config.SpansPerMinute,
			SpansPerTrace:  config.SpansPerTrace,
			TotalSpans:     config.Tokens[i].TotalSpans,
			ServiceName: 	config.Tokens[i].ServiceName,
		}
		configs = append(configs, testConfig)
	}
	return configs
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

func buildTrace(ctx context.Context, tracer trace.Tracer, testConfig TestConfig, traceNumber int) {
	traceCtx, parentSpan := tracer.Start(
		ctx,
		"parent",
		trace.WithNewRoot(),
		trace.WithAttributes(label.String("foo", "bar")))

	currentParent := parentSpan
	currentCtx := traceCtx

	for i := 0; i < testConfig.SpansPerTrace-1; i++ {
		currentParent.End()

		newCtx, childSpan := buildSpan(tracer, currentCtx, i, traceNumber%100)
		currentParent = childSpan
		currentCtx = newCtx
	}
	currentParent.End()
}

func runIteration(ctx context.Context, bsp *sdktrace.BatchSpanProcessor, testCfg TestConfig, i int) int {
	tracer := otel.Tracer("connection-test-tracer")
	buildTrace(ctx, tracer, testCfg, i)
	bsp.ForceFlush()

	return testCfg.SpansPerTrace
}

func runTest(wg *sync.WaitGroup, testCfg TestConfig) {
	ctx := context.Background()
	defer wg.Done()
	bsp, shutdown := initTracing(ctx, testCfg)
	defer shutdown()

	totalCount := 0

	tracesCount := testCfg.TotalSpans / testCfg.SpansPerTrace
	start := time.Now()
	for i := 0; i < tracesCount; i++ {
		totalCount += runIteration(ctx, bsp, testCfg, i)
		log.Println(testCfg.ServiceName + " %d", totalCount)

		duration := time.Now().Sub(start)

		desiredDurationMicros := int64(float64(totalCount*60*1000*1000) / float64(testCfg.SpansPerMinute))
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

func initTracing(ctx context.Context, testCfg TestConfig) (*sdktrace.BatchSpanProcessor, func()) {
	headers := map[string]string{}
	if testCfg.Token != "" {
		headers[TokenKey] = testCfg.Token
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			// the service name used to display traces in backends
			semconv.ServiceNameKey.String(testCfg.ServiceName),
		),
	)
	handleErr("Could not create resource", err)

	exp, err := zipkin.NewRawExporter(testCfg.Address+ "/" + testCfg.Token, testCfg.ServiceName)
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
		fmt.Println("Trace connection-testing")
		os.Exit(0)
	}

	config, err := readConfigFromFile("./cmd/connection-tester/config.yml")
	handleErr("Could not read config file", err)
	configs := createConfigs(config)

	var waitGroup sync.WaitGroup
	for i:=0; i<len(configs); i++ {
		configs[i].printConfig()
		waitGroup.Add(1)
		go runTest(&waitGroup, configs[i])
	}
	fmt.Println("Main: Waiting for workers to finish")
	waitGroup.Wait()
	fmt.Println("Main: Completed")
}
