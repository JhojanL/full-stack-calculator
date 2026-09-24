# Prerequisites with mise

These commands target Linux with Bash. Run the project commands from the repository root. This guide documents setup; `mise.toml` and application manifests have not been created yet.

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

## 3. Create mise.toml and install project tools

Enter your local checkout; replace this example path if needed:

```bash
cd ~/dev-work/projects/full-stack-calculator
```

Run this command once when initializing the tool configuration:

```bash
mise use --path mise.toml --pin node@24 go@1.26 pnpm@12
cat mise.toml
```

This installs Node.js 24.x, Go 1.26.x, and pnpm 10.x and writes their resolved exact versions under `[tools]` in the root `mise.toml`. Node and Go follow the [architecture](architecture.md); pnpm 10 is the initial package-manager choice. `--path` explicitly selects the project file, and `--pin` records concrete versions rather than moving version ranges. See the [mise use reference](https://mise.jdx.dev/cli/use.html).

Keep the generated `mise.toml` in version control when you initialize Git. Re-running this selection command later can change the pins; use the next section for routine installation.

## 4. Install from an existing mise.toml

For another machine or a fresh checkout, review the project configuration, then run from the repository root:

```bash
mise trust
mise install
```

This installs the versions already recorded in the configuration. See [mise getting started](https://mise.jdx.dev/getting-started.html).

## 5. Verify the tools

```bash
mise ls
mise exec -- node --version
mise exec -- pnpm --version
mise exec -- go version
mise doctor
```

Confirm Node reports `v24.x`, Go reports `go1.26.x`, and pnpm reports `12.x`, with exact versions matching `mise.toml`. `mise exec` uses the project tools even without shell activation.

## Later implementation prerequisites

- React, Vite, TypeScript, testing libraries, and JavaScript quality tools will be installed through the planned pnpm manifests.
- Staticcheck and gosec versions still need to be selected when backend checks are configured. `gofmt` and `go vet` come with Go.
- Go's planned race-detector checks require a C compiler. On Ubuntu or Debian, install the build tools with `sudo apt install -y build-essential`. See [Go race detector requirements](https://go.dev/doc/articles/race_detector#Requirements).
- Playwright browser installation belongs to the later end-to-end test setup.

Application installation, development, and test commands will be documented once their manifests and scripts exist.
