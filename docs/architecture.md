# Architecture

**Status:** target architecture with an initial scaffold. The frontend currently renders the Vite starter screen, and the Go entry point prints a greeting. The calculator and API described below remain to be implemented.

## Scope and sources

One repository contains a React SPA and one stateless Go REST service. React owns expression editing and presentation; Go independently parses, validates, and evaluates complete expressions. No accounts, database, persistent history, or memory registers are required.

- [PRD](prd.md): scope and acceptance criteria, including all seven operations.
- [Calculator behavior](calculator-behavior.md): keypad, editing transitions, arithmetic rules, messages, and accessibility.
- [OpenAPI contract](openapi.yaml): authoritative request/response schemas, grammar, status codes, error ordering, and numerical assumptions.
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

Frontend dependencies and their lockfile live in `frontend/`. Go dependencies and tools live in `backend/go.mod` and `backend/go.sum`. Use `github.com/julienschmidt/httprouter` `v1.3.0` for routing alongside `net/http`; routes are not yet implemented. Staticcheck is registered through the Go tool directive as `honnef.co/go/tools/cmd/staticcheck`, with its module pinned to `v0.8.1`, and runs through `go tool staticcheck ./...` from `backend/` as part of `make audit`.

Preserve the module identity and use explicit application dependencies. Standard-library imports should follow implementation needs; examples from other projects do not prescribe this service's file layout or require metrics, query parsing, background tasks, or extra helper packages. A database, ORM, authentication framework, and frontend global state library are unnecessary for this scope.

There is currently no root `package.json`, pnpm workspace, or root JavaScript lockfile. Do not assume root pnpm scripts exist.

## Repository boundaries

The directories exist; most calculator-specific directories are placeholders. Their intended responsibilities are:

| Path                           | Responsibility                                                                                |
| ------------------------------ | --------------------------------------------------------------------------------------------- |
| `frontend/src/components/`     | Calculator shell, labeled display, native keypad buttons, and icons                           |
| `frontend/src/hooks/`          | Calculator state transitions and request lifecycle                                            |
| `frontend/src/lib/`            | Pure editing helpers, submission validation, response schemas, and API client                 |
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
| `415`       | `UNSUPPORTED_MEDIA_TYPE`: missing or unsupported Content-Type; JSON with UTF-8 charset is accepted                                    |
| `422`       | Expression validation or evaluation failure, including empty expressions and every documented domain error                            |
| `500`       | `INTERNAL_ERROR`: unexpected service failure with a generic user message                                                              |

Use the exact error codes and English messages from OpenAPI. Decode exactly one object, reject duplicate keys and extra properties, and require complete body consumption. Ordinary struct decoding alone is insufficient to enforce all these constraints.

Preserve validation order: media type, request shape, empty expression, unsupported syntax/invalid literals, parentheses, then grammar. Only then evaluate; when multiple evaluation errors exist, report the first in left-child-before-right-child traversal. Log internal failures without exposing implementation details in responses.

## Parsing and numerical evaluation

Implement the OpenAPI EBNF with a dedicated lexer/parser, never language-level `eval`. Normalize documented operator aliases, retain token boundaries across allowed whitespace, and require complete token consumption. Do not insert implicit multiplication or repair missing values. Scientific notation is unsupported; malformed numeric literals and unsupported operations have distinct error codes.

The syntax tree encodes percentage before powers, powers before leading minus, then multiplication/division, then addition/subtraction. Powers associate right to left. Square root consumes its next number or parenthesized expression. These rules must preserve examples such as `2^3^2 → 512`, `-2^2 → -4`, `2^-2 → 0.25`, and `√9% → 0.03`. Percentage always divides by 100, including in `200 + 10% → 200.1`.

Check domains on evaluated, unrounded operands: reject zero denominators, negative square roots, `0^0`, zero to a negative power, and negative bases with non-integer exponents. An exponent's mathematical value determines whether it is an integer.

