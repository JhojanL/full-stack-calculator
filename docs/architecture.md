# Architecture

**Status:** the Go API, React calculator, and API integration are implemented. Unit tests, full-stack Playwright scenarios, and a GitHub Actions CI workflow exist; frontend coverage remains unconfigured. Accessibility requirements below are targets, not a certification of conformance.

## Scope and sources

One repository contains a React SPA and one stateless Go REST service. React owns expression editing and presentation; Go independently parses, validates, and evaluates complete expressions. No accounts, database, persistent history, or memory registers are required.

- [PRD](prd.md): scope and acceptance criteria, including all seven operations.
- [Calculator behavior](calculator-behavior.md): keypad, editing transitions, arithmetic rules, messages, and accessibility.
- [OpenAPI contract](openapi.yaml): authoritative request/response schemas, grammar, status codes, error ordering, and numerical assumptions.
- [Backend Makefile](../backend/Makefile): backend development, quality, and build commands.
- [Frontend manifest](../frontend/package.json): declared dependencies and available frontend commands.
- [Design specification](../frontend/docs/DESIGN.md): layout, tokens, typography, and icons.

Keep architecture decisions consistent with these contracts. Implementation gaps below are work remaining, not additional product behavior.

## Stack and dependency ownership

Versions below are declared manifest ranges or tool pins, not claims about the latest releases or verified runtime compatibility.

| Area               | Declared choice                                                                                   |
| ------------------ | ------------------------------------------------------------------------------------------------- |
| UI                 | React and React DOM `^19.2.8`, TypeScript `~6.0.2`, Vite `^8.3.0`                                 |
| Styling            | Tailwind CSS and its Vite plugin `^4.3.3`                                                         |
| Icons and fonts    | Lucide React `^1.47.0`, variable Manrope `^5.3.0`                                                 |
| Runtime validation | Zod `^4.6.5`                                                                                      |
| Frontend testing   | Vitest `^5.0.1`, Playwright `^1.63.0`                                                             |
| Frontend quality   | ESLint `^10.10.0`, typescript-eslint `^8.69.0`, Prettier `^3.9.8`                                 |
| Backend            | Go `1.26.8` in `backend/go.mod` and `mise.toml`; HTTP/JSON through `net/http` and `encoding/json` |
| Development        | Node `24.19.0`, pnpm `12.5.1`, Lefthook `2.1.14` in `mise.toml`                                   |

Frontend dependencies and their lockfile live in `frontend/`. Go dependencies and tools live in `backend/go.mod` and `backend/go.sum`. Use `github.com/julienschmidt/httprouter` `v1.3.0` for routing alongside `net/http`; `POST /calculate` is implemented. Staticcheck is registered through the Go tool directive as `honnef.co/go/tools/cmd/staticcheck`, with its module pinned to `v0.8.1`, and runs through `go tool staticcheck ./...` from `backend/` as part of `make audit`. Audit also invokes standalone `gosec` and `govulncheck` executables from `PATH`; their installation and versions are not managed by the module or mise configuration.

Preserve the module identity and use explicit application dependencies. Standard-library imports should follow implementation needs; examples from other projects do not prescribe this service's file layout or require metrics, query parsing, background tasks, or extra helper packages. A database, ORM, authentication framework, and frontend global state library are unnecessary for this scope.

There is currently no root `package.json`, pnpm workspace, or root JavaScript lockfile. Do not assume root pnpm scripts exist.

## Repository boundaries

The frontend and backend responsibilities are:

| Path                           | Responsibility                                                                                |
| ------------------------------ | --------------------------------------------------------------------------------------------- |
| `frontend/src/components/`     | Labeled display, native keypad buttons, and icons; `App.tsx` assembles the shell              |
| `frontend/src/hooks/`          | Reducer ownership and request lifecycle                                                       |
| `frontend/src/lib/`            | Pure state reducer, submission validation, response schemas, and API client                   |
| `frontend/src/types/`          | Shared action and state contracts; infer transport types from runtime schemas where practical |
| `frontend/src/styles/`         | Tailwind entry point and semantic design tokens                                               |
| `frontend/tests/unit/`         | Vitest tests for pure frontend behavior and API boundaries                                    |
| `frontend/tests/e2e/`          | Playwright tests of rendered behavior and full-stack calculations                             |
| `backend/cmd/api/`             | Application assembly, server startup, logging, and graceful shutdown                          |
| `backend/internal/config/`     | Runtime configuration and validation                                                          |
| `backend/internal/httpapi/`    | Router, middleware, strict JSON decoding, handlers, and response mapping                      |
| `backend/internal/calculator/` | Lexer, parser, numerical evaluation, final formatting, and domain errors                      |
| `docs/`                        | Product, behavior, API, architecture, and development documentation                           |

