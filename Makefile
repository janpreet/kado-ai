VERSION ?= $(shell git describe --tags --always --dirty)

.PHONY: test
test:
	go test ./... -v

.PHONY: build
build:
	go build ./...

.PHONY: tag
tag:
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)

.PHONY: publish
publish: test build tag
	GOPROXY=proxy.golang.org go list -m github.com/yourusername/kado-ai@$(VERSION)

.PHONY: ci
ci: test build