PROJECT_NAME := pltf

.PHONY: build test fmt tidy clean

build:
	go build ./...

test:
	go test ./...

fmt:
	gofmt -w cmd pkg modules

tidy:
	go mod tidy

clean:
	rm -rf .pltf .terraform dist bin coverage.out