Keep Go tests beside their packages as `_test.go` files. Domain code must not depend on HTTP responses or UI state. Keep handlers thin: decode, call the evaluator, map the outcome, and encode JSON. Introduce interfaces only where a concrete dependency needs substitution.

## Request and response flow

1. The user builds an expression through the keypad. React permits incomplete editing and checks submission syntax when `=` is activated.
2. For a valid submission, the client sends `POST /calculate` with `Content-Type: application/json` and a body such as `{"expression":"2 + 3 × 4"}`.
3. Go checks the media type and request shape, validates the complete expression, and evaluates it independently of frontend validation.
4. Success returns `200` with `{"result":"14"}`. Errors use `{"error":{"code":"DIVISION_BY_ZERO","message":"Cannot divide by zero."}}`, with the status determined by the contract.
5. The client validates the response at runtime and displays the returned decimal string directly. It does not recompute or re-round the result with JavaScript `Number`.

| HTTP status | Meaning                                                                                                                               |
| ----------- | ------------------------------------------------------------------------------------------------------------------------------------- |
| `400`       | `INVALID_REQUEST`: malformed JSON, duplicate member names, non-object body, missing/non-string/null expression, or unknown properties |
| `413`       | `INVALID_REQUEST`: JSON body exceeds 64 KiB                                                                                           |
| `415`       | `UNSUPPORTED_MEDIA_TYPE`: missing or unsupported Content-Type; JSON with UTF-8 charset is accepted                                    |
| `422`       | Expression validation or evaluation failure, including empty expressions and every documented domain error                            |
| `429`       | `RATE_LIMIT_EXCEEDED`: depleted per-IP bucket or full client map                                                                      |
| `500`       | `INTERNAL_ERROR`: unexpected service failure with a generic user message                                                              |

Use the exact error codes and English messages from OpenAPI. Decode exactly one object, reject duplicate keys and extra properties, and require complete body consumption. Ordinary struct decoding alone is insufficient to enforce all these constraints.

For `POST /calculate`, apply per-IP rate limiting first. Preserve validation order after admission: media type, body size, request shape, empty expression, unsupported syntax/invalid literals, parentheses, then grammar. Only then evaluate; when multiple evaluation errors exist, report the first in left-child-before-right-child traversal. Log internal failures without exposing implementation details in responses.

## Parsing and numerical evaluation

Implement the OpenAPI EBNF with a dedicated lexer/parser, never language-level `eval`. Normalize documented operator aliases, retain token boundaries across allowed whitespace, and require complete token consumption. Do not insert implicit multiplication or repair missing values. Scientific notation is unsupported; malformed numeric literals and unsupported operations have distinct error codes.

The syntax tree encodes percentage before powers, powers before leading minus, then multiplication/division, then addition/subtraction. Powers associate right to left. Square root consumes its next number or parenthesized expression. These rules must preserve examples such as `2^3^2 → 512`, `-2^2 → -4`, `2^-2 → 0.25`, and `√9% → 0.03`. Percentage always divides by 100, including in `200 + 10% → 200.1`.

Check domains on evaluated, unrounded operands: reject zero denominators, negative square roots, `0^0`, zero to a negative power, and negative bases with non-integer exponents. An exponent is an integer when `math.Trunc(exponent) == exponent` on the evaluated `float64` value.

The evaluator uses Go `float64` throughout, with `math.Sqrt` and `math.Pow`. Literals that parse outside the finite float64 range and operations producing NaN or infinity return `NUMERIC_OUT_OF_RANGE`. Binary representation errors, cancellation, and underflow to zero follow ordinary float64 behavior. Exact decimal arithmetic and preservation of tiny nonzero intermediates are not guaranteed.