The contract bounds the magnitude of every literal and intermediate/final value by `2^1024 - 2^971`. This is a magnitude limit, not permission to use binary64 rounding throughout. Preserve nonzero intermediates without silent underflow and retain enough precision to round the final mathematical result correctly.

Round once, to at most three decimal places, with halfway values away from zero. Serialize ordinary decimal notation without redundant fractional zeros, exponents, or negative zero: `1.2345 → "1.235"`, `-1.2345 → "-1.235"`, and `-0.0004 → "0"`.

**Open implementation decision:** select and verify a numerical strategy for exact decimal/rational operations plus square roots and fractional powers. A plain `float64` evaluator or an arbitrary fixed precision does not establish these guarantees. Evaluate candidate implementations against tie cases, cancellation, domain checks, and range boundaries before choosing a dependency. Document the resulting precision assumptions in the README without silently weakening OpenAPI.

## Frontend state and API boundary

Use local React state, with a reducer for explicit keypad transitions and a hook for network ownership. Model editing, pending, success, and error states with a discriminated union. Keep the expression, insertion position, displayed result string, and request identity explicit rather than relying on unrelated boolean flags. Typed button actions describe intent; components receive typed data and callbacks.

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

Retain the proposed 10-second client deadline as an architectural choice. Abort on timeout, superseding edits, clear, or unmount, and also check request identity before committing either success or failure. Aborting alone does not prevent an already completed stale response from updating state. Editing and clear remain usable while waiting; stale failures must not replace newer results either.

## Styling and accessibility

Use the approved 25-button layout and values from [DESIGN.md](../frontend/docs/DESIGN.md). Tailwind's Vite plugin and the `@import "tailwindcss"` entry point already exist. Define semantic tokens in CSS with `@theme` and reusable button variants for neutral, arithmetic, and editing controls. There is no token generator or generated theme stylesheet currently; keep authored tokens aligned with the design source.

Use the declared `@fontsource-variable/manrope` package for self-hosted Manrope. The design document's static `@fontsource/manrope` import examples differ from the manifest; preserve its intended typeface and weights using the variable package. Use Lucide icons and the specified local exponent SVG, with decorative graphics hidden from assistive technology.

Native buttons provide Tab/Shift+Tab navigation and Enter/Space activation. The expression display must not accept typing, paste, or dropped text. Expose the expression and insertion position accessibly, retain focus on the activated button, and announce cursor changes, results, and errors without moving focus.

Show “Calculating…” in the status area. Errors replace the number in the Result area with wrapping soft-red text while retaining the Result label and expression. Meet WCAG 2.2 AA targets from the behavior specification, including contrast, visible focus, and text feedback. Preserve at least 44 × 44 CSS pixel targets, all controls at 320px width, 200% text enlargement, internal scrolling for long values, and reduced-motion preferences.

## Verification strategy

Verify observable contracts at the smallest useful boundary. Keep pure dependencies real; substitute network failures or delayed responses only where the scenario requires control. Do not duplicate the entire arithmetic suite in browser tests.

| Boundary                       | Required evidence                                                                                                                                                                                                   |
| ------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Go calculator package          | Table-driven tests for precedence, associativity, all seven operations, aliases, grammar rejection, error ordering, domains, range, underflow preservation, and exact rounded strings                               |
| Go routed HTTP handler         | `net/http/httptest` requests through the real router/middleware; assert status, JSON content type, exact envelopes/codes/messages, strict request decoding, and representative successful evaluations               |
| Frontend pure state and client | Vitest scenarios for insertion/deletion, decimal rules, post-result transitions, negative-result reuse, validation, schema rejection, retries, timeout, duplicate prevention, and stale success/failure suppression |
| Rendered UI                    | Playwright checks for keypad-only editing, focus and keyboard activation, visible loading/errors, accessible names/announcements, 320px layout, enlarged text, and long-value scrolling                             |
| Full stack                     | Built frontend against the real Go service: precedence calculation, arithmetic error and correction, and continuation from a rounded result                                                                         |

