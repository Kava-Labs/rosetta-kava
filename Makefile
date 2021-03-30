.PHONY: install
install:
	go install

.PHONY: lint
lint:
	golint ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: build
build:
	go build ./...

.PHONY: test
test:
	go test -v ./...

.PHONY: cover
cover:
	go test -coverprofile=c.out ./...
	go tool cover -html=c.out -o coverage.html

.PHONY: watch
watch:
	while sleep 0.5; do find . -type f -name '*.go' | entr -d go test ./...; done

.PHONY: gen-mocks
gen-mocks:
	mockery --dir services --all --case underscore --outpkg services --output mocks/services;

.PHONY: clean-mocks
clean-mocks:
	rm -r mocks
