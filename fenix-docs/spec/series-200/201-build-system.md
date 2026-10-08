# SPEC-201: Build System

## 1. Executive Summary
The Fénix build system strictly avoids complex, monolithic tooling chains like Make or Mage at its inception. It relies on the native capabilities of modern Go (specifically `go.work`) paired with `Taskfile` for declarative, readable automation. 
This ensures a low barrier to entry for contributors and rapid, reproducible builds across Windows, Linux, and macOS.

## 2. The Minimalist Toolchain
The foundational stack for building Fénix consists of:
- **Go 1.21+:** The core compiler.
- **`go.work`:** Workspace mode to orchestrate multiple isolated `go.mod` files within the monorepo (separating plugins from the core).
- **Task (Taskfile.dev):** A modern, YAML-based task runner replacing `Make`. It provides better cross-platform compatibility (especially critical for Windows users) and declarative syntax.
- **`golangci-lint`:** The definitive linter for Go, enforcing strict quality gates.
- **`gofumpt`:** A stricter version of `gofmt` to guarantee absolute formatting consistency.
- **`govulncheck`:** The official Go vulnerability scanner.

## 3. The Go Workspace (`go.work`)
To ensure that dependencies for the core engine do not get tangled with the heavy dependencies of specific AI drivers, Fénix uses a Go Workspace that links the `core` module with independent plugin modules.

The root `go.work` file links high-level modules, rather than internal folders:
```go
go 1.21

use (
    ./core                  // The central engine (Kernel, Runtime, Parser)
    ./cmd/fenix             // The CLI application
    ./sdk                   // Public developer SDKs
    ./plugins/claude        // Independent plugin module
    ./plugins/openai        // Independent plugin module
    ./plugins/gemini        // Independent plugin module
)
```
This isolates the `go.mod` of `claude` from the `go.mod` of `core`, maintaining a pristine dependency tree for the engine, while minimizing the number of `go.mod` files developers must manage.

## 4. Code Quality & CI Gates
Before any PR can be merged into Fénix, it must pass a rigorous pipeline enforced by the `Taskfile`.

Key `Taskfile` commands:
- `task lint`: Runs `golangci-lint run ./...` across all modules.
- `task format`: Runs `gofumpt -l -w .`.
- `task test`: Executes `go test -race -v ./...`.
- `task audit`: Runs `govulncheck ./...`.
- `task proto`: Compiles `/proto` definitions into `/core/pkg/api` using `buf` or `protoc`.

These identical tasks will be mirrored in the CI environment (e.g., GitHub Actions) ensuring that local development exactly matches the CI pipeline.

## 5. Cross-Compilation & Releases
Fénix is designed to run everywhere. The build system must effortlessly output binaries for multiple platforms.

A standard `task build:all` will invoke Go's native cross-compilation:
- `GOOS=windows GOARCH=amd64 go build -o bin/fenix-windows-amd64.exe ./cmd/fenix`
- `GOOS=linux GOARCH=amd64 go build -o bin/fenix-linux-amd64 ./cmd/fenix`
- `GOOS=darwin GOARCH=arm64 go build -o bin/fenix-darwin-arm64 ./cmd/fenix`

*Note: In the future, this phase will integrate cryptographically signed releases (via `cosign` or similar) to guarantee binary authenticity for enterprise deployments, aligning with the secure plugin manifest rules defined in SPEC-109.*
