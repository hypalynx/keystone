test:
	go test ./...

ensure-lint:
	@which golangci-lint > /dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

lint: ensure-lint
	@echo "Running golangci-lint..."
	@golangci-lint run --config=$(LINT_CONFIG) ./...

lint-fix: ensure-lint
	@echo "Running golangci-lint with auto-fix..."
	@golangci-lint run --config=$(LINT_CONFIG) --fix ./...
