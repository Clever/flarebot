include golang.mk
include lambda.mk
.DEFAULT_GOAL := test # override default goal set in library makefile

SHELL := /bin/bash
PKG = github.com/Clever/flarebot
PKGS := $(shell go list ./... | grep -v /vendor)
EXECUTABLE := flarebot
LAMBDAS := $(shell [ -d "./cmd" ] && ls ./cmd/)
_APP_NAME ?= $(APP_NAME)

TESTS=$(shell cd src/ && find . -name "*.test.ts")

FORMATTED_FILES := $(shell find src/ -name "*.ts")
MODIFIED_FORMATTED_FILES := $(shell git diff --name-only master $(FORMATTED_FILES))

PRETTIER := ./node_modules/.bin/prettier
ESLINT := ./node_modules/.bin/eslint

.PHONY: test $(PKGS) clean vendor format format-all format-check lint-es lint-fix lint

# $(eval $(call golang-version-check,1.24))

all: test build

test: generate $(PKGS)

generate: tool
	go generate ./...
	go mod tidy

$(LAMBDAS): generate
	$(call lambda-build-go,./cmd/$@,$@)

build: $(LAMBDAS)
	go build -o bin/jira-cli github.com/Clever/flarebot/jira/testcmd
	go build -o bin/slack-cli github.com/Clever/flarebot/slack/testcmd
	go build -o bin/$(EXECUTABLE) $(PKG)

# for later, when I want to go strict
#$(PKGS): golang-test-all-strict-deps
#	$(call golang-test-all-strict,$@)

$(PKGS): golang-test-all-deps
	$(call golang-test-all,$@)

install_deps: vendor tool

format:
	@echo "Formatting modified files..."
	@$(PRETTIER) --write $(MODIFIED_FORMATTED_FILES)

format-all:
	@echo "Formatting all files..."
	@$(PRETTIER) --write $(FORMATTED_FILES)

format-check:
	@echo "Running format check..."
	@$(PRETTIER) --list-different $(FORMATTED_FILES) || \
		(echo -e "❌ \033[0;31mPrettier found discrepancies in the above files. Run 'make format(-all)' to fix.\033[0m" && false)

lint-es:
	@echo "Running eslint..."
	@$(ESLINT) $(FORMATTED_FILES)

lint-fix:
	@echo "Running eslint --fix..."
	@$(ESLINT) --fix $(FORMATTED_FILES) || \
		(echo "\033[0;31mThe above errors require manual fixing.\033[0m" && true)

lint: format-check lint-es

test-js: lint $(TESTS)

$(TESTS):
	NODE_ENV=test ./node_modules/jest/bin/jest.js src/$@

build-ts:
	./node_modules/.bin/tsc -p .

run-ts: build-ts
	node dist/app.js

run-local:
ifeq ($(_APP_NAME),flarebot)
	make run-ts
else
	$(call golang-build,./cmd/$(_APP_NAME),$(_APP_NAME))
	IS_LOCAL=true bin/$(_APP_NAME)
endif


# We generally want to use production catapult for testing since we don't run it in dev.
# To test failure cases you can also set env var for secrets to "x" here.

ifeq ($(_IS_LOCAL), true)
  export SERVICE_CATAPULT_HTTP_HOST=production--catapult.int.clever.com
  export SERVICE_CATAPULT_HTTP_PORT=443
  export SERVICE_CATAPULT_HTTP_PROTO=https
# export JIRA_PASSWORD=x
endif

run: run-local

vendor: go.mod go.sum
	go mod tidy
	go mod vendor

tool: vendor
	go install tool
