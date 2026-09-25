# Prerequisites with mise

These commands target Linux with Bash. Run the project commands from the repository root. The repository already includes `mise.toml`, frontend manifests, and a Go module. Install Git and Make before using the project commands.

## 1. Install mise

You need `curl` and CA certificates. On Ubuntu or Debian, install them with:

```bash
sudo apt update
sudo apt install -y curl ca-certificates
```

Skip installation if mise is already installed. Otherwise, use the [official installer](https://mise.jdx.dev/installing-mise.html):

```bash
curl -fsSL https://mise.run | sh
~/.local/bin/mise --version
```

## 2. Activate mise in Bash

For the default installer location, add this line to `~/.bashrc` once:

```bash
eval "$(~/.local/bin/mise activate bash)"
```

Activate it in the current terminal as well:

```bash
eval "$(~/.local/bin/mise activate bash)"
mise --version
```

If mise was installed through a system package manager, use `eval "$(mise activate bash)"` instead. Activation selects project tools when you enter the repository. See [mise getting started](https://mise.jdx.dev/getting-started.html).

## 3. Install the pinned project tools

From your checkout, review [mise.toml](../mise.toml), then run:

```bash
mise trust
mise install
```

Use the existing pins for Node, pnpm, Go, and Lefthook. Running `mise use` to select new versions would change the project configuration. See [mise getting started](https://mise.jdx.dev/getting-started.html).

## 4. Verify the tools

```bash
mise ls
mise exec -- node --version
mise exec -- pnpm --version
mise exec -- go version
mise doctor
```

Confirm Node reports `v24.x`, Go reports `go1.26.x`, and pnpm reports `12.x`, with exact versions matching `mise.toml`. `mise exec` uses the project tools even without shell activation.

## 5. Install application dependencies

```bash
pnpm --dir frontend install --frozen-lockfile
(cd backend && go mod download)
lefthook install
```

The [root README](../README.md#setup-and-development) explains how to start both layers and lists all validation commands. No database or credentials are required.

## Additional check prerequisites

- Staticcheck is pinned in `backend/go.mod` and invoked with `go tool staticcheck`; `gofmt` and `go vet` come with Go.
- Backend audit and pre-push require standalone `gosec` and `govulncheck` executables on `PATH`. This repository does not pin or install these tools.
- Race-enabled tests require a C compiler. On Ubuntu or Debian, the build tools can be installed with `sudo apt install -y build-essential`.
- Install Chromium and its system dependencies before browser tests with `pnpm --dir frontend exec playwright install --with-deps chromium`. System dependency installation may require administrator privileges.
- Frontend coverage tooling remains unconfigured. Backend coverage is available through `make -C backend test/coverage`.
- Docker and the Compose plugin are needed only for the optional container setup.
