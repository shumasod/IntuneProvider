DEFAULT_GOAL := build

BINARY      := terraform-provider-intune
NAMESPACE   := shumasod
NAME        := intune
VERSION     := 0.1.0
OS_ARCH     := $(shell go env GOOS)_$(shell go env GOARCH)

OS_ARCH_BIN  := $(shell go env GOOS)/$(shell go env GOARCH)
INSTALL_PATH := ~/.terraform.d/plugins/registry.terraform.io/$(NAMESPACE)/$(NAME)/$(VERSION)/$(OS_ARCH)

.PHONY: build
build:
	go build -o $(BINARY) .

.PHONY: install
install: build
	mkdir -p $(INSTALL_PATH)
	mv $(BINARY) $(INSTALL_PATH)/$(BINARY)_v$(VERSION)

.PHONY: test
test:
	go test ./... -v -timeout 120s

.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v -timeout 120m

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: fmt
fmt:
	gofmt -s -w .
	goimports -w .

.PHONY: vet
vet:
	go vet ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: docs
docs:
	go generate ./...

.PHONY: clean
clean:
	rm -f $(BINARY)

.PHONY: all
all: tidy fmt vet build
