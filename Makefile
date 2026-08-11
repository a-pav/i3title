APP_NAME := i3title

# Default installation path (can be overridden)
PREFIX ?= $(HOME)/.local
BINDIR ?= $(PREFIX)/bin

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

# Go build variables
GO ?= go
GOFLAGS ?= -trimpath
LDFLAGS ?= -s -w -X main.Version=$(VERSION)

# Build output directory
BUILD_DIR := dist

.PHONY: all build install uninstall clean

# The default target if you just type 'make'
all: build

build:
	@mkdir -p $(BUILD_DIR)
	@GO111MODULE=on CGO_ENABLED=0 $(GO) build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o $(BUILD_DIR)/$(APP_NAME) ./cmd/$(APP_NAME)
	@cp bin/i3toast $(BUILD_DIR)/i3toast
	@sed -i 's/VERSION="@VERSION@"$$/VERSION="$(VERSION)"/' $(BUILD_DIR)/i3toast

install: build
	@echo "Installing to $(BINDIR)..."
	install -d $(BINDIR)
	install -m 755 $(BUILD_DIR)/$(APP_NAME) $(BINDIR)/$(APP_NAME)
	install -m 755 $(BUILD_DIR)/i3toast $(BINDIR)/i3toast
	@echo "Done!"

uninstall:
	@echo "Removing from $(BINDIR)..."
	rm -f $(BINDIR)/$(APP_NAME)
	rm -f $(BINDIR)/i3toast

clean:
	@echo "Cleaning up..."
	rm -rf $(BUILD_DIR)
