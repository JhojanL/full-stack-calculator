const invalid = 'Check the expression.'

/**
 * Check expression syntax before submission, returning a display message or null.
 * expression may be incomplete keypad input. Arithmetic domains and numeric range
 * remain the backend's responsibility; syntactically valid input can still fail.
 */
export function validateExpression(expression: string): string | null {
  if (!expression.trim()) return 'Enter a calculation.'
  const tokens = expression.match(/\d+(?:\.\d*)?|[+−\-×*÷/^%()√]|\S/g) ?? []
  if (
    tokens.some((token) => !/^(?:\d+(?:\.\d*)?|[+−\-×*÷/^%()√])$/.test(token))
  )
    return invalid
  let balance = 0
  for (const token of tokens) {
    if (token === '(') balance++
    if (token === ')') balance--
    if (balance < 0) return 'Check the parentheses.'
  }
  if (balance !== 0) return 'Check the parentheses.'
  let pos = 0
  let depth = 0
  /** Run parse one level deeper; only groups and right-hand powers consume the 128-level budget. */
  function nested(parse: () => void) {
    if (++depth > 128) throw new Error(invalid)
    parse()
    depth--
  }
  /** Consume the next token only if it matches one of values; leave pos unchanged otherwise. */
  function take(values: string[]) {
    if (!values.includes(tokens[pos])) return false
    pos++
    return true
  }
  function number() {
    if (!/^\d+(?:\.\d*)?$/.test(tokens[pos] ?? '')) throw new Error(invalid)
    pos++
  }
  function primary() {
    if (take(['('])) {
      nested(sum)
      if (!take([')'])) throw new Error(invalid)
    } else if (take(['√'])) {
      if (tokens[pos] === '(') primary()
      else {
        take(['-', '−'])
        number()
      }
    } else number()
  }
  function unary() {
    take(['-', '−'])
    primary()
    take(['%'])
    if (take(['^'])) nested(unary)
  }
  function product() {
    unary()
    while (take(['*', '×', '/', '÷'])) unary()
  }
  function sum() {
    product()
    while (take(['+', '-', '−'])) product()
  }
  try {
    sum()
    return pos === tokens.length ? null : invalid
  } catch {
    return invalid
  }
}
