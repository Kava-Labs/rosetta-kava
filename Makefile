.PHONY: install
install:
	go install

.PHONY: lint
lint:
	go run golang.org/x/lint/golint ./...

.PHONY: golangci-lint
golangci-lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint run

.PHONY: vet
vet:
	go vet ./...

.PHONY: build
build:
	go build ./...

.PHONY: run
run:
	MODE=online NETWORK=kava-7 PORT=8000 KAVA_RPC_URL=https://rpc.kava.io:443 go run . run

.PHONY: test
test:
	go test -v ./...

.PHONY: test-integration
test-integration:
	MODE=online go test -v -tags=integration -count=1 ./testing
	MODE=offline go test -v -tags=integration ./testing

.PHONY: cover
cover:
	go test -coverprofile=c.out ./...
	go tool cover -html=c.out -o coverage.html

.PHONY: watch
watch:
	while sleep 0.5; do find . -type f -name '*.go' | entr -d go test ./...; done

.PHONY: watch-integration
watch-integration:
	while sleep 0.5; do find . -type f -name '*.go' | entr -d go test -tags=integration ./testing; done

.PHONY: gen-mocks
gen-mocks:
	mockery --dir services --all --case underscore --outpkg services --output mocks/services;
	mockery --srcpkg github.com/tendermint/tendermint/rpc/client --name Client --case underscore --outpkg tendermint --output mocks/tendermint

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
