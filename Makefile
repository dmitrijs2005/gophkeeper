.PHONY: \
	vet \
	fmt

APP_NAME       ?= cli
CMD_DIR        ?= ./cmd/$(APP_NAME)
OUT_DIR        ?= ./bin

PKG := github.com/dmitrijs2005/gophkeeper

BUILD_VERSION  ?= $(shell git describe --tags --always)
BUILD_DATE     ?= $(shell date +%Y/%m/%d\ %H:%M:%S)
BUILD_COMMIT   ?= $(shell git rev-parse --short HEAD)

LDFLAGS = -X '${PKG}/internal/buildinfo.buildVersion=${BUILD_VERSION}' \
          -X '${PKG}/internal/buildinfo.buildDate=${BUILD_DATE}' \
          -X '${PKG}/internal/buildinfo.buildCommit=${BUILD_COMMIT}'


vet:
	go vet ./...

PKGS := $(shell go list ./... | grep -v '/proto')
COVER_THRESHOLD ?= 80
COVER_PROFILE ?= coverage.out

.PHONY: test
test:
	@echo "→ Packages (excluding /proto):"
	@printf '%s\n' $(PKGS)
	@echo
	@echo "→ Running tests..."
	@coverpkg=$$(printf '%s\n' $(PKGS) | paste -sd, -); \
	go test -coverpkg="$$coverpkg" -coverprofile=$(COVER_PROFILE) $(PKGS)
	@echo
	@echo "→ Coverage summary:"
	@line=$$(go tool cover -func=$(COVER_PROFILE) | awk '/^total:/ {print}'); \
	echo "$$line"; \
	pct=$$(echo "$$line" | awk '{gsub("%","",$$3); print $$3}'); \
	thresh=$(COVER_THRESHOLD); \
	awk -v p="$$pct" -v t="$$thresh" 'BEGIN{ if (p+0 < t+0) exit 1 }' \
	|| { echo "Coverage $$pct% < $$thresh% threshold"; exit 1; }; \
	echo "✓ Coverage OK (≥ $(COVER_THRESHOLD)%)"

fmt:
	go fmt ./...

.PHONY: build.win
build.win: ## Windows (amd64)
	@mkdir -p $(OUT_DIR)
	@GOOS=windows GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) \
		go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/$(APP_NAME).exe $(CMD_DIR)
	@echo "Built $(OUT_DIR)/$(APP_NAME).exe"

.PHONY: build.linux
build.linux: ## Сборка под Linux (amd64)
	@mkdir -p $(OUT_DIR)
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) \
		go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/$(APP_NAME)-linux $(CMD_DIR)

.PHONY: build.darwin
build.darwin: ## Сборка под macOS (arm64)
	@mkdir -p $(OUT_DIR)
	@GOOS=darwin GOARCH=arm64 CGO_ENABLED=$(CGO_ENABLED) \
		go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/$(APP_NAME)-darwin $(CMD_DIR)
