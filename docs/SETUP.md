# 🛠️ Setting Up Your Go Environment

## 1. Install Go

Install Go 1.25 or newer from <https://go.dev/dl/> and verify it:

```bash
go version
```

As of July 2026, Go 1.26 is current and Go 1.25 is the previous supported major
release. This course declares Go 1.25 as its minimum.

## 2. Get the code

```bash
git clone https://github.com/mbrndiar/learning-go.git
cd learning-go
```

## 3. Understand the toolchain

Go includes its compiler, formatter, package manager, test runner, coverage
tool, profiler, and documentation browser. There is no virtual environment to
activate.

Modern Go can automatically download the toolchain required by `go.mod` when
`GOTOOLCHAIN=auto` (the default):

```bash
go env GOTOOLCHAIN
go env GOVERSION
```

To select an installed toolchain explicitly:

```bash
GOTOOLCHAIN=go1.26.5 go version
```

## 4. Download module dependencies

Most lessons use only the standard library. Download the capstone dependency:

```bash
go mod download
```

## 5. Choose an editor

- [VS Code with the official Go extension](https://marketplace.visualstudio.com/items?itemName=golang.Go)
- [GoLand](https://www.jetbrains.com/go/)
- Neovim, Vim, Emacs, or another editor configured with
  [`gopls`](https://go.dev/gopls/)

The Go extension normally installs `gopls`,
[Delve (`dlv`)](https://github.com/go-delve/delve/tree/master/Documentation),
and related tools when they are first needed. The official
[`gopls` editor guide](https://go.dev/gopls/editor/) covers other editors.

## 6. Run the first lesson

```bash
go run ./lessons/01_basics/01_hello_world
```

## 7. Essential commands

```bash
go run ./path/to/package
go test ./path/to/package
go test -race ./...
go vet ./...
go fmt ./...
go mod tidy
go doc package.Symbol
```

At the course root, `go test ./...` also runs intentionally incomplete starter
exercises, so failures there are expected. Follow the workflow in
[`exercises/README.md`](../exercises/README.md) to compile starters or run
reference solutions separately.

## Troubleshooting

### The installed Go version is too old

Install a current release or leave `GOTOOLCHAIN=auto` enabled so the `go`
command can download the version requested by `go.mod`.

### A module cannot be downloaded

Check network/proxy settings:

```bash
go env GOPROXY GOPRIVATE
```

Public modules normally use `https://proxy.golang.org,direct`.

### The editor cannot find imports

Open the repository root—the directory containing `go.mod`—rather than one
lesson subdirectory. Restart `gopls` after changing toolchains.

### Cached module data appears broken

Inspect before cleaning:

```bash
go env GOMODCACHE
go clean -modcache
go mod download
```

`go clean -modcache` removes downloaded dependencies and should be a last
resort, not a routine command.
