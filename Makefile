bench:
	go test -benchmem -run=NONE -bench .

test: lint tools
	@echo "==> Running tests..."
	@gotestsum --format-hide-empty-pkg  ./...

lint: tools
	golangci-lint run

.PHONY: tools
tools:
	@echo "==> Installing tools from tools.go..."
	@awk -F'"' '/_/ {print $$2}' tools.go | xargs -I % go install %