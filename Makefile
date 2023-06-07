.DEFAULT_GOAL := default

.PHONY: run
run:
	go run cmd/scan/main.go

.PHONY: test
test:
	gotestsum ./...

.PHONY: test-no-cache
test-no-cache:
	gotestsum ./... -count=1

.PHONY: test-verbose
test-verbose:
	gotestsum --format standard-verbose ./...

.PHONY: bench
bench:
	go test -bench=. ./internal/service -count 1 -run=^#

.PHONY: coverage
coverage:
	gotestsum ./... -test.coverprofile coverage.out -test.v

.PHONY: install-gotestsum
install-gotestsum:
	go install gotest.tools/gotestsum@latest

.PHONY: install-linter
install-linter:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.42.1

.PHONY: lint
lint:
	golangci-lint run

# static check comes with VSCode's Go extension as one of the tools
# will look into adding an install step for it when I get more time
.PHONY: staticcheck
staticcheck:
	staticcheck ./...
