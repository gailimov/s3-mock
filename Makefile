IMAGE_NAME ?= s3-mock
GOLANGCI_LINT_VERSION ?= v2.12.2

GOLANGCI_LINT := @docker run \
	--rm \
	-t \
	-v $(CURDIR):/app -w /app \
	--user $(shell id -u):$(shell id -g) \
	-v $(shell go env GOCACHE):/.cache/go-build -e GOCACHE=/.cache/go-build \
	-v $(shell go env GOMODCACHE):/.cache/mod -e GOMODCACHE=/.cache/mod \
	-v $(HOME)/.cache/golangci-lint:/.cache/golangci-lint -e GOLANGCI_LINT_CACHE=/.cache/golangci-lint \
	golangci/golangci-lint:$(GOLANGCI_LINT_VERSION) golangci-lint -v

build:
	docker build -t $(IMAGE_NAME) .

run:
	docker run \
		--rm \
		-p 8080:8080 \
		$(IMAGE_NAME)

check: lint-dockerfile lint test

lint-dockerfile:
	docker run --rm -i hadolint/hadolint < ./Dockerfile

lint:
	$(GOLANGCI_LINT) run

lint-fix:
	$(GOLANGCI_LINT) run --fix

fmt:
	$(GOLANGCI_LINT) fmt

fmt-check:
	$(GOLANGCI_LINT) fmt --diff-colored

test:
	go test -v ./...

coverage-html:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