Go test applications must own their router and dependencies rather than mutate process-global state. Use isolated server lifetimes and register cleanup. Vitest tests should import its APIs explicitly, await asynchronous assertions, restore spies/timers, and assert outcomes rather than reducer internals or snapshots of implementation details. Use non-watch mode for automated runs. Review screen-reader feedback and visual accessibility in addition to automated assertions.

Current testing gaps are explicit:

- Vitest selects only `tests/unit/**/*.test.ts` in the Node environment; the unit directory contains no tests. React Testing Library, a DOM environment, and a Vitest coverage provider are not declared. Pure state/client tests fit the current configuration; component tests would require deliberate dependencies and configuration, including `.tsx` selection.
- Playwright currently starts the Vite development server and runs example tests against the Playwright website. It does not start Go or verify this calculator. Replace those examples and configure both application processes for full-stack acceptance testing.
- Backend coverage commands exist, but calculator tests and implementation are absent. Add frontend coverage tooling and a reproducible command before claiming both coverage deliverables are available.

For implementation changes, run focused tests first, then applicable quality gates. Empty suites, skipped scenarios, and unrelated example tests are not evidence that calculator requirements pass. For documentation-only changes, verify formatting, links, and consistency with the source contracts without adding tests that merely match prose.

## Commands, hooks, and CI

These commands exist now; they describe available entry points, not completed calculator verification.

| Command from repository root      | Purpose                                                                |
| --------------------------------- | ---------------------------------------------------------------------- |
| `pnpm --dir frontend dev`         | Start Vite                                                             |
| `make -C backend run/api`         | Run the current Go entry point; it is not yet an HTTP server           |
| `pnpm --dir frontend check`       | TypeScript build-mode checks, ESLint, and read-only Prettier check     |
| `pnpm --dir frontend test:unit`   | Vitest in non-watch mode                                               |
| `pnpm --dir frontend build`       | TypeScript check and Vite production build                             |
| `pnpm --dir frontend validate`    | Frontend check, unit tests, and build                                  |
| `pnpm --dir frontend validate:ci` | Frontend validate followed by Playwright                               |
| `pnpm --dir frontend test:e2e`    | Playwright using its current frontend configuration                    |
| `make -C backend test`            | Go tests                                                               |
| `make -C backend test/coverage`   | Backend text/HTML coverage under `backend/coverage/`                   |
| `make -C backend audit`           | Module tidiness/verification, vet, Staticcheck, and race-enabled tests |
| `make -C backend build/api`       | Build `backend/bin/api`                                                |

Root Lefthook currently auto-formats and lints staged frontend files before commits, then runs frontend `validate` before pushes. Backend hooks are not configured. `.github/workflows/` contains a placeholder; `validate:ci` is a local script, not an installed GitHub Actions workflow.

The intended CI runs on PRs to `main` and pushes to `main`: read-only formatting checks, frontend validation, backend audit, both coverage reports, both builds, and application E2E. Wire these gates after their tests/configuration exist. Use the actual manifests and lockfiles; do not introduce fictional root scripts. The README owns setup and runnable command details.

## Runtime and deployment

Serve static frontend assets and run one Go HTTP service. Prefer a shared public origin with `/calculate` routed to Go. Configure a Vite development proxy for that same path; no proxy exists in the current Vite configuration. If hosting uses separate origins, configure the API base URL and an explicit CORS origin allowlist. Keep asset paths independent of the API base URL.

The Go process owns server timeouts, request cancellation, panic recovery, and graceful shutdown. Parsing and numerical work need bounded resources as well as bounded transport input. Exact body-size, token/depth, and computation limits and their public error mapping remain to be specified in OpenAPI before implementation; do not silently add undocumented rejection behavior. A server socket timeout alone does not bound evaluator CPU work.

The hosting target and numerical implementation remain open decisions. Docker packaging and continuous deployment are optional. No persistence infrastructure is required.
