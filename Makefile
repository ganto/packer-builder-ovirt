TEST?=$(shell go list ./...)
VET?=$(shell go list ./...)

.PHONY: bin install-build-deps fmt fmt-check test

install-build-deps: ## Install dependencies for bin build
	@go get github.com/mitchellh/gox
	@go get ./...

bin: install-build-deps ## Build debug/test build
	@GO111MODULE=off sh -c "$(CURDIR)/scripts/build.sh"

fmt: ## Format Go code
	@go fmt ./...

fmt-check: ## Check go code formatting
	@echo "==> Checking that code complies with go fmt requirements..."
	@git diff --exit-code; if [ $$? -eq 1 ]; then \
		echo "Found files that are not fmt'ed."; \
		echo "You can use the command: \`make fmt\` to reformat code."; \
		exit 1; \
	fi

test: vet ## Run unit tests
	@go test $(TEST) $(TESTARGS) -timeout=3m

vet: ## Vet Go code
	@go vet $(VET); if [ $$? -eq 1 ]; then \
		echo "ERROR: Vet found problems in the code."; \
		exit 1; \
	fi

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
