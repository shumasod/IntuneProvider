DEFAULT_GOAL := build

PROVIDER_BINARY := terraform-provider-intune
CLI_BINARY      := intune
NAMESPACE       := shumasod
NAME            := intune
VERSION         := 0.1.0
OS_ARCH         := $(shell go env GOOS)_$(shell go env GOARCH)
LDFLAGS         := -ldflags="-X main.version=v$(VERSION)"

INSTALL_PATH := ~/.terraform.d/plugins/registry.terraform.io/$(NAMESPACE)/$(NAME)/$(VERSION)/$(OS_ARCH)
CLI_INSTALL  := /usr/local/bin

# ─── Provider ─────────────────────────────────────────────────────────────────

.PHONY: build
build:
	go build -o $(PROVIDER_BINARY) .

.PHONY: install
install: build
	mkdir -p $(INSTALL_PATH)
	mv $(PROVIDER_BINARY) $(INSTALL_PATH)/$(PROVIDER_BINARY)_v$(VERSION)

# ─── CLI ──────────────────────────────────────────────────────────────────────

.PHONY: build-cli
build-cli:
	go build $(LDFLAGS) -o $(CLI_BINARY) ./cmd/intune/

.PHONY: install-cli
install-cli: build-cli
	install -m 0755 $(CLI_BINARY) $(CLI_INSTALL)/$(CLI_BINARY)
	@echo "Installed $(CLI_INSTALL)/$(CLI_BINARY)"

# Install shell completions (bash / zsh / fish)
.PHONY: completions
completions: build-cli
	@mkdir -p completions
	./$(CLI_BINARY) completion bash  > completions/intune.bash
	./$(CLI_BINARY) completion zsh   > completions/intune.zsh
	./$(CLI_BINARY) completion fish  > completions/intune.fish
	@echo "Completions written to ./completions/"

# ─── Quality ──────────────────────────────────────────────────────────────────

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

.PHONY: vet
vet:
	go vet ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: docs
docs:
	go generate ./...

# ─── Housekeeping ─────────────────────────────────────────────────────────────

.PHONY: clean
clean:
	rm -f $(PROVIDER_BINARY) $(CLI_BINARY)
	rm -rf completions/

.PHONY: all
all: tidy fmt vet build build-cli
