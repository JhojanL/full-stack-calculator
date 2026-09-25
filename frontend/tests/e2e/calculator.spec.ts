import { expect, test, type Page } from '@playwright/test'

/** Activate names in order as exact keypad button labels on page. */
async function press(page: Page, ...names: string[]) {
  for (const name of names)
    await page.getByRole('button', { name, exact: true }).click()
}

/** Locate the labeled result region on page for text, focus, and overflow assertions. */
const result = (page: Page) =>
  page.getByRole('region', { name: 'Result', exact: true })
/** Locate the expression region on page by its changing accessible description. */
const expression = (page: Page) =>
  page.getByRole('region', { name: /^Expression:/ })

test.beforeEach(async ({ page }) => {
  await page.goto('/')
})

test('real API: precedence, error correction, rounded continuation, and fresh input', async ({
  page,
}) => {
  await expect(page.getByRole('button')).toHaveCount(25)
  await expect(result(page)).toHaveText('0')
  await press(page, '2', 'Add', '3', 'Multiply', '4', 'Calculate')
  await expect(result(page)).toHaveText('14')
  await expect(expression(page)).toContainText('2+3×4')
  await press(page, 'Calculate')
  await expect(result(page)).toHaveText('14')
  await press(page, 'Clear all', '1', 'Divide', '0', 'Calculate')
  await expect(result(page)).toHaveText('Cannot divide by zero.')
  await expect(
    page.getByRole('button', { name: 'Calculate', exact: true }),
  ).toBeFocused()
  await press(page, 'Backspace', '3', 'Calculate')
  await expect(result(page)).toHaveText('0.333')
  await press(page, 'Multiply', '3', 'Calculate')
  await expect(result(page)).toHaveText('0.999')
  await press(page, '7')
  await expect(expression(page)).toHaveAttribute(
    'aria-label',
    'Expression: 7. Cursor at position 1 of 1.',
  )
  await expect(result(page)).toHaveText('')
})

test('real API: grouped root, powers, percent, and negative result reuse', async ({
  page,
}) => {
  await press(
    page,
    'Square root',
    'Open parenthesis',
    '9',
    'Add',
    '5',
    'Close parenthesis',
    'Calculate',
  )
  await expect(result(page)).toHaveText('3.742')
  await press(
    page,
    '2',
    'Raise to a power',
    '3',
    'Raise to a power',
    '2',
    'Calculate',
  )
  await expect(result(page)).toHaveText('512')
  await press(
    page,
    '2',
    '0',
    '0',
    'Multiply',
    '1',
    '0',
    'Percentage',
    'Calculate',
  )
  await expect(result(page)).toHaveText('20')
  await press(page, '5', 'Subtract or enter a negative value', '8', 'Calculate')
  await expect(result(page)).toHaveText('-3')
  await press(page, 'Raise to a power', '2', 'Calculate')
  await expect(result(page)).toHaveText('9')
})

test('edits in the middle, rejects direct input, and validates before requesting', async ({
  page,
}) => {
  let requests = 0
  page.on('request', (request) => {
    if (request.url().endsWith('/calculate')) requests++
  })
  await press(page, 'Calculate')
  await expect(result(page)).toHaveText('Enter a calculation.')
  await press(page, 'Open parenthesis', '1', 'Add', 'Calculate')
  await expect(result(page)).toHaveText('Check the parentheses.')
  expect(requests).toBe(0)
  await press(
    page,
    'Clear all',
    '1',
    '2',
    '3',
    'Move cursor left',
    'Backspace',
    '4',
  )
  await expect(expression(page)).toHaveAttribute(
    'aria-label',
    'Expression: 143. Cursor at position 2 of 3.',
  )
  await page.keyboard.type('999+1')
  await expression(page).dispatchEvent('paste')
  await expression(page).dispatchEvent('drop')
  await expect(expression(page)).toContainText('143')
  await press(page, 'Decimal point', 'Decimal point')
  await expect(expression(page)).toContainText('14.3')
  await press(page, 'Calculate')
  await expect(result(page)).toHaveText('14.3')
})

