# Full-stack Calculator

An expression calculator being built with React, TypeScript, Vite, and a stateless Go REST API. The planned keypad supports addition, subtraction, multiplication, division, exponentiation, square root, and percentage, with parentheses and operator precedence.

**Status:** initial scaffold. Dependencies and development tooling are configured, but the frontend still shows the Vite starter screen and the calculation API is not implemented. Backend package commands currently fail because `backend/cmd/api/errors.go` and `healthcheck.go` are empty Go files. See [setup and development](#setup-and-development) to explore the frontend and [the specifications](#project-documents) for intended behavior.

## Project structure

| Directory                                    | Purpose                                                             |
| -------------------------------------------- | ------------------------------------------------------------------- |
| [frontend/](frontend/)                       | React UI, API client, styles, and frontend tooling                  |
| [frontend/tests/unit/](frontend/tests/unit/) | Planned Vitest tests; currently a placeholder                       |
| [frontend/tests/e2e/](frontend/tests/e2e/)   | Playwright configuration's test location; currently example tests   |
| [backend/](backend/)                         | Go entry point, module, Makefile, and placeholder internal packages |
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

The intended development command is:

```bash
make -C backend run/api
```

It currently stops with `expected 'package', found 'EOF'` because `errors.go` and `healthcheck.go` are empty. The existing `main.go` only prints `Hello world!`; even after the placeholders are addressed, HTTP startup and routing still need implementation. No backend address or environment-variable configuration has been defined.

`make -C backend help` lists the available targets. `make -C backend build/api` targets `backend/bin/api` and is subject to the same compilation blocker.

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

Backend test, audit, build, and coverage commands are blocked by the empty Go files. No backend tests exist yet. Once implemented, the coverage target writes `backend/coverage/coverage.out` and `backend/coverage/coverage.html`. Race-enabled checks require a C compiler. Staticcheck is managed through Go's tool directive at module version `v0.8.1` and invoked by the audit target; no separate global installation is needed.

## API

The [OpenAPI specification](docs/openapi.yaml) defines `POST /calculate`. This contract is not implemented yet. Requests contain a complete expression; results are decimal strings.

Request body with `Content-Type: application/json`:

```json
{ "expression": "2 + 3 × 4" }
```

Expected `200` response:

```json
{ "result": "14" }
```

For `{"expression":"1 / 0"}`, the expected `422` response is:

```json
{ "error": { "code": "DIVISION_BY_ZERO", "message": "Cannot divide by zero." } }
```

Once an HTTP server is implemented, set `CALCULATOR_API_URL` to its actual origin and use this example:

```bash
curl --request POST "${CALCULATOR_API_URL:?Set this to the running backend origin}/calculate" \
  --header 'Content-Type: application/json' \
  --data '{"expression":"2 + 3 * 4"}'
```

`CALCULATOR_API_URL` is a shell variable for this example, not an implemented application setting. OpenAPI specifies `400` for invalid request envelopes, `415` for unsupported/missing media types, `422` for expression or arithmetic errors, and `500` for unexpected service failures.

## Design decisions and assumptions

- React owns keypad editing and presentation; Go independently parses, validates, and calculates every submitted expression. Routing will use `httprouter v1.3.0` alongside `net/http`.
- Preserve precision within an expression and round only the final result to at most three decimal places, with halfway values rounded away from zero. Return ordinary decimal strings without redundant fractional zeros or negative zero. The frontend displays these strings without converting them to JavaScript numbers.
- Continuing from a result uses the displayed rounded value: `1 / 3 → 0.333`, then `× 3 → 0.999`. Each API request is independent.
- OpenAPI limits literal and intermediate/final magnitudes to `2^1024 - 2^971`, without allowing silent underflow of nonzero intermediates. The numerical implementation is still undecided; a plain `float64` evaluator does not establish the required rounding guarantees.
- The planned interface accepts keypad input only and follows [DESIGN.md](frontend/docs/DESIGN.md), using Tailwind CSS, variable Manrope, and Lucide icons. It targets WCAG 2.2 AA, keyboard button activation, and a 320px minimum viewport width.
- Accounts, a database, and persistent calculation history are outside scope. Hosting is undecided; Docker packaging is optional.

## Project documents

- [Product requirements](docs/prd.md)
- [Calculator behavior](docs/calculator-behavior.md)
- [OpenAPI contract](docs/openapi.yaml)
- [Architecture](docs/architecture.md)
- [Visual design](frontend/docs/DESIGN.md)
- [Development prompts](docs/prompts.md)
