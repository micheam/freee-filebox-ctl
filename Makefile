BIN_NAME = ffbox
TARGET = bin/$(BIN_NAME)
SOURCE = $(shell find . -name '*.go')
INSTALL_PATH ?= /usr/local/bin
SCHEMA_FILE = internal/freeeapi/api-schema.json

# Version detection: use git tag if available, otherwise DEVEL_${SHORT_HASH}
GIT_TAG = $(shell git describe --tags --exact-match 2>/dev/null)
GIT_COMMIT = $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
VERSION = $(if $(GIT_TAG),$(GIT_TAG),DEVEL_$(GIT_COMMIT))

# ldflags for version injection
LDFLAGS = -ldflags "-X main.version=$(VERSION)"

.PHONY : clean install latest-freeeapi-schema version

$(TARGET) : $(SOURCE) ## Build the binary
	go build $(LDFLAGS) -o $(TARGET) ./cmd/ffbox

clean: ## Clean the build artifacts
	rm -f $(TARGET)

install : $(TARGET) ## Install the binary to INSTALL_PATH (default: /usr/local/bin)
	install -m 755 $(TARGET) $(INSTALL_PATH)/$(BIN_NAME)

test : ## Run tests and vet the code
	go test ./...
	go vet ./...

latest-freeeapi-schema : ## Download and update the freee API schema
	bash scripts/download-freeeapi-schema.sh -o $(SCHEMA_FILE)
	cd internal/freeeapi && go generate .

version : ## Show the version that will be embedded in the binary
	@echo $(VERSION)
