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

test:
	go test -v ./...

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
