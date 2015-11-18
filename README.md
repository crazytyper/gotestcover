# Go test cover with multiple packages support

A fork of https://github.com/pierrre/gotestcover that integrates https://github.com/wadey/gocovmerge.

## Features
- Coverage profile with multiple packages (`go test` doesn't support that)
- Merges coverage profiles

## Install
`go get github.com/crazytyper/gotestcover`

## Usage
```sh
gotestcover -coverprofile=cover.out mypackage
go tool cover -html=cover.out -o=cover.html
```

Run on multiple package with:
- `package1 package2`
- `package/...`

Some `go test / build` flags are available.
