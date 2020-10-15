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
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
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
}

const (
	EnvSpansPerMin             = "SPANS_PER_MIN"
	EnvSpansPerTrace           = "SPANS_PER_TRACE"
	EnvTotalSpans              = "TOTAL_SPANS"
	EnvLateTraceDelayS         = "LATE_TRACE_DELAY_S"
	EnvLateTraceFrequency      = "LATE_TRACE_FREQ"
	EnvSpansCreatedImmediately = "LATE_TRACE_SPANS_CREATED_IMM"
)

func (cfg *stressTestConfig) printConfig() {
	log.Printf("%s = %d\n", EnvSpansPerMin, cfg.spansPerMinute)
	log.Printf("%s = %d\n", EnvSpansPerTrace, cfg.spansPerTrace)
	log.Printf("%s = %d\n", EnvTotalSpans, cfg.totalSpans)
	log.Printf("%s = %d\n", EnvSpansCreatedImmediately, cfg.spansCreatedImmediately)
	log.Printf("%s = %d\n", EnvLateTraceDelayS, cfg.lateTraceDelay)
	log.Printf("%s = %d\n", EnvLateTraceFrequency, cfg.lateTraceFrequency)
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

	return stressTestConfig{
		spansPerMinute:          spm,
		spansPerTrace:           spansPerTrace,
		spansCreatedImmediately: spansCreatedImmediately,
		totalSpans:              totalSpans,
		lateTraceDelay:          lateTraceDelay,
		lateTraceFrequency:      lateTraceFrequency,
	}
}

type traceToFinishLater struct {
	rootSpan      *opentracing.Span
	toFinishSpans []*opentracing.Span
}

func (ttfl *traceToFinishLater) finishAll() {
	for _, s := range ttfl.toFinishSpans {
		(*s).Finish()
	}
	(*ttfl.rootSpan).Finish()
}

func (ttfl *traceToFinishLater) setMagicTag() {
	for _, s := range ttfl.toFinishSpans {
		(*s).SetTag("magicTag", "late")
	}
}

func buildSpan(tracer opentracing.Tracer, parentSpan *opentracing.Span, countNumber int, magicValue int, magicTag *string) opentracing.Span {
	childSpan := tracer.StartSpan(
		"child",
		opentracing.ChildOf((*parentSpan).Context()),
	)

	childSpan.SetBaggageItem("baggageKey", "baggageValue")
	childSpan.SetTag("tagKey", "tagValue")
	childSpan.SetTag("countNumber", countNumber)
	childSpan.SetTag("magicValue", magicValue)
	if magicTag != nil {
		childSpan.SetTag("magicTag", *magicTag)
	}
	childSpan.SetOperationName(fmt.Sprintf("ancestor-%d", countNumber+1))

	return childSpan
}

func buildTrace(testConfig stressTestConfig, traceNumber int, isLate bool) traceToFinishLater {
	tracer := opentracing.GlobalTracer()
	parentSpan := tracer.StartSpan("parent")
	parentSpan.SetOperationName("root-span")
	parentSpan.SetTag("late", isLate)

	toFinishSpans := make([]*opentracing.Span, 0)
	currentParent := &parentSpan

	for i := 0; i < testConfig.spansPerTrace-1; i++ {
		if i < testConfig.spansCreatedImmediately {
			(*currentParent).Finish()
		} else {
			toFinishSpans = append(toFinishSpans, currentParent)
		}

		var magicTag *string
		if traceNumber%11 == 0 {
			val := "true"
			magicTag = &val
		}
		childSpan := buildSpan(tracer, currentParent, i, traceNumber%100, magicTag)
		childSpan.SetTag("late", isLate)
		currentParent = &childSpan
	}

	return traceToFinishLater{
		rootSpan:      &parentSpan,
		toFinishSpans: toFinishSpans,
	}
}

func runStressTest(testCfg stressTestConfig, jagerCfg *jaegercfg.Configuration) {
	totalCount := 0
	lateTracesSent := 0

	lateTraceDelayInSpans := testCfg.lateTraceDelay * testCfg.spansPerMinute / 60
	lateTraceDelayInTraces := lateTraceDelayInSpans / testCfg.spansPerTrace
	lateTraceDelayFinishQueueSize := lateTraceDelayInTraces / testCfg.lateTraceFrequency

	tracesCount := testCfg.totalSpans / testCfg.spansPerTrace
	start := time.Now()
	tracesToFinishLater := make([]traceToFinishLater, 0)
	for i := 0; i < tracesCount; i++ {
		isLate := i%testCfg.lateTraceFrequency == 0
		trace := buildTrace(testCfg, i, isLate)
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
			log.Printf("[Queue size: %d, Late traces queue: %d, Late traces sent: %d] ", jagerCfg.Reporter.QueueSize, len(tracesToFinishLater), lateTracesSent)
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

func main() {
	flag.Parse()
	if help {
		fmt.Println("Trace stress-testing")
		os.Exit(0)
	}

	jagerCfg, err := jaegercfg.FromEnv()
	handleErr("Could not parse Jaeger env", err)

	jagerCfg.Sampler = &jaegercfg.SamplerConfig{Type: jaeger.SamplerTypeConst, Param: 1.0}
	jagerCfg.ServiceName = "jaeger_stress_tester"
	log.Printf("Sampler type: %s param: %.3f\n", jagerCfg.Sampler.Type, jagerCfg.Sampler.Param)

	tracer, closer, err := jagerCfg.NewTracer()
	handleErr("Could not initialize tracer", err)
	log.Printf("CollectorEndpoint: %s\n", jagerCfg.Reporter.CollectorEndpoint)
	log.Printf("LocalAgentHostPort: %s\n", jagerCfg.Reporter.LocalAgentHostPort)
	defer closer.Close()

	opentracing.SetGlobalTracer(tracer)

	testCfg := createStressTestConfig()
	testCfg.printConfig()

	runStressTest(testCfg, jagerCfg)
}
