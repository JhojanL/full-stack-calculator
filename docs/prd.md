# PRD: Calculator

## Objective

Build a full-stack expression calculator with a React frontend and a Go REST API. Users build expressions with an on-screen keypad and receive results calculated by the backend. Prioritize correctness, a clear interface, maintainable code, and testability.

## Functional requirements

### Operations

- **Required:** addition, subtraction, multiplication, and division.
- **Selected enhancements:** exponentiation, square root, and percentage conversion (`a / 100`).

All seven operations are included in this project's scope. Evaluate complete expressions using parentheses and operator precedence, following `calculator-behavior.md`.

### Frontend

- Provide the approved 25-button keypad, including left/right cursor navigation, backspace, and clear. Accept button input only; disable free-text entry, paste, and dropped text.
- Allow incomplete expressions while editing; validate and submit on `=`. Use `√9` for a single value and `√(9 + 5)` for a grouped expression.
- Keep the expression visible and display positive, negative, or zero results with up to three decimal places. Omit unnecessary trailing zeros, `-0`, and scientific notation; scientific notation is also unsupported in input.
- Show calculation errors, network failures, and timeouts as descriptive text in the Result area, replacing the number while preserving the expression for correction.
- Show a loading state and prevent duplicate submissions while a request is pending. Keep editing and clear available; ignore stale responses after an edit or reset.
- Support desktop and mobile layouts, including 320px viewport width. Provide keyboard access through Tab/Shift+Tab navigation and Enter/Space button activation.
- Follow `DESIGN.md` for the layout, colors, typography, and icons. Target WCAG 2.2 AA with accessible button names, visible focus, result/error announcements, and project touch targets of at least 44 × 44 CSS pixels.

### Backend

- Run as one stateless Go service exposing a REST API.
- Accept complete expressions and return results or descriptive errors in JSON, with appropriate HTTP status codes.
- Parse, validate, and evaluate expressions independently of the frontend, including missing values, invalid types, malformed expressions, and unsupported syntax.
- Reject division by zero, negative square roots, unsupported powers, and non-finite inputs or results.

Accounts, a database, and persistent calculation history are outside the initial scope.

## Technical and quality requirements

- **Frontend:** React with Vite and TypeScript 6.x.
- **Backend:** Go.
- **Development runtime:** Node.js 24 LTS; select compatible, maintained tool versions.
- Keep arithmetic logic separate from HTTP handling and presentation.
- Write clean, readable, idiomatic code with unit tests covering key behavior in both layers.
- Generate frontend and backend coverage reports.
- Use Go `float64` arithmetic, including `math.Sqrt` and `math.Pow`, throughout each expression. Apply three-place display rounding only to the final result, with scaled halfway values rounded away from zero. Normal floating-point representation errors and underflow are accepted; exact decimal arithmetic is not required. Continuing from a result uses the displayed, rounded value.
- Document numerical precision and display-rounding assumptions in the README, consistent with `calculator-behavior.md`.

Detailed tooling configuration belongs in the implementation and README.

## Deliverables

- A Git repository containing `frontend/` and `backend/`.
- A root `README.md` covering setup, how to run both layers, how to run tests and generate coverage, API-call examples, and design decisions or assumptions.
- Unit tests and coverage reports for both layers.
- This document at `docs/prd.md`.
- The supporting `calculator-behavior.md` and `DESIGN.md` specifications.
- A `docs/prompts.md` file recording the prompts used during development.
- **Optional:** Dockerfile(s) and supporting configuration to run the complete application.

## Acceptance criteria

- [ ] All seven operations work with positive, negative, zero, and decimal operands where valid, following `calculator-behavior.md`.
- [ ] Expressions respect grouping and precedence: `2 + 3 × 4` produces `14`, and `(2 + 3) × 4` produces `20`.
- [ ] Users can build, navigate, edit, and clear expressions through the keypad; typing, paste, and dropped text cannot modify them.
- [ ] Valid UI submissions produce results through the real backend API.
- [ ] Results use up to three decimal places without scientific notation, and continuing calculations use the displayed result.
- [ ] Invalid expressions, arithmetic errors, and request failures show understandable messages in the Result area without losing the expression or crashing either layer.
- [ ] Loading prevents duplicate submissions, and stale responses cannot overwrite newer input.
- [ ] The interface follows `DESIGN.md`, works on desktop and mobile, and supports keyboard button activation and accessible result/error feedback.
- [ ] Unit tests pass in both layers, and coverage reports can be reproduced.
- [ ] A reviewer can run the application from a clean checkout using the README.
- [ ] Development prompts are recorded in `docs/prompts.md`.
