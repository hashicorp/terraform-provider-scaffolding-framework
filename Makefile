GO := go
TOOLS_GOMOD := -modfile=./tools/go.mod
GO_TOOL := $(GO) run $(TOOLS_GOMOD) -mod=mod

.PHONY: update
update:
	@echo "Updating submodules..."
	git pull --recurse-submodules
	git submodule update --remote --recursive

.PHONY: build
build:
	@echo "Building..."
	go build -v ./...

.PHONY: install
install: build
	@echo "Installing..."
	go install -v ./...

.PHONY: lint
lint:
	@echo "Linting..."
	$(GO_TOOL) github.com/golangci/golangci-lint/v2/cmd/golangci-lint run --verbose -c .golangci.yml

.PHONY: generate
generate:
	@echo "Generating documentation..."
	cd tools; go generate ./...

.PHONY: fmt
fmt:
	@echo "Formating..."
	$(GO_TOOL) mvdan.cc/gofumpt -w .

.PHONY: test
test:
	@echo "Running unit tests..."
	go test -v -cover -timeout=120s -parallel=10 ./...

.PHONY: testacc
testacc:
	@echo "Running acceptance tests..."
	TF_ACC=1 go test -v -cover -timeout 120m ./...
