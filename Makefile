bin_dir=build
function=spark_server_staging
psql_connection=postgresql://postgres:postgres@127.0.0.1:45432/spark_server_dev

# env.mk can overwrite these env.
-include env.mk

.PHONY: prepare abigen test build

prepare:
	@git submodule update --init --recursive
	@docker-compose up -d postgres
	@cd contract && yarn && yarn hardhat compile
	@go mod download
	@if [ ! -f config/config.test.json ]; then cp config/config.{sample,test}.json; fi;
	@if [ ! -f config/config.json ]; then cp config/config.{sample,}.json; fi;
	@echo 'Done. Modify config/config.{test,}.json for your environment.'

abigen:
	@-mkdir abi
	jq .abi contract/artifacts/contracts/SparkLink.sol/SparkLink.json > abi/SparkLink.json
	abigen --abi abi/SparkLink.json --pkg abi --type SparkLink --out abi/spark_nft.go
	abigen --abi abi/ERC20.json --pkg abi --type ERC20 --out abi/erc20.go

test-prepare:
	@psql ${psql_connection} -c 'CREATE DATABASE spark_server_test;'

test:
	@go test -v ./...

psql:
	@psql ${psql_connection}

build:
	@go build -o ${bin_dir}/ ./cmd/...
