.DEFAULT_GOAL := help
.DELETE_ON_ERROR:

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@printf 'Usage: make <target>\n\n'
	@awk 'BEGIN {FS = ": "} /^## / {printf "  %-16s %s\n", substr($$1, 4), $$2}' $(MAKEFILE_LIST)

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## run/api: run the API application with optional ARGS
.PHONY: run/api
run/api:
	go run ./cmd/api $(ARGS)

## tidy: tidy, verify, and vendor dependencies
.PHONY: tidy
tidy:
	go mod tidy
	go mod verify
	go mod vendor

## fmt: format Go source files
.PHONY: fmt
fmt:
	go fmt ./...

# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## test: run all tests
.PHONY: test
test:
	go test ./...

## test/coverage: generate text and HTML coverage reports
.PHONY: test/coverage
test/coverage:
	mkdir -p ./coverage
	go test -coverpkg=./... -covermode=atomic -coverprofile=./coverage/coverage.out ./...
	go tool cover -func=./coverage/coverage.out
	go tool cover -html=./coverage/coverage.out -o=./coverage/coverage.html

## audit: check dependencies, vet code, gosec, govulncheck, and run tests with race detection
.PHONY: audit
audit:
	go mod tidy -diff
	go mod verify
	go vet ./...
	go tool staticcheck ./...
	gosec ./...
	govulncheck ./...
	go test -race -vet=off ./...

# ==================================================================================== #
# BUILD
# ==================================================================================== #

## build/api: build the cmd/api application
.PHONY: build/api
build/api:
	go build -ldflags='-s' -o=./bin/api ./cmd/api
	GOOS=linux GOARCH=amd64 go build -ldflags='-s' -o=./bin/linux_amd64/api ./cmd/api
