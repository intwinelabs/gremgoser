PKG := "github.com/intwinelabs/gremgoser"
PKG_NAME := "gremgoser"
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)
GO_VENDOR := $(shell which govendor)
USER := $(shell whoami)
IP := $(shell hostname -I | sed 's/ //')
.PHONY: all 

all: test ## Make all

compiletest: ## Compiles test
	@go test -v  ./... -run XXxxxXXXxxx  # ensures tests compile before running

test: ## Tests the code
	@go test -v ./... -count=1

cover: ## Generates test coverage report
	@echo "==> Running go test coverage tools: "
	@echo " "
	@go test -v -coverprofile /tmp/${PKG_NAME}.coverage.out ./... || { echo ""; echo "======> Go Tests Failed"; return 1; }
	@echo ""
	@echo "===> Generating go test coverage report: "
	@echo ""
	@go tool cover -html=/tmp/${PKG_NAME}.coverage.out -o /tmp/${PKG_NAME}.coverage.htm || { echo ""; echo "======> Go Coverage Reports Failed"; return 1; }
	@echo ""
	@echo "===> Opening coverage report: "
	@echo ""
	@echo "Browse to: http://${IP}:42280/${PKG_NAME}.coverage.htm"
	@cd /tmp && python -m SimpleHTTPServer 42280

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
