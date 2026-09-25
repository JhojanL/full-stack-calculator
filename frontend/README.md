# Calculator frontend

React, TypeScript, and Vite frontend for the [full-stack calculator](../README.md). It provides keypad-only expression editing, syntax checks, accessible feedback, and requests to the Go evaluator. Arithmetic is performed by Go; returned decimal strings are displayed unchanged.

## Run locally

Install the tools pinned in [mise.toml](../mise.toml) using the [prerequisites guide](../docs/prerequisites.md). Run these commands from the repository root:

```bash
pnpm --dir frontend install --frozen-lockfile
make -C backend run/api
```

In a second terminal, also from the repository root:

```bash
pnpm --dir frontend dev --host 127.0.0.1
```

Open the URL printed by Vite. Its development and preview proxies forward `/calculate` to `http://127.0.0.1:4000`. Both processes must be running to calculate; no frontend environment variables are required.

## Code map

- `src/components/`: read-only display, keypad, and exponent icon.
- `src/hooks/useCalculator.ts`: state ownership and request cancellation.
- `src/lib/calculator.ts`: pure editor reducer and request identity checks.
- `src/lib/validation.ts`: submission syntax checks without arithmetic.
- `src/lib/api.ts`: fetch, 10-second deadline, and strict Zod response validation.
- `src/types/calculator.ts`: shared state and action contracts.
- `src/styles/global.css`: Tailwind tokens and calculator styles.

## Checks

From the repository root:

```bash
pnpm --dir frontend check
pnpm --dir frontend test:unit
pnpm --dir frontend build
```

`validate` combines these checks. Unit tests run in Node and cover editing, syntax, and the API client. There is no DOM unit-test environment or frontend coverage provider configured.

For browser tests, install Chromium once, then run Playwright:

```bash
pnpm --dir frontend exec playwright install --with-deps chromium
pnpm --dir frontend test:e2e
```

Playwright builds and serves production assets with Vite preview, starts Go with rate limiting disabled, and runs desktop/mobile Chromium scenarios. It reuses existing servers on ports 5173 and 4000, so stop unrelated processes first and ensure any reused backend has its limiter disabled. Reports are written to `playwright-report/`.

See the [root command reference](../README.md#frontend-commands) for all scripts, [architecture](../docs/architecture.md) for state and request flow, [behavior contract](../docs/calculator-behavior.md) for editing rules, and [DESIGN.md](docs/DESIGN.md) for visual requirements.
