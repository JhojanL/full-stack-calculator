# Calculator behavior

A simple expression calculator supporting addition, subtraction, multiplication, division, exponentiation, square root, and percentage.

## 1. Input and editing

Build expressions using the on-screen buttons only. The display does not accept typing, pasting, or dropped text. Scientific notation is not supported in input or results.

Button faces use digits, mathematical symbols, or action icons. Give each button an accessible name; visible word labels are unnecessary.

| Button             | Behavior                                                                                                                    |
| ------------------ | --------------------------------------------------------------------------------------------------------------------------- |
| `0–9`              | Insert a digit at the current insertion point.                                                                              |
| `.`                | Insert a decimal point. Start a new decimal with `0.` and prevent a second decimal point in the same number.                |
| `+`, `−`, `×`, `÷` | Insert the operator. Use `−` before a value to enter a negative number.                                                     |
| `xʸ`               | Insert the power operator, displayed as `^`.                                                                                |
| `√`                | Insert `√`, which applies to the next number. Use parentheses for an expression: `√9` or `√(9 + 5)`.                        |
| `%`                | Apply percentage to the preceding value.                                                                                    |
| `(` and `)`        | Insert parentheses to control grouping.                                                                                     |
| `←` and `→`        | Move the insertion point one digit or symbol left or right, skipping decorative spacing. Stop at the expression boundaries. |
| Backspace icon     | Remove the digit or symbol immediately before the insertion point.                                                          |
| Clear icon         | Reset to an empty expression and result `0`; clear the insertion point and errors.                                          |
| `=`                | Validate and calculate the complete expression.                                                                             |

Keep the insertion point visible. Buttons insert at that position, allowing a user to navigate, delete, and replace part of an expression. Arrow movement does not alter the calculation. Do not automatically change operators or insert missing multiplication signs.

## 2. Calculation rules

- Respect parentheses. Percentage binds before powers and applies to the preceding number, parenthesized expression, or square-root value.
- Evaluate powers before multiplication/division, then addition/subtraction. Powers associate right to left; multiplication/division and addition/subtraction each evaluate left to right.
- A power binds before a leading minus: `-2^2` is `-4`, while `(-2)^2` is `4`. Negative exponents are supported.
- Percentage always divides by 100: `200 × 10%` is `20`; `200 + 10%` is `200.1`. To add 10% to 200, use `200 × (1 + 10%)`.
- Square root applies to the next number or parenthesized expression and returns the nonnegative root: `√9 + 5` is `8`; `√(9 + 5)` displays `3.742`. Negative bases in powers require whole-number exponents. Reject negative square roots, division by zero, `0^0`, and zero raised to a negative power.

| Expression    | Displayed result |
| ------------- | ---------------- |
| `2 + 3 × 4`   | `14`             |
| `(2 + 3) × 4` | `20`             |
| `2^3^2`       | `512`            |
| `5 − 8`       | `-3`             |
| `1 ÷ 3`       | `0.333`          |
| `−2 ÷ 3`      | `-0.667`         |
| `√2`          | `1.414`          |
| `200 × 10%`   | `20`             |

## 3. Results and the next action

Calculate only after `=` is activated. Keep the expression visible and show a positive number, a negative number, or zero in the result area.

- Round the final result to at most **3 decimal places**, using the float64 display-rounding rule in OpenAPI (nearest, scaled halfway values away from zero).
- Remove unnecessary trailing zeros: show `2`, `2.5`, or `2.125`.
- Use ordinary decimal notation for every result. A value that rounds to zero displays `0`, never `-0`.
- Use Go `float64` arithmetic throughout, with `math.Sqrt` and `math.Pow`. Only the final result is rounded to three decimal places for display. Binary representation errors and underflow to zero can occur; exact decimal arithmetic is not guaranteed. Domain checks use the evaluated `float64` values before display rounding. See OpenAPI for the precise display-rounding rule.

Immediately after a result:

- A digit, `.`, `(`, or `√` starts a new expression and clears the earlier result.
- An arithmetic operator or `%` starts an expression using the **displayed, rounded result**. Wrap negative results in parentheses when reusing them. For example, after `1 ÷ 3 → 0.333`, continuing with `× 3` produces `0.999`.
- An arrow or backspace returns to editing the original expression. Subsequent buttons edit at the insertion point. Clear the earlier result on the first actual edit. A closing parenthesis also edits the original expression.
- Repeated `=` leaves the result unchanged. Clear resets everything.

To start a fresh negative expression after a result, activate clear and then `−`.

## 4. Validation and loading

Allow incomplete expressions while the user builds them. Validate on `=` and retain the expression when an error occurs.

| Problem                              | Message                                      |
| ------------------------------------ | -------------------------------------------- |
| Empty expression                     | “Enter a calculation.”                       |
| Missing value or misplaced operator  | “Check the expression.”                      |
| Unmatched parentheses                | “Check the parentheses.”                     |
| Division by zero                     | “Cannot divide by zero.”                     |
| Negative square root                 | “Square root requires a nonnegative value.”  |
| Unsupported power                    | “This power cannot be calculated.”           |
| Numeric overflow or nonfinite result | “The result is outside the supported range.” |
| Rate-limited request                 | “Rate limit exceeded. Try again shortly.”    |
| Failed or timed-out request          | “Could not calculate. Try again.”            |

Show the error message in the **Result area**, replacing the number while keeping the expression visible for correction. Use smaller, soft-red text; wrap the message without truncation and announce it to screen readers without moving focus. Keep the “Result” label and clear the message on the first actual edit or reset. A successful calculation restores the normal numeric result. The backend validates and evaluates the expression even though input comes from buttons.

While waiting, show “Calculating…” and prevent duplicate submissions. Keep editing and clear available; an edit or reset invalidates the pending request so its response cannot overwrite newer work. Preserve the expression after a request failure and let `=` retry.

## 5. Accessibility and responsive behavior

Use **WCAG 2.2 AA** as the accessibility target, following the [W3C overview](https://www.w3.org/WAI/standards-guidelines/wcag/) and [standard](https://www.w3.org/TR/WCAG22/).

- Use native buttons with accessible names such as “Multiply,” “Move cursor left,” and “Clear all.” Hide decorative icon graphics from screen readers.
- Every action must work through Tab/Shift+Tab navigation and Enter/Space activation. Keep focus visible and on the activated button. Button-based keyboard access does not enable free-text entry or paste.
- Expose the expression and insertion position accessibly; announce cursor movements, completed results, and errors without moving focus.
- Use at least 4.5:1 contrast for text and 3:1 for essential icons and control indicators. Explain errors in text, not color alone.
- Use touch targets of at least 44 × 44 CSS pixels. Keep all operations available at 320 CSS pixels wide and with text enlarged to 200%; preserve browser zoom.
- Let long expressions and results scroll within the display while keeping the keypad usable.

Button semantics follow the [W3C button pattern](https://www.w3.org/WAI/ARIA/apg/patterns/button/).

## 6. Review checklist

- Verify the examples above, including negative results and three-decimal rounding.
- Insert and delete a digit in the middle using the arrow and backspace buttons.
- Confirm typing, pasting, and dropping text do not change the expression.
- Correct an invalid expression without clearing the entire calculation.
- Complete and edit a calculation using only keyboard activation of buttons.
- Check result/error announcements and mobile layout.

Keep the scope to this calculator. History, memory registers, extra scientific functions, and visual themes are not required.
