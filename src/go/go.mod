module github.com/SumoLogic/sumologic-kubernetes-collection/tools

go 1.14

require (
	github.com/aws/aws-sdk-go v1.34.5
	github.com/codahale/hdrhistogram v0.0.0-20161010025455-3a0bb77429bd // indirect
	github.com/opentracing/opentracing-go v1.2.0
	github.com/uber/jaeger-client-go v2.23.1+incompatible
	github.com/uber/jaeger-lib v2.2.0+incompatible // indirect
	go.opentelemetry.io/otel v0.14.0
	go.opentelemetry.io/otel/exporters/otlp v0.14.0
	go.opentelemetry.io/otel/sdk v0.14.0
	go.uber.org/atomic v1.7.0 // indirect
	google.golang.org/grpc v1.32.0
	k8s.io/apimachinery v0.17.12
	k8s.io/client-go v0.17.12
)

replace k8s.io/client-go => k8s.io/client-go v0.0.0-20190620085101-78d2af792bab
