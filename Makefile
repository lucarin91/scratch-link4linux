
init:
	go install golang.org/x/tools/cmd/goimports@v0.17.0
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ./bin v1.55.2

fmt:
	goimports -w .

test:
	go test -v ./...

lint:
	./bin/golangci-lint run ./...
