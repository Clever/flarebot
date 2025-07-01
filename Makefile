include golang.mk
.DEFAULT_GOAL := test # override default goal set in library makefile

SHELL := /bin/bash
PKG = github.com/Clever/flarebot
PKGS := $(shell go list ./... | grep -v /vendor)
EXECUTABLE := flarebot

TESTS=$(shell cd src/ && find . -name "*.test.js")

FORMATTED_FILES := $(shell find src/ -name "*.js")
MODIFIED_FORMATTED_FILES := $(shell git diff --name-only master $(FORMATTED_FILES))

PRETTIER := ./node_modules/.bin/prettier
ESLINT := ./node_modules/.bin/eslint

.PHONY: test $(PKGS) clean vendor format format-all format-check lint-es lint-fix lint

# $(eval $(call golang-version-check,1.24))

all: test build

test: $(PKGS)

build:
	go build -o bin/jira-cli github.com/Clever/flarebot/jira/testcmd
	go build -o bin/slack-cli github.com/Clever/flarebot/slack/testcmd
	go build -o bin/$(EXECUTABLE) $(PKG)


# for later, when I want to go strict
#$(PKGS): golang-test-all-strict-deps
#	$(call golang-test-all-strict,$@)

$(PKGS): golang-test-all-deps
	$(call golang-test-all,$@)

install_deps:
	go mod vendor


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
	./node_modules/jest/bin/jest.js src/$@

run:
	node src/app.js