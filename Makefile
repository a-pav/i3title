APP_NAME := i3title

GO := GO111MODULE=on CGO_ENABLED=0 go
LDFLAGS = -s -w

.PHONY: build

build:
	@echo "Building..."
	$(GO) build -trimpath -ldflags '$(LDFLAGS)' -o $(APP_NAME) ./cmd/$(APP_NAME)
