.PHONY: golangci-lint-run
golangci-lint-run:
	golangci-lint run -c .golangci.yml
