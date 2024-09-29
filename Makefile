default: all

build:
	@echo "Building..."
	@go build -gcflags="" -o bin/ ./...


cross-build:
	@echo "Cross-building..."
	@GOOS=linux GOARCH=amd64 go build -o bin/dedust-linux main.go

all: build
