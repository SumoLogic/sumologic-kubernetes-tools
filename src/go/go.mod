module github.com/SumoLogic/sumologic-kubernetes-collection/tools

go 1.14

require (
	github.com/codahale/hdrhistogram v0.0.0-20161010025455-3a0bb77429bd // indirect
	github.com/opentracing/opentracing-go v1.2.0
	github.com/pkg/errors v0.9.1 // indirect
	github.com/uber/jaeger-client-go v2.23.1+incompatible
	github.com/uber/jaeger-lib v2.2.0+incompatible // indirect
	go.opentelemetry.io/otel v0.15.0
	go.opentelemetry.io/otel/exporters/otlp v0.14.0
	go.opentelemetry.io/otel/exporters/trace/zipkin v0.15.0
	go.opentelemetry.io/otel/sdk v0.15.0
	go.uber.org/atomic v1.7.0 // indirect
	golang.org/x/net v0.0.0-20200202094626-16171245cfb2 // indirect
	google.golang.org/grpc v1.32.0
	k8s.io/apimachinery v0.17.12
	k8s.io/client-go v0.17.12
)

replace k8s.io/client-go => k8s.io/client-go v0.0.0-20190620085101-78d2af792bab
