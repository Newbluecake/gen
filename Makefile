.PHONY: all test 

export CC := clang
export CXX := clang++

LLVM_CONFIG ?= /usr/lib/llvm-14/bin/llvm-config
CGO_CFLAGS=
CGO_LDFLAGS=$(strip -L$(shell ${LLVM_CONFIG} --libdir) -Wl,-rpath,$(shell ${LLVM_CONFIG} --libdir))

all: test

test:
	CGO_CFLAGS='${CGO_CFLAGS}' CGO_LDFLAGS='${CGO_LDFLAGS}' go test -v -race -shuffle=on ./...

coverage:
	CGO_CFLAGS='${CGO_CFLAGS}' CGO_LDFLAGS='${CGO_LDFLAGS}' go test -v -covermode=atomic -coverpkg=./... -coverprofile=coverage.out ./...

install:
	CGO_CFLAGS='${CGO_CFLAGS}' CGO_LDFLAGS='${CGO_LDFLAGS}' go build -o ${GOPATH}/bin/go-clang-gen ./cmd/go-clang-gen