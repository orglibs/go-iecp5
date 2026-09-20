# go-iecp5

A pure Go library for IEC 60870-5-104 communication over TCP/IP.
Requires Go 1.23 or later.

[![Go Reference](https://pkg.go.dev/badge/github.com/orglibs/go-iecp5.svg)](https://pkg.go.dev/github.com/orglibs/go-iecp5)
[![Go Report Card](https://goreportcard.com/badge/github.com/orglibs/go-iecp5)](https://goreportcard.com/report/github.com/orglibs/go-iecp5)
[![Tests](https://github.com/orglibs/go-iecp5/actions/workflows/go.yml/badge.svg)](https://github.com/orglibs/go-iecp5/actions/workflows/go.yml)

## Installation

```sh
go get github.com/orglibs/go-iecp5@latest
```

Import the packages you need:

```go
import (
    "github.com/orglibs/go-iecp5/asdu"
    "github.com/orglibs/go-iecp5/cs104"
)
```

The module root groups the packages; it is not an importable Go package.

## Packages

- [asdu](https://pkg.go.dev/github.com/orglibs/go-iecp5/asdu): application messages, information objects and time encoding.
- [cs104](https://pkg.go.dev/github.com/orglibs/go-iecp5/cs104): IEC 60870-5-104 client/server communication, including TLS.
- [cs101](https://pkg.go.dev/github.com/orglibs/go-iecp5/cs101): frame constants and types only; a complete IEC 60870-5-101 transport is not implemented.

Most application message types are supported. File transfer is not implemented.
See [UPSTREAM.md](UPSTREAM.md) for upstream provenance and compatibility details.

## Development and publishing

Run `make test` for tests with the race detector, coverage, vet and formatting checks.
See [RELEASING.md](RELEASING.md) for publishing and verifying pkg.go.dev indexing.
See [REVIEW.md](REVIEW.md) for the latest review and remaining limitations.

## References

- [lib60870 C library](https://github.com/mz-automation/lib60870)
- [lib60870 documentation](https://support.mz-automation.de/doc/lib60870/latest/group__CS104__MASTER.html)
- [License text](LICENSE)
