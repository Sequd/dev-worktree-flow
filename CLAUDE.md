# Build & Install Instructions

## Prerequisites

- Go 1.22+
- Make (GNU Make or compatible)
- `$GOBIN` or `$HOME/go/bin` must be in your `$PATH`

## Quick Start

```bash
make build      # Build binary to project root
make install    # Build + copy binary to $GOBIN (~/go/bin by default)
make run        # Build + run immediately
make tidy       # go mod tidy
make clean      # Remove built binaries
```

## How It Works

1. **Entrypoint** is auto-detected: Makefile finds the first `main.go` under `cmd/`.
2. **Build** produces a binary in the project root (`.exe` on Windows, no extension on Linux/macOS).
3. **Install** copies the binary to `$GOBIN` (defaults to `~/go/bin`). Ensure this directory is in your `$PATH`:
   ```bash
   export PATH="$HOME/go/bin:$PATH"
   ```
4. Binary name is defined by `BINARY_NAME` variable in the Makefile.

## Overrides

| Variable | Default | Example |
|----------|---------|---------|
| `GOBIN` | `~/go/bin` | `make install GOBIN=/usr/local/bin` |
| `BINARY_NAME` | `flow` | `make build BINARY_NAME=my-tool` |

## For Agents: Applying This Pattern to Another Project

To set up an identical build/install workflow for a new Go CLI project:

1. Ensure the project has a standard Go layout:
   ```
   project-root/
   ├── cmd/
   │   └── <app-name>/
   │       └── main.go      # entrypoint
   ├── internal/             # private packages
   ├── go.mod
   ├── go.sum
   └── Makefile
   ```

2. Copy the Makefile from this project.

3. Change only one variable:
   ```makefile
   BINARY_NAME = <new-app-name>
   ```

4. Everything else (OS detection, entrypoint discovery, install target) works automatically.

5. Verify:
   ```bash
   make build    # binary appears in project root
   make install  # binary copied to ~/go/bin
   which <new-app-name>  # should resolve
   ```
