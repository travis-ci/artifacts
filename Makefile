ARTIFACTS_PACKAGE := github.com/meatballhat/artifacts
TARGETS := $(ARTIFACTS_PACKAGE)

VERSION_VAR := main.VersionString
REPO_VERSION := $(shell git describe --always --dirty --tags)

REV_VAR := main.RevisionString
REPO_REV := $(shell git rev-parse --sq HEAD)

GO ?= go
GODEP ?= godep
GOBUILD_LDFLAGS := -ldflags "-X $(VERSION_VAR) $(REPO_VERSION) -X $(REV_VAR) $(REPO_REV)"
GOBUILD_FLAGS ?=

all: clean test save

test: build fmtpolice test-deps coverage.html

test-deps:
	$(GO) test -i $(GOBUILD_LDFLAGS) $(GO_TAG_ARGS) $(GOBUILD_ARGS) $(TARGETS)

test-race:
	$(GO) test -race $(GOBUILD_LDFLAGS) $(GO_TAG_ARGS) $(GOBUILD_ARGS) $(TARGETS)

coverage.html: coverage.out
	$(GO) tool cover -html=$^ -o $@

coverage.out: path-coverage.out upload-coverage.out
	echo 'mode: count' > $@
	grep -h -v 'mode: count' $^ >> $@
	$(GO) tool cover -func=$@

path-coverage.out:
	$(GO) test -covermode=count -coverprofile=$@ $(GOBUILD_ARGS) $(GOBUILD_LDFLAGS) $(GO_TAG_ARGS) $(ARTIFACTS_PACKAGE)/path

upload-coverage.out:
	$(GO) test -covermode=count -coverprofile=$@ $(GOBUILD_ARGS) $(GOBUILD_LDFLAGS) $(GO_TAG_ARGS) $(ARTIFACTS_PACKAGE)/upload

build: deps
	$(GO) install $(GOBUILD_FLAGS) $(GOBUILD_LDFLAGS) $(GO_TAG_ARGS) $(TARGETS)

deps:
	$(GO) get $(GOBUILD_FLAGS) $(GOBUILD_LDFLAGS) $(GO_TAG_ARGS) $(TARGETS)
	$(GODEP) restore

clean:
	rm -vf $${GOPATH%%:*}/bin/artifacts
	rm -vf coverage.html *coverage.out
	$(GO) clean $(TARGETS) || true
	if [ -d $${GOPATH%%:*}/pkg ] ; then \
		find $${GOPATH%%:*}/pkg -name '*artifacts*' | xargs rm -rfv || true; \
	fi

save:
	$(GODEP) save -copy=false

fmtpolice:
	set -e; for f in $(shell git ls-files '*.go'); do gofmt $$f | diff -u $$f - ; done
