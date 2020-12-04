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
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	"log"
	"time"
)


func buildSpan(tracer opentracing.Tracer, ref ...opentracing.StartSpanOption) *opentracing.Span {
	childSpan := tracer.StartSpan(
		"child-operation",
		ref...,
	)
	//defer childSpan.Finish()

	childSpan.SetTag("span.kind", "client")
	time.Sleep(100 * time.Millisecond)

	return &childSpan
}

func buildTrace(tracer1 opentracing.Tracer, tracer2 opentracing.Tracer) {
	parentSpan := tracer1.StartSpan("parent")
	parentSpan.SetTag("span.kind", "server")

	parentSpan.SetOperationName("root-span")
	parentSpan.Finish()

	time.Sleep(50 * time.Millisecond)
	//parentSpan.Finish()

	child1 := buildSpan(tracer1, opentracing.ChildOf(parentSpan.Context()))
	(*child1).Finish()

	time.Sleep(5 * time.Millisecond)

	child2 := buildSpan(tracer2, opentracing.ChildOf((*child1).Context()), opentracing.FollowsFrom((*child1).Context()))

	time.Sleep(5 * time.Millisecond)

	child3 := buildSpan(tracer2, opentracing.ChildOf((*child2).Context()))
	(*child3).Finish()

	time.Sleep(5 * time.Millisecond)

	child3a := buildSpan(tracer2, opentracing.ChildOf((*child2).Context()))
	(*child3a).Finish()

	time.Sleep(5 * time.Millisecond)

	child3b := buildSpan(tracer2, opentracing.ChildOf((*child2).Context()))
	(*child3b).Finish()

	(*child2).Finish()


	time.Sleep(10 * time.Millisecond)

	child4 := buildSpan(tracer1, opentracing.ChildOf((*child2).Context()), opentracing.FollowsFrom((*child2).Context()))
	(*child4).Finish()

	child5 := buildSpan(tracer1, opentracing.ChildOf((*child4).Context()))
	(*child5).Finish()
}

func handleErr(message string, err error) {
	if err != nil {
		log.Fatalf("%s: %v\n", message, err)
	}
}

func main() {
	jagerCfg1, err := jaegercfg.FromEnv()
	handleErr("Could not parse Jaeger env", err)
	jagerCfg2, err := jaegercfg.FromEnv()
	handleErr("Could not parse Jaeger env", err)

	jagerCfg1.Sampler = &jaegercfg.SamplerConfig{Type: jaeger.SamplerTypeConst, Param: 1.0}
	jagerCfg1.ServiceName = "parent_service"
	jagerCfg2.Sampler = &jaegercfg.SamplerConfig{Type: jaeger.SamplerTypeConst, Param: 1.0}
	jagerCfg2.ServiceName = "child_service"

	tracer1, closer1, _ := jagerCfg1.NewTracer()
	tracer2, closer2, _ := jagerCfg2.NewTracer()

	defer closer1.Close()
	defer closer2.Close()

	buildTrace(tracer1, tracer2)
}
