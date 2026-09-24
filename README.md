# Full-stack Calculator

An expression calculator being built with React, TypeScript, Vite, and a stateless Go REST API. The planned keypad supports addition, subtraction, multiplication, division, exponentiation, square root, and percentage, with parentheses and operator precedence.

**Status:** the stateless Go calculator API is implemented and tested. Frontend calculator UI and API integration remain separate work. See [setup and development](#setup-and-development) to run the backend and [the specifications](#project-documents) for the application contract.

## Project structure

| Directory                                    | Purpose                                                             |
| -------------------------------------------- | ------------------------------------------------------------------- |
| [frontend/](frontend/)                       | React UI, API client, styles, and frontend tooling                  |
| [frontend/tests/unit/](frontend/tests/unit/) | Planned Vitest tests; currently a placeholder                       |
| [frontend/tests/e2e/](frontend/tests/e2e/)   | Playwright configuration's test location; currently example tests   |
| [backend/](backend/)                         | Go HTTP service, expression evaluator, tests, and build commands  |
| [.github/workflows/](.github/workflows/)     | Placeholder for future GitHub Actions workflows                     |
| [docs/](docs/)                               | Requirements, API contract, architecture, and development prompts   |

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

Use the existing pins; there is no need to generate a new `mise.toml`. The older [prerequisites guide](docs/prerequisites.md) includes mise installation instructions, but its scaffold status and version-selection sections are outdated; use the configuration and commands above for this checkout.

### Run the frontend

```bash
pnpm --dir frontend dev --host 127.0.0.1
```

Open the local URL printed by Vite. The current outcome is the starter screen, not a working calculator. No API proxy is configured yet.

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

The API listens on port `4000` by default. Use `go run ./cmd/api -port=8080` from `backend/` to change it. No database, credentials, or environment files are needed. Ctrl+C or SIGTERM initiates a five-second graceful shutdown.

For direct browser requests from a separate frontend origin, set an explicit allowlist:

```bash
(cd backend && go run ./cmd/api -cors-trusted-origins='http://localhost:5173')
```

Origins must include the scheme and optional port, with no path or trailing slash. The default empty allowlist grants no cross-origin access. A shared origin with `/calculate` proxied to Go is also supported; frontend proxy configuration is separate work.

`make -C backend help` lists the available targets. `make -C backend build/api` produces `backend/bin/api`; run that binary with the same flags.

### Git hooks

To enable the checked-in hooks locally:

```bash
lefthook install
```

Pre-commit runs Prettier and ESLint fixes on staged frontend files and stages those fixes. Pre-push runs frontend `validate`, including unit tests. The empty unit suite currently prevents that validation from completing successfully. Backend hook checks and GitHub Actions workflows are not configured yet.

## Tests and coverage

Available commands are listed below. They describe configured tooling, not passing calculator coverage.

| Command from repository root      | Purpose                                                    |
| --------------------------------- | ---------------------------------------------------------- |
| `pnpm --dir frontend check`       | Type checking, ESLint, and read-only formatting checks     |
| `pnpm --dir frontend test:unit`   | Vitest in non-watch mode                                   |
| `pnpm --dir frontend validate`    | Frontend checks, unit tests, and build                     |
| `pnpm --dir frontend test:e2e`    | Playwright tests                                           |
| `pnpm --dir frontend validate:ci` | Frontend validation followed by Playwright                 |
| `make -C backend test`            | Go tests                                                   |
| `make -C backend audit`           | Module checks, go vet, Staticcheck, and race-enabled tests |
| `make -C backend test/coverage`   | Go coverage profile, terminal summary, and HTML report     |

Vitest currently selects `frontend/tests/unit/**/*.test.ts` in a Node environment, with no tests present. A DOM/component testing setup and frontend coverage provider/script have not been added.

Playwright starts Vite automatically and defines desktop and mobile Chromium projects. Install its browser and required system dependencies before running it:

```bash
pnpm --dir frontend exec playwright install --with-deps chromium
pnpm --dir frontend test:e2e
```

System dependency installation may require administrator privileges. The current example tests visit the Playwright website and require internet access; they do not test this calculator or start Go. Reports are written under `frontend/playwright-report/`; open the HTML report with `pnpm --dir frontend test:e2e:report`.

Backend tests cover calculation rules, float64 behavior, strict HTTP envelopes, CORS, panic recovery, configuration, and server lifecycle. The coverage target writes `backend/coverage/coverage.out` and `backend/coverage/coverage.html`. Race-enabled checks require a C compiler. Staticcheck is managed through Go's tool directive at module version `v0.8.1` and invoked by the audit target; no separate global installation is needed.

## API

The [OpenAPI specification](docs/openapi.yaml) defines `POST /calculate`. Requests contain a complete expression; results are decimal strings.

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

OpenAPI specifies `400` for invalid request envelopes, `413 INVALID_REQUEST` for bodies over 64 KiB, `415` for unsupported/missing media types, `422` for expression or arithmetic errors, and `500` for unexpected service failures.

## Design decisions and assumptions

- React owns keypad editing and presentation; Go independently parses, validates, and calculates every submitted expression. Routing uses `httprouter v1.3.0` alongside `net/http`.
- Use Go `float64` arithmetic throughout, including `math.Sqrt` and `math.Pow`. Do not round intermediates to three places. For final magnitudes below `2^52`, apply `math.Round(value * 1000) / 1000` and fixed three-place formatting; larger float64 values have no fractional bits and are formatted without scaling. Remove trailing fractional zeros and negative zero, and never use exponent notation. The frontend should display these strings directly.
- Continuing from a result uses the displayed rounded value: `1 / 3 → 0.333`, then `× 3 → 0.999`. Each API request is independent.
- Literal conversion and every operation follow ordinary binary floating-point rounding. Exact decimal arithmetic is not promised: cancellation can lose precision, decimal ties can be affected by representation/scaling, and tiny values may underflow to zero. NaN and infinity are rejected as `NUMERIC_OUT_OF_RANGE`. Domain checks use unrounded float64 operands; integer exponents satisfy `math.Trunc(exponent) == exponent`. For example, `10^-400` yields `0`, and `1/(10^-400)` reports division by zero.
- The planned interface accepts keypad input only and follows [DESIGN.md](frontend/docs/DESIGN.md), using Tailwind CSS, variable Manrope, and Lucide icons. It targets WCAG 2.2 AA, keyboard button activation, and a 320px minimum viewport width.
- Parser nesting is limited to 128 levels, counting parentheses and right-hand power operands together. Excessive nesting returns `422 INVALID_EXPRESSION`. No separate token limit is imposed.
- Accounts, a database, and persistent calculation history are outside scope. Hosting is undecided; Docker packaging is optional.

## Project documents

- [Product requirements](docs/prd.md)
- [Calculator behavior](docs/calculator-behavior.md)
- [OpenAPI contract](docs/openapi.yaml)
- [Architecture](docs/architecture.md)
- [Visual design](frontend/docs/DESIGN.md)
- [Development prompts](docs/prompts.md)
