.PHONY: build run test lint clean container

BINARY := apsystems-mcp
MAIN   := ./cmd/server

build:
	go build -ldflags="-s -w" -o $(BINARY) $(MAIN)

run: build
	APS_DASHBOARD=true APS_LOG_LEVEL=debug ./$(BINARY)

test:
	go test -v -race -coverprofile=coverage.out ./...

coverage: test
	go tool cover -html=coverage.out -o coverage.html
	@echo "Open coverage.html in your browser"

lint:
	golangci-lint run ./...

vet:
	go vet ./...

clean:
	rm -f $(BINARY) coverage.out coverage.html

container:
	podman build -t apsystems-mcp-server -f Containerfile .

container-run: container
	podman run --rm -it \
		--env-file .env.local \
		-e APS_DASH_ADDR=:$$(grep -m1 '^APS_DASH_ADDR_PORT=' .env.local | cut -d= -f2) \
		-p 8080:$$(grep -m1 '^APS_DASH_ADDR_PORT=' .env.local | cut -d= -f2) \
		apsystems-mcp-server

tidy:
	go mod tidy
	go mod verify

deps:
	go mod download