Intermediate values are not rounded to three decimal places. Final values below magnitude `2^52` use `math.Round(value * 1000) / 1000`, then fixed three-place formatting and removal of redundant zeros. Larger values have no fractional bits and are formatted without scaling to avoid overflow. Results never use scientific notation or negative zero. Binary scaling errors can affect results near decimal ties. This intentionally simple numerical policy is also specified in OpenAPI and the README.

## Frontend state and API boundary

`useCalculator` owns local React state through `calculatorReducer` and submits pending work through the API client. Model editing, pending, success, and error states with a discriminated union. Keep the expression, insertion position, displayed result string, and request identity explicit rather than relying on unrelated boolean flags. Typed button actions describe intent; components receive typed data and callbacks.

Store the editable expression without decorative spacing. Move the insertion position by one digit or symbol; render spacing separately. Button actions enforce decimal-entry rules, while submission validation follows the documented grammar. The browser never evaluates arithmetic.

| Event                                                | State transition                                                                                |
| ---------------------------------------------------- | ----------------------------------------------------------------------------------------------- |
| Clear                                                | Empty expression, initial insertion position, result `0`, no error; invalidate pending work     |
| `=` while editing or after failure                   | Validate, then submit; preserve the expression on failure so `=` can retry                      |
| `=` while pending or immediately after success       | Suppress duplicate submissions; repeated success leaves the result unchanged                    |
| Digit, `.`, `(`, or `√` immediately after success    | Start a new expression and clear the previous result                                            |
| Arithmetic operator or `%` immediately after success | Seed the expression from the returned rounded string, wrapping a negative result in parentheses |
| Arrow or backspace after success                     | Return to the original expression; clear the result only when content actually changes          |
| Closing parenthesis after success                    | Edit the original expression                                                                    |
| Actual edit during pending/error/result state        | Clear old feedback and invalidate any pending request                                           |

Cursor movement alone does not change the calculation or clear feedback. After `1 / 3` returns `"0.333"`, continuing with `× 3` submits `0.333 × 3` and produces `"0.999"`.

The API client owns fetch, response validation, and cancellation. Use Zod schemas matching the strict success/error envelopes, decimal-result format, and error-code enumeration; TypeScript types alone do not validate network data. Treat malformed responses, unexpected statuses, and network failures as “Could not calculate. Try again.”

The API client enforces a 10-second deadline. Abort on timeout, superseding edits, clear, or unmount, and also check request identity before committing either success or failure. Aborting alone does not prevent an already completed stale response from updating state. Editing and clear remain usable while waiting; stale failures must not replace newer results either.

## Styling and accessibility

Use the approved 25-button layout and values from [DESIGN.md](../frontend/docs/DESIGN.md). Tailwind's Vite plugin and the `@import "tailwindcss"` entry point already exist. Define semantic tokens in CSS with `@theme` and reusable button variants for neutral, arithmetic, and editing controls. There is no token generator or generated theme stylesheet currently; keep authored tokens aligned with the design source.

The stylesheet imports `@fontsource-variable/manrope` for self-hosted Manrope, using the `Manrope Variable` family and the design-specified weights. Use Lucide icons and the specified local exponent SVG, with decorative graphics hidden from assistive technology.

Native buttons provide Tab/Shift+Tab navigation and Enter/Space activation. The expression display must not accept typing, paste, or dropped text. Expose the expression and insertion position accessibly, retain focus on the activated button, and announce cursor changes, results, and errors without moving focus.

Show “Calculating…” in the status area. Errors replace the number in the Result area with wrapping soft-red text while retaining the Result label and expression. Meet WCAG 2.2 AA targets from the behavior specification, including contrast, visible focus, and text feedback. Preserve at least 44 × 44 CSS pixel targets, all controls at 320px width, 200% text enlargement, internal scrolling for long values, and reduced-motion preferences.

## Verification strategy

Verify observable contracts at the smallest useful boundary. Keep pure dependencies real; substitute network failures or delayed responses only where the scenario requires control. Do not duplicate the entire arithmetic suite in browser tests.

