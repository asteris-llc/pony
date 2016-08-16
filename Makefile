TEST?=$$(glide nv)
NAME = $(shell awk -F\" '/^const Name/ { print $$2 }' main.go)
VERSION = $(shell awk -F\" '/^const Version/ { print $$2 }' main.go)
DEPS = $(shell go list -f '{{range .TestImports}}{{.}} {{end}}' ./...)

GOOS=$(shell go env GOOS)
ARCH=$(shell go env GOARCH)

all: dev

dev: 
	@mkdir -p bin/
	gox -os=$(GOOS) -arch=$(ARCH) -output "bin/pony" \
		$$(glide nv)

build: plugins
	@mkdir -p bin/
	go build -o bin/$(NAME)

test: 
	go test $(TEST) $(TESTARGS) -timeout=30s -parallel=4
	go vet $(TEST)

plugins: 
	go build 

xcompile:
	@rm -rf build/
	@mkdir -p build
	gox \
		-os="darwin" \
		-output="build/pony_$(VERSION)_{{.OS}}_{{.Arch}}/pony-{{.Dir}}" $$(glide nv)

#		-os="freebsd" \
#		-os="linux" \
#		-os="windows" \

package: xcompile 
	$(eval FILES := $(shell ls build))
	@mkdir -p build/tgz
	for f in $(FILES); do \
		(cd $(shell pwd)/build && tar -zcvf tgz/$$f.tar.gz $$f); \
		echo $$f; \
	done

vendor:
	-rm -rf vendor/github.com/hashicorp/terraform
	glide install
	glide update
	rm -rf vendor/github.com/hashicorp/terraform/vendor

.PHONY: all build test xcompile package vendor
