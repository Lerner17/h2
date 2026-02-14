SHELL := /bin/bash

GO ?= go
BIN_DIR ?= bin
CLI_BIN ?= vpn-cli
PREFIX ?= /usr/local
INSTALL ?= install

.PHONY: help deps deps-go deps-tools deps-system build build-cli test run-cli wire install-cli uninstall-cli clean

help:
	@echo "Available targets:"
	@echo "  make deps           Install all dependencies (go + system tools)"
	@echo "  make deps-go        Download go modules"
	@echo "  make deps-tools     Install go dev tools (wire)"
	@echo "  make deps-system    Install ansible + qrencode via package manager"
	@echo "  make build          Build cli binary"
	@echo "  make test           Run all tests"
	@echo "  make install-cli    Install cli globally (default: /usr/local/bin/$(CLI_BIN))"
	@echo "  make uninstall-cli  Remove global cli binary"

deps: deps-go deps-tools deps-system

deps-go:
	$(GO) mod tidy
	$(GO) mod download

deps-tools:
	$(GO) install github.com/google/wire/cmd/wire@v0.6.0

deps-system:
	@if command -v brew >/dev/null 2>&1; then \
		echo "Installing ansible and qrencode via brew"; \
		brew install ansible qrencode; \
	elif command -v apt-get >/dev/null 2>&1; then \
		echo "Installing ansible and qrencode via apt-get"; \
		sudo apt-get update; \
		sudo apt-get install -y ansible qrencode; \
	elif command -v dnf >/dev/null 2>&1; then \
		echo "Installing ansible and qrencode via dnf"; \
		sudo dnf install -y ansible qrencode; \
	elif command -v yum >/dev/null 2>&1; then \
		echo "Installing ansible and qrencode via yum"; \
		sudo yum install -y epel-release; \
		sudo yum install -y ansible qrencode; \
	elif command -v apk >/dev/null 2>&1; then \
		echo "Installing ansible and qrencode via apk"; \
		sudo apk add --no-cache ansible qrencode; \
	elif command -v pacman >/dev/null 2>&1; then \
		echo "Installing ansible and qrencode via pacman"; \
		sudo pacman -Sy --noconfirm ansible qrencode; \
	else \
		echo "No supported package manager found. Install ansible and qrencode manually."; \
		exit 1; \
	fi

build: build-cli

build-cli:
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/$(CLI_BIN) ./cmd/cli

test:
	$(GO) test ./...

run-cli:
	$(GO) run ./cmd/cli

wire:
	$(GO) generate ./internal/hysteria/app/...

install-cli: build-cli
	$(INSTALL) -d $(PREFIX)/bin
	$(INSTALL) -m 0755 $(BIN_DIR)/$(CLI_BIN) $(PREFIX)/bin/$(CLI_BIN)
	@echo "Installed: $(PREFIX)/bin/$(CLI_BIN)"

uninstall-cli:
	rm -f $(PREFIX)/bin/$(CLI_BIN)
	@echo "Removed: $(PREFIX)/bin/$(CLI_BIN)"

clean:
	rm -rf $(BIN_DIR)
