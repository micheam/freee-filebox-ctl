BIN_NAME = ffbox
TARGET = bin/$(BIN_NAME)
SOURCE = $(shell find . -name '*.go')
INSTALL_PATH ?= /usr/local/bin
SCHEMA_FILE = freeeapi/api-schema.json
.PHONY : clean install latest-freeeapi-schema

$(TARGET) : $(SOURCE) ## Build the binary
	go build -o $(TARGET) .

clean: ## Clean the build artifacts
	rm -f $(TARGET)

install : $(TARGET) ## Install the binary to INSTALL_PATH (default: /usr/local/bin)
	install -m 755 $(TARGET) $(INSTALL_PATH)/$(BIN_NAME)

test : ## Run tests and vet the code
	go test ./...
	go vet ./...

latest-freeeapi-schema : ## Download and update the freee API schema
	bash scripts/download-freeeapi-schema.sh -o $(SCHEMA_FILE)
	cd freeeapi && go generate .