| Boundary                       | Required evidence                                                                                                                                                                                                   |
| ------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Go calculator package          | Table-driven tests for precedence, associativity, all seven operations, aliases, grammar rejection, error ordering, domains, range, float64 underflow, and expected display strings                                 |
| Go routed HTTP handler         | `net/http/httptest` requests through the real router/middleware; assert status, JSON content type, exact envelopes/codes/messages, strict request decoding, and representative successful evaluations               |
| Frontend pure state and client | Vitest scenarios for insertion/deletion, decimal rules, post-result transitions, negative-result reuse, validation, schema rejection, retries, timeout, duplicate prevention, and stale success/failure suppression |
| Rendered UI                    | Playwright checks for keypad-only editing, focus and keyboard activation, visible loading/errors, accessible names/announcements, 320px layout, enlarged text, and long-value scrolling                             |
| Full stack                     | Built frontend against the real Go service: precedence calculation, arithmetic error and correction, and continuation from a rounded result                                                                         |

Go test applications must own their router and dependencies rather than mutate process-global state. Use isolated server lifetimes and register cleanup. Vitest tests should import its APIs explicitly, await asynchronous assertions, restore spies/timers, and assert outcomes rather than reducer internals or snapshots of implementation details. Use non-watch mode for automated runs. Review screen-reader feedback and visual accessibility in addition to automated assertions.

Current test coverage and gaps are explicit:

- Vitest selects only `tests/unit/**/*.test.ts` in the Node environment; the suites cover editing, grammar, request identities, strict responses, cancellation, and timeout. React Testing Library, a DOM environment, and a Vitest coverage provider are not declared. Pure state/client tests fit the current configuration; component tests would require deliberate dependencies and configuration, including `.tsx` selection.
- Playwright builds the frontend, serves it through Vite preview, and starts Go with the limiter disabled. Desktop and mobile Chromium scenarios cover real calculations, editing, keyboard focus, delayed/reset requests, retry, and narrow/enlarged layouts. Existing servers on ports 5173 and 4000 are reused; their configuration must match the test setup. Automated assertions do not replace manual screen-reader and visual review.
- Backend calculator and HTTP tests generate coverage through `make -C backend test/coverage`. Frontend coverage remains separate work; do not claim both reports are available until its tooling exists.

For implementation changes, run focused tests first, then applicable quality gates. Empty suites, skipped scenarios, and unrelated example tests are not evidence that calculator requirements pass. For documentation-only changes, verify formatting, links, and consistency with the source contracts without adding tests that merely match prose.

## Commands, hooks, and CI

