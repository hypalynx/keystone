.PHONY: test clean lint release patch minor major help

VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
NEXT_PATCH = $(shell echo $(VERSION) | awk -F. '{ printf("v%d.%d.%d", $$1, $$2, $$3+1) }' | sed 's/vv/v/g')
NEXT_MINOR = $(shell echo $(VERSION) | awk -F. '{ printf("v%d.%d.%d", $$1, $$2+1, 0) }' | sed 's/vv/v/g')
NEXT_MAJOR = $(shell echo $(VERSION) | awk -F. '{ printf("v%d.%d.%d", $$1+1, 0, 0) }' | sed 's/vv/v/g')

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

clean:
	rm -rf ./dist
	go clean

version:
	@echo "Current version: $(VERSION)"
	@echo "Next patch: $(NEXT_PATCH)"
	@echo "Next minor: $(NEXT_MINOR)"
	@echo "Next major: $(NEXT_MAJOR)"

# Release helpers
patch:
	@$(MAKE) release VERSION=$(NEXT_PATCH)

minor:
	@$(MAKE) release VERSION=$(NEXT_MINOR)

major:
	@$(MAKE) release VERSION=$(NEXT_MAJOR)

release:
	@echo "Preparing release $(VERSION)..."
	@echo "Running tests..."
	@$(MAKE) test || (echo "Tests failed. Release canceled."; exit 1)
	@echo "Creating git tag $(VERSION)..."
	@git tag -a $(VERSION) -m "Release $(VERSION)"
	@echo "Pushing tag to origin..."
	@git push origin $(VERSION)
	@echo "Release $(VERSION) completed successfully!"
	@echo "Don't forget to update the changelog!"
