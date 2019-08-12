
# Go parameters
GO := go
GOBUILD := $(GO) build
GOCLEAN := $(GO) clean
GOTEST := $(GO) test
GOGET := $(GO) get

bundle:
	cd javascript && \
	packr2

build: bundle
	$(GOBUILD) -o plasma ./cmd

runtest: build
	cd ./cmd/testdata && \
	../../plasma -d .

runtest2: build
	cd ./cmd/testdata && \
	../../plasma -d . -dp ./features/import.js
