SOURCES := $(shell \
	find . -not \( \( -name .git -o -name .go -o -name vendor -o -name '*.pb.go' -o -name '*.pb.gw.go' -o -name '*_gen.go' -o -name '*mock*' \) -prune \) \
	-name '*.go')

.PHONY: help
help: Makefile ## Show list of commands.
	@echo "Choose a command to run in "$(APP_NAME)":"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-40s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort

# hidden targets (that will not be shown in help)

.PHONY: setup
setup: ## Download the dependencies to the project.
	@go get ./...
	@go mod download

#-------------------------------------------------------------------------------
#   Lint and tests
#-------------------------------------------------------------------------------

.PHONY: goimports
goimports: ## Execute goimports to standardize modules declaration and code.
	@go run golang.org/x/tools/cmd/goimports -w $(SOURCES)

.PHONY: lint
lint: lint/go lint/protobuf ## Execute linters.

.PHONY: lint/go
lint/go: ## Execute golangci-lint.
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint run


.PHONY: run/tests
run/tests: run/unit-tests run/integration-tests ## Execute all unit and integration tests

.PHONY: run/unit-tests
run/unit-tests: ## Execute unit tests.
	@go test -count=1 -tags=unit -coverprofile=coverage.out -covermode=atomic $(OPTS) ./...

.PHONY: run/integration-tests
run/integration-tests: ## Execute integration tests.
	@go test -tags=integration -count=1 -timeout 20m -coverprofile=coverage.out -covermode=atomic $(OPTS) ./...


#-------------------------------------------------------------------------------
#  Build and run
#-------------------------------------------------------------------------------

.PHONY: build
build: build-linux-x86_64 ## Build the project and generates a binary.
	@rm -f ./bin/udpproxy || true
	@go build -o ./bin/udpproxy ./

.PHONY: build-linux-x86_64
build-linux-x86_64: ## Build the project and generates a binary for x86_64 architecture.
	@rm -f ./bin/udpproxy-linux-x86_64 || true
	@env GOOS=linux GOARCH=amd64 go build -o ./bin/udpproxy-linux-x86_64 ./

