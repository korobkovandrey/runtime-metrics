.PHONY: golangci-lint-run
golangci-lint-run:
	golangci-lint run -c .golangci.yml

coverprofile:
	go test ./... -covermode=count -coverprofile cover.out.tmp && cat cover.out.tmp | grep -v -e "mock" -e "test" -e "logging" > cover.out \
 		&& rm cover.out.tmp && go tool cover -html cover.out -o coverprofile.html && go tool cover -func cover.out