test('native keyboard navigation preserves focus and activates with Enter and Space', async ({
  page,
}) => {
  await page.keyboard.press('Tab')
  await expect(result(page)).toBeFocused()
  const buttons = page.getByRole('button')
  for (let i = 0; i < 25; i++) {
    await page.keyboard.press('Tab')
    await expect(buttons.nth(i)).toBeFocused()
  }
  await page.keyboard.press('Shift+Tab')
  await expect(
    page.getByRole('button', { name: 'Add', exact: true }),
  ).toBeFocused()
  // Tab back from Add to 0 and enter a complete calculation using only buttons.
  await page.keyboard.press('Shift+Tab')
  await page.keyboard.press('Shift+Tab')
  await page.keyboard.press('Space')
  await expect(expression(page)).toContainText('0')
  await page.keyboard.press('Tab')
  await page.keyboard.press('Tab')
  await page.keyboard.press('Enter')
  await page.keyboard.press('Shift+Tab')
  await page.keyboard.press('Shift+Tab')
  await page.keyboard.press('Enter')
  await page.keyboard.press('Tab')
  await page.keyboard.press('Tab')
  await page.keyboard.press('Tab')
  await page.keyboard.press('Enter')
  await expect(result(page)).toHaveText('0')
  await expect(
    page.getByRole('button', { name: 'Calculate', exact: true }),
  ).toBeFocused()
})

test('shows loading, suppresses duplicate requests, and allows reset during a delayed response', async ({
  page,
}) => {
  let release!: () => void
  const gate = new Promise<void>((resolve) => {
    release = resolve
  })
  let requests = 0
  await page.route('**/calculate', async (route) => {
    requests++
    await gate
    await route.fulfill({ json: { result: '5' } })
  })
  await press(page, '2', 'Add', '3', 'Calculate')
  await expect(page.getByRole('status')).toHaveText('Calculating…')
  const equals = page.getByRole('button', { name: 'Calculate', exact: true })
  await expect(equals).toHaveAttribute('aria-disabled', 'true')
  await equals.press('Enter')
  expect(requests).toBe(1)
  await press(page, 'Clear all', '9')
  release()
  await expect(page.getByRole('status')).toBeEmpty()
  await expect(expression(page)).toContainText('9')
  await expect(result(page)).toHaveText('')
})

test('recovers from a network failure without losing the expression', async ({
  page,
}) => {
  await page.route('**/calculate', (route) => route.abort(), { times: 1 })
  await press(page, '4', 'Calculate')
  await expect(result(page)).toHaveText('Could not calculate. Try again.')
  await expect(expression(page)).toContainText('4')
  await press(page, 'Calculate')
  await expect(result(page)).toHaveText('4')
})

test('keeps all controls usable at 320px and 200% text with long values', async ({
  page,
}) => {
  await page.setViewportSize({ width: 320, height: 640 })
  await page.addStyleTag({ content: 'html { font-size: 200%; }' })
  for (const button of await page.getByRole('button').all()) {
    const box = await button.boundingBox()
    expect(box!.width).toBeGreaterThanOrEqual(44)
    expect(box!.height).toBeGreaterThanOrEqual(44)
  }
  expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBe(
    320,
  )
  await press(page, ...Array<string>(20).fill('9'))
  const caretVisible = await expression(page).evaluate((element) => {
    const caret = element
      .querySelector('.insertion-caret')!
      .getBoundingClientRect()
    const region = element.getBoundingClientRect()
    return (
      element.scrollWidth > element.clientWidth &&
      caret.left >= region.left &&
      caret.right <= region.right
    )
  })
  expect(caretVisible).toBe(true)
  await press(page, 'Calculate')
  await expect(result(page)).toHaveText('100000000000000000000')
  expect(
    await result(page).evaluate(
      (element) => element.scrollWidth > element.clientWidth,
    ),
  ).toBe(true)
  expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBe(
    320,
  )
})
