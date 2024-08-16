default: fmt lint generate build install    

build:
	go build -v ./...

install: build
	go install -v ./...

lint:
	golangci-lint run

# Generate docs and copywrite headers
generate:
	cd tools; go generate ./...

fmt:
	gofmt -s -w -e .

test:
	go test -v -cover -timeout=120s -parallel=10 ./...

testacc:
	TF_ACC=1 go test -v -cover -timeout 120m ./...

.PHONY: build install lint generate fmt test testacc
