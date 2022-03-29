build:
	go build \
		-ldflags '-w -extldflags "-static"' \
		-o k8s-api-test cmd/k8s-api-test/main.go

	go build \
		-ldflags '-w -extldflags "-static"' \
		-o stress-tester cmd/stress-tester/main.go

	go build \
		-ldflags '-w -extldflags "-static"' \
		-o k8s-api-test cmd/customer-trace-tester/main.go
