BINARY ?= $(shell basename "$(PWD)")# binary name
CMD := $(wildcard cmd/*.go)
temp = $(subst /, ,$@)
os = $(word 1, $(temp))
arch = $(word 2, $(temp))

.PHONY: run
run:
	go run cmd/*

# Clean the build directory (before committing code, for example)
.PHONY: clean
clean: 
	rm -rv bin

PLATFORMS := linux/amd64 windows/amd64 darwin/amd64 darwin/arm64

release: $(PLATFORMS)

$(PLATFORMS):
	GOOS=$(os) GOARCH=$(arch) go build -o 'bin/$(BINARY)-$(os)-$(arch)' $(CMD)

.PHONY: release $(PLATFORMS)

