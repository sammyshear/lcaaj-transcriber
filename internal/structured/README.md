# Experimental Converter

This package is a draft alternative to `internal/transcribe.go`. Normal builds
keep the existing converter; the `structured` build tag selects this parser.

Run the web application with the experimental converter:

```sh
go run -tags structured .
```
