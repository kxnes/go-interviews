ifdef pkg
    package = ./${pkg}/...
else
    package = ./...
endif

all: lint test

lint:
	gofmt -s -w .
	golangci-lint run ${package}
	golint ${package}

test:
	go test -race -cover ${package}
