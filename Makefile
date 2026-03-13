.PHONY: help
help:
	@awk -f help.awk $(MAKEFILE_LIST)
	@echo ""

.DEFAULT_GOAL := help

# === CI/CD ===

.PHONY: lint
lint: ## Запустить линтер
	golangci-lint run

.PHONY: test
test: ## Запустить тесты с покрытием
	go test -race -coverprofile=coverage.out -covermode=atomic \
		$(shell go list ./... | grep -v -E '/(vendor|cmd|testdata|mocks)/')
	sed -i -E '/[_.]mock\.go:|\.pb\.go:|\.gen\.go:/d' coverage.out || true
	go tool cover -func=coverage.out | tee coverage.humanize | tail -1