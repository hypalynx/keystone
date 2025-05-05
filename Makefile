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

cover:
	go test -coverprofile=coverage.out ./... >/dev/null || true
	(go tool cover -func=coverage.out | grep -v "total:" | sort -k3 -nr && go tool cover -func=coverage.out | grep "total:")
