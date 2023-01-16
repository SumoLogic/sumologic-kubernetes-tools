.PHONY: build
build:
	env GOOS=windows GOARCH=arm64 go build -o collection-sumo-windows_arm64
	env GOOS=windows GOARCH=amd64 go build -o collection-sumo-windows_amd64
	env GOOS=linux GOARCH=arm64 go build -o collection-sumo-linux_arm64
	env GOOS=linux GOARCH=amd64 go build -o collection-sumo-linux_amd64
	env GOOS=darwin GOARCH=arm64 go build -o collection-sumo-darwin_arm64
	env GOOS=darwin GOARCH=amd64 go build -o collection-sumo-darwin_amd64
test:
	go test ./...
