
# Go parameters
GO := go
GOBUILD := $(GO) build
GOCLEAN := $(GO) clean
GOTEST := $(GO) test
GOGET := $(GO) get

bundle:
	cd cmd && \
	packr2

build: bundle
	$(GOBUILD) -o plasma ./cmd

runtest: build
	cd ./cmd/testdata && \
	../../plasma -d .
