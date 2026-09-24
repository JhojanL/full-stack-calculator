const invalid = 'Check the expression.'

/** Validate keypad expression syntax without evaluating any arithmetic. */
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
  // Only parentheses and right-hand powers consume the API's nesting budget.
  function nested(parse: () => void) {
    if (++depth > 128) throw new Error(invalid)
    parse()
    depth--
  }
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
