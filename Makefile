lint: golangci-lint-run
lint: build-staticlint-if-not-exists
lint: staticlint

golangci-lint-run:
	golangci-lint run -c .golangci.yml

build-staticlint-if-not-exists:
	[ -f ./cmd/staticlint/staticlint ] || go build -o ./cmd/staticlint/staticlint ./cmd/staticlint

build-staticlint:
	go build -o ./cmd/staticlint/staticlint ./cmd/staticlint

staticlint:
	go vet -vettool=./cmd/staticlint/staticlint ./...

coverprofile:
	go test ./... -covermode=count -coverprofile cover.out.tmp && cat cover.out.tmp | grep -v -e "mock" > cover.out \
 		&& rm cover.out.tmp && go tool cover -html cover.out -o coverprofile.html \
 		&& go tool cover -func cover.out > coverprofile.txt && cat coverprofile.txt

coverprofile-with-mocks:
	go test ./... -covermode=count -coverprofile cover.out && go tool cover -html cover.out -o coverprofile.html \
		&& go tool cover -func cover.out > coverprofile.txt && cat coverprofile.txt

version:
	go generate ./cmd/...

