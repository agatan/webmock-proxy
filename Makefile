NAME := webmock-proxy
SRCS := $(shell find . -name "*.go" -type f)

LDFLAGS = -ldflags="-w -s -extldflags -static"

.DEFAULT_GOAL := bin/$(NAME)

bin/$(NAME): $(SRCS)
	go build $(LDFLAGS) -o bin/$(NAME)

.PHONY: glide
glide:
ifeq ($(shell command -v glide 2>/dev/null),)
	curl https://glide.sh/get | sh
endif

.PHONY: deps
deps: glide
	glide install

.PHONY: clean
clean:
	$(RM) -r bin/*
	$(RM) -r vendor/*

.PHONY: install
install:
	go install $(LDFLAGS)

.PHONY: test
test:
	go test -cover `glide novendor`
