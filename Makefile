#
# Copyright (c) 2024 Circutor S.A.
#

lint:
	golangci-lint run

test:
	go test -race -coverprofile=profile.cov ./...
	go tool cover -func profile.cov
	rm profile.cov
	go vet ./...
	@test -z "$$(gofmt -l .)" || (gofmt -l .; exit 1)
