---
version: alpha
name: Graphite Mint Calculator
description: A compact dark calculator with tactile keys, mint arithmetic controls, and warm-gray editing actions.
colors:
  primary: '#85D4AE'
  primary-hover: '#99DFC0'
  primary-pressed: '#72C49B'
  on-primary: '#17221C'
  secondary: '#A8A49A'
  secondary-hover: '#BAB6AC'
  secondary-pressed: '#949086'
  on-secondary: '#1C1E1B'
  background: '#111210'
  surface: '#252624'
  display: '#10110F'
  key: '#333430'
  key-hover: '#3F403A'
  key-pressed: '#2B2D28'
  text: '#F5F5F2'
  muted: '#B9B9B3'
  error: '#FFB4AB'
typography:
  result:
    fontFamily: Manrope
    fontSize: 3.5rem
    fontWeight: 500
    lineHeight: 1.1
    letterSpacing: -0.02em
    fontFeature: '"tnum" 1, "lnum" 1'
  expression:
    fontFamily: Manrope
    fontSize: 1.5rem
    fontWeight: 400
    lineHeight: 1.4
    fontFeature: '"tnum" 1, "lnum" 1'
  key:
    fontFamily: Manrope
    fontSize: 1.5rem
    fontWeight: 500
    lineHeight: 1.2
  label:
    fontFamily: Manrope
    fontSize: 0.75rem
    fontWeight: 400
    lineHeight: 1.5
  status:
    fontFamily: Manrope
    fontSize: 0.875rem
    fontWeight: 400
    lineHeight: 1.5
  error:
    fontFamily: Manrope
    fontSize: 1rem
    fontWeight: 400
    lineHeight: 1.5
rounded:
  key: 24px
  display: 20px
  shell: 36px
spacing:
  xs: 4px
  sm: 8px
  md: 16px
  lg: 20px
  xl: 24px
components:
  icon:
    size: 24px
  page:
    backgroundColor: '{colors.background}'
    textColor: '{colors.text}'
    padding: '{spacing.md}'
  calculator:
    backgroundColor: '{colors.surface}'
    textColor: '{colors.text}'
    rounded: '{rounded.shell}'
    padding: '{spacing.lg}'
  display:
    backgroundColor: '{colors.display}'
    textColor: '{colors.text}'
    rounded: '{rounded.display}'
    padding: '{spacing.lg}'
  display-label:
    backgroundColor: '{colors.display}'
    textColor: '{colors.muted}'
    typography: '{typography.label}'
  button-key:
    backgroundColor: '{colors.key}'
    textColor: '{colors.text}'
    typography: '{typography.key}'
    rounded: '{rounded.key}'
    height: 56px
  button-key-hover:
    backgroundColor: '{colors.key-hover}'
    textColor: '{colors.text}'
  button-key-pressed:
    backgroundColor: '{colors.key-pressed}'
    textColor: '{colors.text}'
  button-primary:
    backgroundColor: '{colors.primary}'
    textColor: '{colors.on-primary}'
    typography: '{typography.key}'
    rounded: '{rounded.key}'
    height: 56px
  button-primary-hover:
    backgroundColor: '{colors.primary-hover}'
    textColor: '{colors.on-primary}'
  button-primary-pressed:
    backgroundColor: '{colors.primary-pressed}'
    textColor: '{colors.on-primary}'
  button-secondary:
    backgroundColor: '{colors.secondary}'
    textColor: '{colors.on-secondary}'
    rounded: '{rounded.key}'
    height: 48px
  button-secondary-hover:
    backgroundColor: '{colors.secondary-hover}'
    textColor: '{colors.on-secondary}'
  button-secondary-pressed:
    backgroundColor: '{colors.secondary-pressed}'
    textColor: '{colors.on-secondary}'
  status:
    backgroundColor: '{colors.surface}'
    textColor: '{colors.muted}'
    typography: '{typography.status}'
  error:
    backgroundColor: '{colors.display}'
    textColor: '{colors.error}'
    typography: '{typography.error}'
---

# Graphite Mint Calculator

## Overview

A focused expression calculator for desktop and mobile. Preserve the approved dark, tactile design: an inset display, softly raised keys, mint arithmetic buttons, and distinct warm-gray delete/reset buttons. The result is the strongest typographic element; the full-width equals button is the primary action.

