.PHONY: install
install:
	go install

.PHONY: lint
lint:
	golint -set_exit_status ./...

.PHONY: golangci-lint
golangci-lint:
	docker run --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:v1.57.2 golangci-lint run -v --timeout 10m

.PHONY: vet
vet:
	go vet ./...

.PHONY: build
build:
	go build ./...

.PHONY: run
run:
	MODE=online NETWORK=kava-mainnet PORT=8000 KAVA_RPC_URL=https://rpc.data.kava.io:443 go run . run

.PHONY: run
run-offline:
	MODE=offline NETWORK=kava-mainnet PORT=8001 KAVA_RPC_URL=https://rpc.data.kava.io:443 go run . run

.PHONY: run-testnet
run-testnet:
	MODE=online NETWORK=kava-testnet PORT=8000 KAVA_RPC_URL=https://rpc.data-testnet.kava.io:443 go run . run

.PHONY: run-local
run-local:
	MODE=online NETWORK=kava-localnet PORT=8000 KAVA_RPC_URL=http://localhost:26657 go run . run

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
	while sleep 0.5; do find . -type f -name '*.go' | entr -d go test -tags=integration -count=1 ./testing; done

.PHONY: rosetta-check-data
rosetta-check-data:
	rosetta-cli --configuration-file rosetta-cli-conf/kava-mainnet/config.json check:data

.PHONY: rosetta-check-construction
rosetta-check-construction:
	rosetta-cli --configuration-file rosetta-cli-conf/kava-mainnet/config.json check:construction

.PHONY: rosetta-check-data-local
rosetta-check-data-local:
	rosetta-cli --configuration-file rosetta-cli-conf/kava-localnet/config.json check:data

.PHONY: gen-mocks
gen-mocks:
	mockery --dir services --all --case underscore --outpkg services --output mocks/services;
	mockery --dir kava --all --case underscore --outpkg mocks --output kava/mocks

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
	docker run -p 8081:8080 -e SWAGGER_JSON=/spec/api.yaml -v $(PWD)/swagger:/spec swaggerapi/swagger-ui

.PHONY: clean
clean:
	rm -rf rosetta-cli-conf/cache
