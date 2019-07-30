
# Go parameters
GO := go
GOBUILD := $(GO) build
GOCLEAN := $(GO) clean
GOTEST := $(GO) test
GOGET := $(GO) get

build:
	cd cmd && \
	packr2
	$(GOBUILD) -o plasma ./cmd

runtest: build
	cd ./cmd/testdata && \
	../../plasma -d .
