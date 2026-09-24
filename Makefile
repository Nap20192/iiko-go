.PHONY: test cover race lint gen gen-server gen-cloud integration tidy invariants

test:
	go test ./...

race:
	go test ./... -race

# Cross-package: the domain packages are largely exercised through each other.
cover:
	go test ./... -coverpkg=./... -coverprofile=coverage.out
	@echo "hand-written only (generated types inflate the denominator):"
	@grep -vE '/(types|operations)\.go:' coverage.out > coverage.hand.out
	@go tool cover -func=coverage.hand.out | tail -1

lint:
	go vet ./...
	@command -v golangci-lint >/dev/null && golangci-lint run || echo "golangci-lint not installed; ran go vet only"

gen: gen-server gen-cloud
gen-server:
	go run ./tools/gen/server
gen-cloud:
	go run ./tools/gen/cloud

# Against a real iiko stand. Needs IIKO_BASE_URL / IIKO_LOGIN / IIKO_PASSWORD.
integration:
	go test -tags integration ./iikoserver/ -run TestLive -v

tidy:
	go mod tidy
	gofmt -s -w .

invariants:
	./scripts/invariants.sh
