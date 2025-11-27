SHELL := bash
MAKEFLAGS += --no-print-directory

#######################
## Tools
#######################
export PATH := $(CURDIR)/bin:$(PATH)
OCB ?= $(CURDIR)/bin/builder

## @help:install-ocb:Install ocb.
.PHONY: install-ocb
install-ocb:
	GOBIN=$(CURDIR)/bin go install go.opentelemetry.io/collector/cmd/builder@v0.140.0

## MAKE GOALS
.PHONY: build
build: install-ocb ## Build the binary
	@$(OCB) --config builder-config.yml

.PHONY: run
run: ## Run the binary
	./bin/custom --config config.yml
