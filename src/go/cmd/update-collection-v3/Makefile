.PHONY: build
build: $(eval SHELL:=/bin/bash)
	for os_arch in linux_amd64 linux_arm64 darwin_arm64 darwin_amd64 windows_arm64 windows_amd64; do \
		os=`echo $$os_arch | awk -F_ '{print $$1}'`; \
		arch=`echo $$os_arch | awk -F_ '{print $$2}'`; \
		env GOOS=$$os GOARCH=$$arch go build -o update-collection-v3-sumo-$$os_arch; \
	done
test:
	go test ./...
