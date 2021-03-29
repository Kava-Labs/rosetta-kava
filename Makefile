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

.PHONY: watch
watch:
	while sleep 0.5; do find . -type f -name '*.go' | entr -d go test -v ./...; done

.PHONY: gen-mocks
gen-mocks:
	mockery --dir services --all --case underscore --outpkg services --output mocks/services;

.PHONY: clean-mocks
clean-mocks:
	rm -r mocks
