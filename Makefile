default: all

build:
	@echo "Building..."
	@go build -o bin/ ./...

all: build
