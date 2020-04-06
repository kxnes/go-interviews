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
	make -C apicache
	go test -race -cover ./counter/...
	go test -race -cover ./parallels/...
