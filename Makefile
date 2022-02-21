BUILD=go build
CLEAN=go clean
INSTALL=go install
BUILDPATH=./_build
PACKAGES=$(shell go list ./... | grep -v /examples/)
PLATFORM=local

cmd: dumper grapher apisrv

all: dep test dumper grapher apisrv

dumper: dir
	go build -o "$(BUILDPATH)/" "./cmd/dumper/..."

grapher: dir
	go build -o "$(BUILDPATH)/" "./cmd/grapher/..."

apisrv: dir
	go build -o "$(BUILDPATH)/" "./cmd/apisrv/..."

dir:
	mkdir -p $(BUILDPATH)

clean:
	rm -rf $(BUILDPATH)

dep:
	go get ./...

test:
	for pkg in ${PACKAGES}; do \
		go test -coverprofile="../../../$$pkg/coverage.txt" -covermode=atomic $$pkg || exit; \
	done

build:
	go build ./...

docker-build-all: dir
	docker build . --target bin-all --output _build/ --platform ${PLATFORM}

.PHONY: clean dumper grapher
