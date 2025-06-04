.PHONY: golangci-lint-run
golangci-lint-run:
	golangci-lint run -c .golangci.yml

staticlint:
	go run ./cmd/staticlint ./...

coverprofile:
	go test ./... -covermode=count -coverprofile cover.out.tmp && cat cover.out.tmp | grep -v -e "mock" > cover.out \
 		&& rm cover.out.tmp && go tool cover -html cover.out -o coverprofile.html \
 		&& go tool cover -func cover.out > coverprofile.txt && cat coverprofile.txt

coverprofile-with-mocks:
	go test ./... -covermode=count -coverprofile cover.out && go tool cover -html cover.out -o coverprofile.html \
		&& go tool cover -func cover.out > coverprofile.txt && cat coverprofile.txt
