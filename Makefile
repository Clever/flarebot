include golang.mk
.DEFAULT_GOAL := test # override default goal set in library makefile

SHELL := /bin/bash
PKG = github.com/Clever/flarebot
PKGS := $(shell go list ./... | grep -v /vendor)
EXECUTABLE := flarebot
.PHONY: test $(PKGS) clean vendor

$(eval $(call golang-version-check,1.7))

all: test build

test: $(PKGS)

build:
	go build -o bin/jira-cli github.com/Clever/flarebot/jira/testcmd
	go build -o bin/$(EXECUTABLE) $(PKG)

# for later, when I want to go strict
#$(PKGS): golang-test-all-strict-deps
#	$(call golang-test-all-strict,$@)

$(PKGS): golang-test-all-deps
	$(call golang-test-all,$@)

vendor: golang-godep-vendor-deps
	$(call golang-godep-vendor,$(PKGS))