This file follows the [Google Labs DESIGN.md format](https://github.com/google-labs-code/design.md/blob/main/docs/spec.md). Tokens define exact implementation values; the generated image guides appearance. Use [calculator-behavior.md](../../docs/calculator-behavior.md) for evaluation and editing rules. The visual references are Linear's dark surfaces and tonik's tactile calculator buttons.

## Colors

| Role         | Token                              | Application                                                                                |
| ------------ | ---------------------------------- | ------------------------------------------------------------------------------------------ |
| Mint accent  | `primary`                          | Only `÷`, `×`, `−`, `+`, and `=` button faces; also the insertion caret and focus outline. |
| Warm neutral | `secondary`                        | Backspace and clear/reset button faces.                                                    |
| Graphite     | `key`                              | Digits, decimal point, navigation arrows, parentheses, square root, power, and percentage. |
| Dark layers  | `background`, `surface`, `display` | Page, calculator shell, and inset display.                                                 |
| Foreground   | `text`, `muted`                    | Numbers/symbols and supporting labels respectively.                                        |
| Error        | `error`                            | Soft-red error text replacing the number in the Result area.                               |

Keep mint symbols dark and graphite-key symbols off-white. Numeric results stay off-white. Mint indicates actions, not a success status. Warm gray makes correction controls easy to locate without giving them the same emphasis as calculation controls.

Base contrast ratios: dark symbols on mint **9.38:1**; dark symbols on warm gray **6.75:1**; off-white on graphite keys **11.48:1**. All defined hover/pressed foreground pairs also exceed 4.5:1. Recheck contrast after applying gradients and effects.

## Typography

Use **[Manrope from Fontsource](https://fontsource.org/fonts/manrope)**, with `system-ui, sans-serif` as fallbacks. Its geometric forms add character while keeping the interface close to the approved design. The implementation self-hosts the variable font with `@fontsource-variable/manrope`, imported once in `src/styles/global.css`, and uses the family `Manrope Variable`. Use tabular numerals for expressions and results. Keep labels sentence case: “Expression” and “Result.” Values align right; labels align left.

Result: 56px, expression/key symbols: 24px, labels: 12px, status: 14px, error messages: 16px at the default 16px root size. Use rem sizing so text scales with user preferences. Keep the result noticeably larger than the expression; allow the display to grow instead of clipping enlarged text.

## Layout

One calculator card, maximum width **400px**, with **16px page padding**. Center it when the viewport has enough space; otherwise allow natural vertical scrolling. Use 20px card padding, reduced to 16px below 480px viewport width. Keep four equal columns and 8px gaps at every breakpoint.

Order: display, reserved status area, editing row, function row, number grid, equals. The status area is normally empty, has a 24px minimum height, and grows for wrapped messages; it has no visible placeholder or dashed border.

| Row       | Column 1 | Column 2 | Column 3       | Column 4   |
| --------- | -------- | -------- | -------------- | ---------- |
| Editing   | `←`      | `→`      | Backspace icon | Reset icon |
| Functions | `(`      | `)`      | `√`            | `xʸ`       |
| Numbers   | `7`      | `8`      | `9`            | `÷`        |
| Numbers   | `4`      | `5`      | `6`            | `×`        |
| Numbers   | `1`      | `2`      | `3`            | `−`        |
| Numbers   | `%`      | `0`      | `.`            | `+`        |

Place one **full-width `=` button** below the grid. Keep all **25 controls** visible and in this order on desktop and mobile. Editing keys have a 48px minimum height, other keys 56px, and equals 60px. Height tokens are minimums, not clipping limits. Every target is at least 44 × 44 CSS pixels.

At 320px viewport width, retain all four columns. At 200% text size, allow keys, display, and status area to grow. Keep long expressions/results readable through internal horizontal scrolling; scroll the expression to follow the caret. Do not truncate values or change them to scientific notation.

## Elevation & Depth

Use CSS effects to reproduce the soft physical depth:

- Shell: `0 16px 40px rgba(0,0,0,0.35)` and a subtle 1px light inner edge.
- Keys: `0 4px 7px rgba(0,0,0,0.45)`, with a light top inset edge and dark bottom inset edge.
- Display: dark inset surface with `inset 0 2px 5px rgba(0,0,0,0.35)`.
- Key faces: restrained vertical shading over their base color, no more than 4% white at the top or 4% black at the bottom.

On press, lower the key by 1px and reduce its shadow. Use 120ms transitions for color, shadow, and transform. Respect reduced-motion preferences. Avoid colored glow, glass effects, and large animated movement.

## Shapes

Use rounded rectangles throughout: 36px shell corners, 20px display corners, and 24px key corners. Equals uses the same corner treatment as the keys. Maintain consistent radii and spacing across every button family.

## Components

### Display and status

Keep separate labeled expression and result areas, with 16px between their groups. The initial/reset state has an empty expression and result `0`. Show a mint caret while editing; button input is the only way to change the expression.

| State       | Presentation                                                                                                                                                                               |
| ----------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Editing     | Show the expression and caret. Clear the previous result on the first actual edit.                                                                                                         |
| Calculating | Show “Calculating…” in the status area; keep `=` visible and suppress duplicate activation. Editing and reset remain available.                                                            |
| Result      | Keep the submitted expression and show the result with up to three decimal places, without unnecessary trailing zeros or `-0`.                                                             |
| Error       | Preserve the expression; replace the number in the Result area with a specific message such as “Cannot divide by zero.” Use the existing soft-red `error` color and 16px error typography. |

Keep the “Result” label and the result area's normal minimum height. Error text aligns right and wraps without truncation; allow the area to grow if needed. The separate status area continues to show loading only. Clear the error on the first actual edit or reset; successful calculation restores the normal numeric result. Announce the error once without moving focus, so feedback does not depend on red text alone.

Use `2 + 3 × 4` and result `14` for the reference success state. An arrow moves the caret without changing the calculation. Backspace deletes before it; reset clears everything. The `√` key inserts only `√`: `√9` is valid, while grouped expressions use `√(9 + 5)`.

### Buttons and icons

Use native buttons with digit, symbol, or icon faces. Keep visual word labels off the keypad. Use **[Lucide for React](https://lucide.dev/guide/react)** (`lucide-react`) for arithmetic and editing icons, importing only the required components. This gives these controls consistent proportions and strokes. Keep digits, the decimal point, and individual parentheses as Manrope text.

| Button face    | Rendering                 | Accessible button name             |
| -------------- | ------------------------- | ---------------------------------- |
| `+`            | Lucide `Plus`             | Add                                |
| `−`            | Lucide `Minus`            | Subtract or enter a negative value |
| `×`            | Lucide `X`                | Multiply                           |
| `÷`            | Lucide `Divide`           | Divide                             |
| `=`            | Lucide `Equal`            | Calculate                          |
| `%`            | Lucide `Percent`          | Percentage                         |
| `√`            | Lucide `Radical`          | Square root                        |
| `←`            | Lucide `ArrowLeft`        | Move cursor left                   |
| `→`            | Lucide `ArrowRight`       | Move cursor right                  |
| Backspace icon | Lucide `Delete`           | Backspace                          |
| Reset icon     | Lucide `RotateCcw`        | Clear all                          |
| `xʸ`           | Custom `ExponentIcon` SVG | Raise to a power                   |
| `0–9`          | Manrope text              | The visible digit                  |
| `.`            | Manrope text              | Decimal point                      |
| `(`            | Manrope text              | Open parenthesis                   |
| `)`            | Manrope text              | Close parenthesis                  |

Use `components.icon.size` (**24px**) with a **2px stroke**, round caps and joins, `fill="none"`, and `stroke="currentColor"`. Center each icon in its existing button; inherit its foreground color from the button's color tokens. The icon size is independent of the button's touch target.

Implement `ExponentIcon` as one local React SVG component with `viewBox="0 0 24 24"`. Draw `x` and a smaller raised `y` using paths, matching Lucide's size and stroke treatment. The raised `y` communicates a user-entered exponent; keep this distinct from the fixed square shown by Lucide's `Superscript` icon.

Give each button the accessible name above using `aria-label` or visually hidden text. Hide its decorative SVG from screen readers with `aria-hidden="true"`; the button remains the focusable, named control. All controls keep their existing positions, colors, sizes, and behavior.

Use the corresponding hover/pressed color tokens without changing layout. Focus uses a **2px mint outline with 3px offset**, visible against the surrounding graphite even on mint keys. Preserve focus after activation. While calculating, retain `=` contrast and focus; expose its unavailable state accessibly and guard repeat activation.

### Accessibility

Target [WCAG 2.2 AA](https://www.w3.org/TR/WCAG22/). Implementation must verify:

- Accessible button names, including “Move cursor left,” “Move cursor right,” “Backspace,” “Clear all,” “Square root,” “Raise to a power,” and “Calculate.” Hide decorative icons from screen readers.
- Tab/Shift+Tab navigation and Enter/Space activation for every action. No typing, paste, or dropped-text input.
- Accessible expression and insertion position; announcements for cursor movement, results, loading, and errors without moving focus.
- Text contrast of at least 4.5:1 and essential icons/focus indicators of at least 3:1. Explain errors with text as well as color.
- Usable zoom, reflow, and touch targets. The 44px target size is a project requirement above the AA minimum.

## Do's and Don'ts

- Preserve the approved keypad order, warm-gray correction controls, and five mint action buttons.
- Use the exact tokens when the generated image's lighting makes a color appear different.
- Build real text, icons, and buttons; use the mockup as a visual reference.
- Keep evaluation, rounding, and result-continuation behavior aligned with `calculator-behavior.md`.
- Do not add history, memory keys, theme switching, extra scientific functions, or decorative dashboards.
- Do not use yellow accents, neon green, color-only errors, hidden operations, or tiny controls to fit the screen.
