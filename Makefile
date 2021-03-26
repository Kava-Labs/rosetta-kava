.PHONY: install
install:
	go install

.PHONY: lint
lint:
	golint ./...

.PHONY: vet
vet:
	go vet ./...
