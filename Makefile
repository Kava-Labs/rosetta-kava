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

.PHONY: run
run:
	MODE=online NETWORK=kava-5.1 PORT=8000 go run . run

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

.PHONY: install-swagger-cli
install-swagger-cli:
	npm install -g @apidevtools/swagger-cli@4.0.2

.PHONY: vaidate-swagger
validate-swagger:
	swagger-cli validate swagger/api.yaml

.PHONY: run-swagger
run-swagger:
	docker run -p 8080:8080 -e SWAGGER_JSON=/spec/api.yaml -v $(PWD)/swagger:/spec swaggerapi/swagger-ui
