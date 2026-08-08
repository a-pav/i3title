APP_NAME := i3title

# Default installation path (can be overridden)
PREFIX ?= $(HOME)/.local
BINDIR ?= $(PREFIX)/bin

# Go build variables
GO ?= go
GOFLAGS ?= -trimpath
LDFLAGS ?= -s -w

.PHONY: all build install uninstall clean

# The default target if you just type 'make'
all: build

build:
	GO111MODULE=on CGO_ENABLED=0 $(GO) build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o $(APP_NAME) ./cmd/$(APP_NAME)

install: build
	@echo "Installing to $(BINDIR)..."
	install -d $(BINDIR)
	install -m 755 $(APP_NAME) $(BINDIR)/$(APP_NAME)
	install -m 755 bin/i3toast $(BINDIR)/i3toast
	@echo "Done!"

uninstall:
	@echo "Removing from $(BINDIR)..."
	rm -f $(BINDIR)/$(APP_NAME)
	rm -f $(BINDIR)/i3toast

clean:
	@echo "Cleaning up..."
	rm -f $(APP_NAME)
