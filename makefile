.PHONY: help build run test clean
GO_MODULE_PATH = asciify

# ROOT_DIR is the path of the makefile (including trailing slash)
ROOT_DIR := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))
PROJECT_PATH := $(ROOT_DIR:/=)
BIN_NAME = todo

help: ## Display this help message
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*##"; printf "\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ General

build: ## Build the binary
	go build -C '${ROOT_DIR}cmd' -o '${ROOT_DIR}${BIN_NAME}'

run: build ## Build and run the binary bin/imageservice
	${ROOT_DIR}/${BIN_NAME}

test: ## run all tests
	go test ${GO_MODULE_PATH}/...

clean: ## remove build files
	rm -rv '${ROOT_DIR}${BIN_NAME}'