These commands exist now; they describe available entry points, not completed calculator verification. The [README command reference](../README.md#command-reference) lists every frontend script and backend target.

| Command from repository root      | Purpose                                                                                    |
| --------------------------------- | ------------------------------------------------------------------------------------------ |
| `pnpm --dir frontend dev`         | Start Vite                                                                                 |
| `make -C backend run/api`         | Run the HTTP server on port 4000; forward flags with `ARGS='...'`                          |
| `pnpm --dir frontend check`       | TypeScript build-mode checks, ESLint, Prettier check, and design lint                      |
| `pnpm --dir frontend test:unit`   | Vitest in non-watch mode                                                                   |
| `pnpm --dir frontend build`       | TypeScript check and Vite production build                                                 |
| `pnpm --dir frontend validate`    | Frontend check, unit tests, and build                                                      |
| `pnpm --dir frontend validate:ci` | Frontend validate followed by Playwright                                                   |
| `pnpm --dir frontend test:e2e`    | Playwright using its current frontend configuration                                        |
| `make -C backend test`            | Go tests                                                                                   |
| `make -C backend test/coverage`   | Backend text/HTML coverage under `backend/coverage/`                                       |
| `make -C backend audit`           | Module tidiness/verification, vet, Staticcheck, gosec, govulncheck, and race-enabled tests |
| `make -C backend build/api`       | Build native `backend/bin/api` and Linux AMD64 `backend/bin/linux_amd64/api`               |

The backend also exposes `make -C backend help` (the default target), `make -C backend tidy`, and `make -C backend fmt`. Tidy runs module tidying, checksum verification, and vendoring; formatting is a separate `go fmt ./...` recipe. Tidy does not format source or run `go fix`. Both maintenance targets remain manual.

Root Lefthook runs Prettier and ESLint fixes on matching staged frontend files, then gofmt on matching staged Go files, staging those fixes automatically. Matching staged backend Go/module/Makefile changes also run `make test`. Pre-push runs frontend `validate` and backend `make audit build/api` without file filters. Audit requires a C compiler for race-enabled tests plus gosec and govulncheck on `PATH`; it runs those scanners after vet and Staticcheck and before the race-enabled tests. Both builds use `-ldflags='-s'`; the second sets `GOOS=linux GOARCH=amd64`. A failed audit prevents the subsequent builds in the hook.

[The CI workflow](../.github/workflows/ci.yml) runs only on pull requests targeting `master`. It grants `contents: read` permission and groups concurrent runs by workflow and PR number, cancelling an earlier run when a newer run starts for the same PR. Actions are pinned to immutable commit SHAs.

Two independent jobs run on `ubuntu-24.04`, each with a 20-minute timeout. All workflow commands run from the repository root:

- `frontend` installs Node `24.19.0`, pnpm `12.5.1`, and Go `1.26.8`, matching `mise.toml`. It runs `pnpm --dir frontend install --frozen-lockfile`, installs Chromium and its system dependencies with `pnpm --dir frontend exec playwright install --with-deps chromium`, then runs `pnpm --dir frontend validate:ci`. That script includes frontend checks, unit tests, the production build, and Playwright tests; Playwright starts the Go API for full-stack scenarios.
- `backend` installs Go `1.26.8` and the standalone tools `gosec` `v2.29.0` and `govulncheck` `v1.8.0`. It runs `make -C backend audit`, followed by `make -C backend build/api` only if the audit succeeds. Staticcheck remains managed by `backend/go.mod` and runs through the audit target.

Coverage reports remain outside this workflow. The workflow configuration has been checked with Actionlint; execution on GitHub Actions remains unverified. The README owns setup and runnable command details.

## Runtime and deployment

The client posts to the relative `/calculate` URL. Vite development and preview both proxy that path to `http://127.0.0.1:4000`; neither proxies `/healthcheck`. Static hosting must provide equivalent API routing. Separate-origin hosting would require a client URL change and an explicit backend CORS allowlist; no configurable API base URL currently exists.

The Go process owns server timeouts, request cancellation, panic recovery, and a five-second graceful shutdown drain. Request bodies are limited to 64 KiB, returning `413 INVALID_REQUEST` when exceeded. Parser nesting is limited to 128 levels, counting parenthesized expressions and recursive right-hand power operands together; exceeding this returns `422 INVALID_EXPRESSION`. There is no separate token or adaptive computation budget. The body limit bounds total input size, and tree evaluation checks request cancellation at each node.

The hosting target remains an open decision. The working tree includes optional Docker packaging: Nginx serves the built frontend on port 8080 and proxies `/calculate` and `/healthcheck` to Go on the Compose network. The backend keeps its default limiter, so clients through that proxy share a bucket. Docker runtime behavior has not been verified by this documentation review. Continuous deployment remains unconfigured. No persistence infrastructure is required.

## Rate limiting and healthcheck

`POST /calculate` uses a per-process, mutex-protected IP map with `golang.org/x/time/rate` token buckets. Defaults match the supplied example: enabled, 2 requests per second, burst 4; flags configure these values. Exhausted buckets return `429 RATE_LIMIT_EXCEEDED` before validation. The connection IP supplies identity; forwarded headers are ignored, so a reverse proxy shares a bucket for its clients. The map holds at most 10,000 IPs and rejects new identities with 429 when full.

Cleanup runs within requests at most once per minute, removing fully refilled buckets idle for over three minutes. Retaining depleted buckets prevents extra bursts at very low configured rates. There is no cleanup goroutine to stop. CORS runs before the limiter, and preflights do not consume tokens.

`GET /healthcheck` bypasses rate limiting and returns process availability, the configured environment, and the version obtained in `main.go` through `internal/vcs.Version()`. It does not probe dependencies. Environment defaults to `development`; `staging` and `production` are also accepted. Build version may be `(devel)` or empty, matching the existing version helper.
