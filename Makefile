GO ?= go
PACKAGES ?= ./...

.PHONY: check test package
check:
	$(GO) mod tidy
	$(GO) fmt ./...
	$(GO) vet $(PACKAGES)

test:
	LC_ALL=C $(GO) test -race -count=1 $(PACKAGES)
	python -m unittest discover -s tests

package:
	python scripts/package.py
