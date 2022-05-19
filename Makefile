PWD = $(shell pwd -L)
GOCMD=go
DOCKERCMD=docker
DOCKERCOMPOSECMD=docker-compose
GOTEST=$(GOCMD) test
IMAGE_NAME = go-money
LIBRARY_ENV ?= dev

.PHONY: all test vendor

all: help

help: ## Display help screen
	@echo "Usage:"
	@echo "	make [COMMAND]"
	@echo "	make help \n"
	@echo "Commands: \n"
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

clean: fmt ## Remove unused files
	rm -f ./coverage.out
	rm -rf bin/

test: fmt test-clean ## Run the tests of the project
	$(GOTEST) -cover -p=1 ./...

test-clean: fmt ## Run the clean cache tests of the project
	$(GOCMD) clean -testcache

coverage: fmt ## Run the tests of the project and open the coverage in a Browser
	$(GOTEST) -cover -p=1 -covermode=count -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

tidy: ## Downloads go dependencies
	$(GOCMD) mod tidy

fmt: tidy ## Run go fmt
	$(GOCMD) fmt ./...