# Full-stack Calculator

[![CI](https://github.com/JhojanL/full-stack-calculator/actions/workflows/ci.yml/badge.svg?event=pull_request)](https://github.com/JhojanL/full-stack-calculator/actions/workflows/ci.yml)
[![Deploy frontend](https://github.com/JhojanL/full-stack-calculator/actions/workflows/deploy-frontend.yml/badge.svg?branch=master)](https://github.com/JhojanL/full-stack-calculator/actions/workflows/deploy-frontend.yml)
[![Deploy backend](https://github.com/JhojanL/full-stack-calculator/actions/workflows/deploy-backend.yml/badge.svg?branch=master)](https://github.com/JhojanL/full-stack-calculator/actions/workflows/deploy-backend.yml)
[![Live demo](https://img.shields.io/badge/GitHub%20Pages-Live%20demo-222222?logo=github)](https://calculator.jhojanlerma.dev)

An expression calculator built with React, TypeScript, Vite, and a stateless Go REST API. The keypad supports addition, subtraction, multiplication, division, exponentiation, square root, and percentage, with parentheses and operator precedence.

**Status:** the calculator UI, API client, and Go API are implemented, with unit and full-stack browser tests. GitHub Actions CI is configured for pull requests targeting `master`; frontend coverage reporting remains unconfigured. See [setup and development](#setup-and-development) to run both layers and [the specifications](#project-documents) for the application contract.

## Project structure

| Directory                                    | Purpose                                                           |
| -------------------------------------------- | ----------------------------------------------------------------- |
| [frontend/](frontend/)                       | React UI, API client, styles, and frontend tooling                |
| [frontend/tests/unit/](frontend/tests/unit/) | Vitest editing, syntax, and API-client tests                      |
| [frontend/tests/e2e/](frontend/tests/e2e/)   | Playwright UI and real-API integration tests                      |
| [backend/](backend/)                         | Go HTTP service, expression evaluator, tests, and build commands  |
| [.github/workflows/](.github/workflows/)     | Pull request CI and frontend/backend deployment workflows        |
| [docs/](docs/)                               | Requirements, API contract, architecture, and development prompts |

Frontend dependencies and `pnpm-lock.yaml` live in `frontend/`; Go dependencies live in `backend/go.mod` and `backend/go.sum`. There is no root pnpm workspace or root `package.json`. Empty tracked directories use `.gitkeep` placeholders.

## Setup and development

### Prerequisites

The commands below use Bash and run from the repository root. Install Git, Make, and mise, with mise activated in your shell. The checked-in [mise.toml](mise.toml) pins:

| Tool     | Version |
| -------- | ------- |
| Node.js  | 24.19.0 |
| pnpm     | 12.5.1  |
| Go       | 1.26.8  |
| Lefthook | 2.1.14  |

After reviewing the project configuration, install its pinned tools and dependencies:

```bash
mise trust
mise install
pnpm --dir frontend install --frozen-lockfile
(cd backend && go mod download)
```

Backend audit and the pre-push hook also require a C compiler for race-enabled tests and the standalone `gosec` and `govulncheck` executables on `PATH`. These two security tools are not installed or version-pinned by `mise.toml` or `backend/go.mod`; Staticcheck is managed by the Go module.

Use the existing pins; there is no need to generate a new `mise.toml`. The [prerequisites guide](docs/prerequisites.md) covers mise installation and component setup.

### Run the frontend

```bash
pnpm --dir frontend dev --host 127.0.0.1
```

Open the local URL printed by Vite. Start the [backend](#backend-entry-point) in a second terminal before calculating. Both Vite development and preview servers proxy `/calculate` to `http://127.0.0.1:4000`; `/healthcheck` is accessed directly on Go. Use the keypad to enter `2 + 3 × 4`, then activate `=` to display `14`.

To build and preview the static frontend:

```bash
pnpm --dir frontend build
pnpm --dir frontend preview --host 127.0.0.1
```

### Backend entry point

Start the HTTP service:

```bash
make -C backend run/api
```

The API listens on port `4000` by default. Use `make -C backend run/api ARGS='-port=8080'` to change it. The `ARGS` variable forwards flags to the application. No database, credentials, or environment files are needed. Ctrl+C or SIGTERM initiates a five-second graceful shutdown.

For direct browser requests from a separate frontend origin, set an explicit allowlist:

```bash
make -C backend run/api ARGS='-cors-trusted-origins=http://localhost:5173'
```

Origins must include the scheme and optional port, with no path or trailing slash. The default empty allowlist grants no cross-origin access. The current frontend uses a relative `/calculate` URL through the Vite proxy, so local setup needs no CORS flag. If you change Go's port, update both proxy targets in `frontend/vite.config.ts`. Direct cross-origin use would also require changing the client URL.

The calculation endpoint uses per-IP rate limiting by default: 2 requests per second with a burst of 4. Configure it with `-limiter-rps`, `-limiter-burst`, or disable it with `-limiter-enabled=false`. Excess requests return `429 RATE_LIMIT_EXCEEDED`. Each server process has its own buckets. Client IPs come from the connection, so requests through the same reverse proxy share a bucket; forwarded IP headers are ignored. The limiter retains at most 10,000 IPs and rejects new IPs with 429 when full. On incoming requests, it checks at most once per minute for fully refilled buckets idle for over three minutes.

`GET /healthcheck` reports `status`, `system_info.environment`, and `system_info.version`. It is exempt from rate limiting, as are CORS preflights. Set `-env=development|staging|production` (default `development`). Version comes from `internal/vcs.Version()` and may be `(devel)` for local builds or empty if build metadata is unavailable. This is a process liveness check, with no downstream probes.

```bash
curl http://localhost:4000/healthcheck
```

`make -C backend help` lists the available targets. `make -C backend build/api` produces a native binary at `backend/bin/api` and a Linux AMD64 binary at `backend/bin/linux_amd64/api`, both with `-ldflags='-s'`. Run the native binary with the same flags, for example `./backend/bin/api -port=8080`.

### Optional Docker setup

The working tree includes [Compose configuration](compose.yaml), component Dockerfiles, and an Nginx proxy. With Docker and the Compose plugin installed, run from the repository root:

```bash
docker compose up --build
```

The configuration publishes the frontend on `http://localhost:8080` and proxies `/calculate` and `/healthcheck` to the internal Go service. It uses the default API limiter, so clients behind Nginx share its connection-IP bucket. This packaging has not been runtime-verified as part of the documentation review.

### Git hooks

To enable the checked-in hooks locally:

```bash
lefthook install
```

The hooks in [lefthook.yml](lefthook.yml) run commands from each component's directory:

| Hook       | Frontend                                                                             | Backend                                                                                                                         |
| ---------- | ------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------- |
| Pre-commit | Prettier, then ESLint fixes on matching staged files; fixes are staged automatically | `gofmt -w` on staged Go files, with fixes staged automatically; then `make test` for matching staged Go/module/Makefile changes |
| Pre-push   | `pnpm validate`                                                                      | `make audit build/api`                                                                                                          |

Both pre-push commands are configured without file filters. Backend audit checks module tidiness and checksums, runs vet, Staticcheck, gosec, and govulncheck, then runs race-enabled tests. A successful audit is followed by both API builds. Backend tests and audit check the working tree; review partially staged changes before committing.

Run `make -C backend tidy` manually when needed. It tidies and verifies modules and regenerates vendored dependencies. Run `make -C backend fmt` separately to format Go source; review both commands' changes before staging. Generate coverage separately with `make -C backend test/coverage`.

### Continuous integration

[The GitHub Actions workflow](.github/workflows/ci.yml) runs only on pull requests targeting `master`. Its independent `frontend` and `backend` jobs run on Ubuntu 24.04 with 20-minute timeouts. New runs cancel earlier runs for the same PR; the workflow grants read-only access to repository contents and pins actions to commit SHAs.

- `frontend` uses the pinned Node, pnpm, and Go versions listed above, installs frontend dependencies with the frozen lockfile and Playwright Chromium with its system dependencies, then runs `pnpm --dir frontend validate:ci`. This includes frontend checks, unit tests, the build, and full-stack browser tests.
- `backend` uses the pinned Go version, installs standalone `gosec` `v2.29.0` and `govulncheck` `v1.8.0`, then runs `make -C backend audit` followed by `make -C backend build/api`. Staticcheck remains managed by the Go module.

All workflow commands run from the repository root. Coverage reports remain separate from CI. See [the architecture CI details](docs/architecture.md#commands-hooks-and-ci) for the workflow's validation status.

### Frontend deployment

The frontend deploys to GitHub Pages at **https://calculator.jhojanlerma.dev** on pushes to `master` or manual runs from `master`. See [frontend deployment](docs/frontend-deployment.md) for workflow details, Pages setup, and API/CORS configuration.

### Backend deployment

The backend deployment workflow audits and builds the Go API, deploys it over SSH to the production droplet, and restarts `calculator.service` on pushes to `master` or manual runs from `master`. The expected public API URL is **https://calculator-api.jhojanlerma.dev**. See [backend deployment](docs/backend-deployment.md) for the deployment overview and server layout.

DigitalOcean referral link:

[![DigitalOcean Referral Badge](https://web-platforms.sfo2.cdn.digitaloceanspaces.com/WWW/Badge%203.svg)](https://www.digitalocean.com/?refcode=cd2290237531&utm_campaign=Referral_Invite&utm_medium=Referral_Program&utm_source=badge)

## Command reference

Run these commands from the repository root. Setup commands are listed under [setup and development](#setup-and-development). The tables below cover every script in [frontend/package.json](frontend/package.json) and every target in [backend/Makefile](backend/Makefile).

### Frontend commands

| Command                               | Purpose                                                                          |
| ------------------------------------- | -------------------------------------------------------------------------------- |
| `pnpm --dir frontend dev`             | Start the Vite development server                                                |
| `pnpm --dir frontend build`           | Run TypeScript build-mode checks and build production assets in `frontend/dist/` |
| `pnpm --dir frontend preview`         | Preview the built frontend locally; run `build` first                            |
| `pnpm --dir frontend typecheck`       | Run TypeScript build-mode checks                                                 |
| `pnpm --dir frontend lint`            | Run ESLint                                                                       |
| `pnpm --dir frontend lint:fix`        | Run ESLint with automatic fixes                                                  |
| `pnpm --dir frontend format`          | Rewrite frontend files with Prettier                                             |
| `pnpm --dir frontend format:check`    | Check frontend formatting without rewriting files                                |
| `pnpm --dir frontend design:lint`     | Lint `frontend/docs/DESIGN.md` with designmd                                     |
| `pnpm --dir frontend test`            | Alias for `test:unit`                                                            |
| `pnpm --dir frontend test:unit`       | Run Vitest once                                                                  |
| `pnpm --dir frontend test:unit:watch` | Run Vitest in watch mode                                                         |
| `pnpm --dir frontend test:e2e`        | Run Playwright tests                                                             |
| `pnpm --dir frontend test:e2e:ui`     | Open Playwright's interactive test UI                                            |
| `pnpm --dir frontend test:e2e:report` | Open the existing HTML report in `frontend/playwright-report/html/`              |
| `pnpm --dir frontend check`           | Run typecheck, ESLint, Prettier checks, and design lint                          |
| `pnpm --dir frontend validate`        | Run check, unit tests, and build                                                 |
| `pnpm --dir frontend validate:ci`     | Run validate, then Playwright tests                                              |

### Backend commands

| Command                         | Purpose                                                                                                    |
| ------------------------------- | ---------------------------------------------------------------------------------------------------------- |
| `make -C backend help`          | List targets; also the default for `make -C backend`                                                       |
| `make -C backend run/api`       | Run the HTTP service; optionally pass application flags with `ARGS='...'`                                  |
| `make -C backend tidy`          | Run `go mod tidy`, `go mod verify`, and `go mod vendor`; updates dependencies and vendor files             |
| `make -C backend fmt`           | Format Go source with `go fmt ./...`                                                                       |
| `make -C backend test`          | Run all Go tests with `go test ./...`                                                                      |
| `make -C backend test/coverage` | Generate a coverage profile, terminal summary, and HTML report under `backend/coverage/`                   |
| `make -C backend audit`         | Check module tidiness/checksums, run vet, Staticcheck, gosec, govulncheck, and race-enabled tests          |
| `make -C backend build/api`     | Build native `backend/bin/api` and Linux AMD64 `backend/bin/linux_amd64/api` binaries with `-ldflags='-s'` |

`tidy` maintains dependencies; use `fmt` for formatting. No target runs `go fix`.

## Tests and coverage

Backend tests exercise the calculator API; frontend tests exercise editing, validation, response handling, and rendered full-stack behavior. See the [command reference](#command-reference) for individual checks and combined validation commands.

Vitest selects `frontend/tests/unit/**/*.test.ts` in a Node environment. The suites cover cursor edits, decimal entry, result reuse, duplicate/stale completions, grammar, strict response validation, transport failures, cancellation, and timeout. A DOM/component testing setup and frontend coverage provider/script have not been added.

Playwright builds and serves production assets through Vite preview and starts Go with rate limiting disabled. It defines desktop and mobile Chromium projects. Install its browser and required system dependencies before running it:

```bash
pnpm --dir frontend exec playwright install --with-deps chromium
pnpm --dir frontend test:e2e
```

System dependency installation may require administrator privileges. Tests cover real calculations, error correction, rounded continuation, keypad editing, keyboard focus, loading/reset, network retry, and 320px layout with enlarged text and long values. Delayed responses and transport failures use controlled routes. The configuration reuses servers already listening on ports 5173 and 4000; stop unrelated servers first, and ensure any reused backend has rate limiting disabled. Reports are written under `frontend/playwright-report/`; open the HTML report with `pnpm --dir frontend test:e2e:report`.

Backend tests cover calculation rules, float64 behavior, strict HTTP envelopes, CORS, rate limiting, healthchecks, panic recovery, configuration, and server lifecycle. The coverage target writes `backend/coverage/coverage.out` and `backend/coverage/coverage.html`. Race-enabled checks require a C compiler. Staticcheck is managed through Go's tool directive at module version `v0.8.1` and invoked by the audit target; no separate global installation is needed.

## API

The [OpenAPI specification](docs/openapi.yaml) defines `POST /calculate` and `GET /healthcheck`. Requests contain a complete expression; results are decimal strings.

Request body with `Content-Type: application/json`:

```json
{ "expression": "2 + 3 × 4" }
```

`200` response:

```json
{ "result": "14" }
```

For `{"expression":"1 / 0"}`, the expected `422` response is:

```json
{ "error": { "code": "DIVISION_BY_ZERO", "message": "Cannot divide by zero." } }
```

With the backend running on the default port:

```bash
curl --request POST http://localhost:4000/calculate \
  --header 'Content-Type: application/json' \
  --data '{"expression":"2 + 3 * 4"}'
```

OpenAPI specifies `400` for invalid request envelopes, `413 INVALID_REQUEST` for bodies over 64 KiB, `415` for unsupported/missing media types, `422` for expression or arithmetic errors, `429` for rate limiting, and `500` for unexpected service failures.

## Design decisions and assumptions

- React owns keypad editing and presentation; Go independently parses, validates, and calculates every submitted expression. Routing uses `httprouter v1.3.0` alongside `net/http`.
- Use Go `float64` arithmetic throughout, including `math.Sqrt` and `math.Pow`. Do not round intermediates to three places. For final magnitudes below `2^52`, apply `math.Round(value * 1000) / 1000` and fixed three-place formatting; larger float64 values have no fractional bits and are formatted without scaling. Remove trailing fractional zeros and negative zero, and never use exponent notation. The frontend displays these strings directly.
- Continuing from a result uses the displayed rounded value: `1 / 3 → 0.333`, then `× 3 → 0.999`. Each API request is independent.
- Literal conversion and every operation follow ordinary binary floating-point rounding. Exact decimal arithmetic is not promised: cancellation can lose precision, decimal ties can be affected by representation/scaling, and tiny values may underflow to zero. NaN and infinity are rejected as `NUMERIC_OUT_OF_RANGE`. Domain checks use unrounded float64 operands; integer exponents satisfy `math.Trunc(exponent) == exponent`. For example, `10^-400` yields `0`, and `1/(10^-400)` reports division by zero.
- The interface accepts keypad input only and follows [DESIGN.md](frontend/docs/DESIGN.md), using Tailwind CSS, variable Manrope, and Lucide icons. It targets WCAG 2.2 AA, keyboard button activation, and a 320px minimum viewport width.
- Parser nesting is limited to 128 levels, counting parentheses and right-hand power operands together. Excessive nesting returns `422 INVALID_EXPRESSION`. No separate token limit is imposed.
- Accounts, a database, and persistent calculation history are outside scope. The frontend deployment targets GitHub Pages at `https://calculator.jhojanlerma.dev`, with the API at `https://calculator-api.jhojanlerma.dev`; optional Docker packaging is also available.

## Project documents

The [AI development process](docs/ai-development-process.md) records the major prompts and my review of the results. I used AI to support planning and implementation, reviewed and edited the output, tested the work, and made the final technical and design decisions. Personal development skills, agent instructions, and internal reference lists are intentionally omitted.

- [Product requirements](docs/prd.md)
- [Calculator behavior](docs/calculator-behavior.md)
- [OpenAPI contract](docs/openapi.yaml)
- [Architecture](docs/architecture.md)
- [Visual design](frontend/docs/DESIGN.md)
- [Prerequisites](docs/prerequisites.md)
- [Development prompts](docs/prompts.md)
