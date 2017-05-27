include golang.mk
.DEFAULT_GOAL := test # override default goal set in library makefile

SHELL := /bin/bash
PKG ?= github.com/Clever/flarebot

PKGS := $(shell go list ./... | grep -v /vendor)
EXECUTABLE := flarebot
.PHONY: test $(PKGS) clean vendor

$(eval $(call golang-version-check,1.8))

$(GOPATH)/bin/glide:
	@go get github.com/Masterminds/glide

all: install_deps test build

test: $(PKGS)

build: install_deps
	go build -o bin/jira-cli $(PKG)/jira/testcmd
	go build -o bin/slack-cli $(PKG)/slack/testcmd
	GOOS=linux go build -o bin/$(EXECUTABLE) $(PKG)

install_deps: $(GOPATH)/bin/glide
	@$(GOPATH)/bin/glide install

# for later, when I want to go strict
#$(PKGS): golang-test-all-strict-deps
#	$(call golang-test-all-strict,$@)

$(PKGS): golang-test-all-deps
	$(call golang-test-all,$@)

vendor: golang-godep-vendor-deps
	$(call golang-godep-vendor,$(PKGS))
