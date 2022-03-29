build:
	go build \
		-ldflags '-w -extldflags "-static"' \
		-o k8s-api-test cmd/k8s-api-test/main.go

	go build \
		-ldflags '-w -extldflags "-static"' \
		-o stress-tester cmd/stress-tester/main.go

	go build \
		-ldflags '-w -extldflags "-static"' \
		-o customer-trace-tester cmd/customer-trace-tester/main.go

	go build \
		-ldflags '-w -extldflags "-static"' \
		-o update-collection-v3 ./cmd/update-collection-v3/
