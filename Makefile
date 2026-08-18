PARENT_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

include $(PARENT_DIR).parent/parent.mk

# ---- overrides ------------------------------------------------------------
SERVICE_NAME := go-web-skeleton
SOURCE.DIR   := ./cmd/server

# parent's service.mk derives SOURCE.PKGS from SOURCE.DIR, which would restrict
# `make test` to ./cmd/server/... and silently run zero tests for common/, pkg/,
# security/ and the generated modules. Test the whole module instead.
SOURCE.PKGS  := ./...

GOLANGCI_LINT.VERSION := v2.11.4
LINT.CONFIG := ./golangci.yml

TAILWIND.VERSION := v4.1.14
TAILWIND.BIN     := $(TOOLS.DIR)/tailwindcss
TAILWIND.INPUT   := common/http/static/files/css/tailwind.css
TAILWIND.OUTPUT  := common/http/static/files/css/styles.css

.PHONY: css setup-css-tools

setup-css-tools:
	@$(call log.info, Setup tailwind CLI started)
	@mkdir -p $(TOOLS.DIR)
	@test -x $(TAILWIND.BIN) || curl -sSfL -o $(TAILWIND.BIN) \
		"https://github.com/tailwindlabs/tailwindcss/releases/download/$(TAILWIND.VERSION)/tailwindcss-$(GO.OS)-x64" \
		|| ( $(call log.error, Download tailwind CLI failed) && false )
	@chmod +x $(TAILWIND.BIN)
	@$(call log.info, Setup tailwind CLI finished successfully)

# styles.css is a build artifact: never hand-edit it, regenerate it.
css: setup-css-tools
	@$(call log.info, Build stylesheet started)
	@$(TAILWIND.BIN) -i $(TAILWIND.INPUT) -o $(TAILWIND.OUTPUT) --minify \
		|| ( $(call log.error, Build stylesheet failed) && false )
	@$(call log.info, Stylesheet built successfully)
