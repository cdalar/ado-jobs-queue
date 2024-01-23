GO_CMD=go
BINARY_NAME=jobs

# Mark targets as phony (not files)
.PHONY: all build clean run test

# Default target
all: build

# Build the binary
build:
	$(GO_CMD) build -o $(BINARY_NAME) .

# Clean up the binary
clean:
	rm $(BINARY_NAME)

# Test the application
test:
	$(GO_CMD) test ./...
