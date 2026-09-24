import { describe, expect, it } from 'vitest'
import {
  calculatorReducer as reduce,
  initialState,
} from '../../src/lib/calculator'
import { validateExpression } from '../../src/lib/validation'
import type { CalculatorState } from '../../src/types/calculator'

function enter(expression: string, state = initialState): CalculatorState {
  return [...expression].reduce(
    (current, value) => reduce(current, { type: 'insert', value }),
    state,
  )
}
function success(expression: string, result: string) {
  const pending = reduce(enter(expression), { type: 'calculate' })
  return reduce(pending, { type: 'success', id: pending.requestId, result })
}

describe('keypad editing', () => {
  it('inserts and removes at the cursor, with bounded movement', () => {
    let state = reduce(enter('123'), { type: 'move', direction: -1 })
    state = reduce(state, { type: 'backspace' })
    expect(state.expression).toBe('13')
    expect(enter('4', state).expression).toBe('143')
    expect(reduce(initialState, { type: 'move', direction: -1 }).cursor).toBe(0)
    expect(reduce(enter('1'), { type: 'move', direction: 1 }).cursor).toBe(1)
  })
  it('starts decimals with zero and rejects duplicate dots on either side', () => {
    expect(enter('.5+.2').expression).toBe('0.5+0.2')
    expect(enter('1.2.3').expression).toBe('1.23')
    const state = { ...enter('12.3'), cursor: 1 }
    expect(enter('.', state)).toBe(state)
    expect(enter('.', { ...enter('12'), cursor: 0 }).expression).toBe('0.12')
  })
  it.each(['4', '.', '(', '√'])(
    'starts fresh after a result with %s',
    (value) => {
      expect(enter(value, success('2+3', '5')).expression).toBe(
        value === '.' ? '0.' : value,
      )
    },
  )
  it('continues from the rounded string and groups negative results', () => {
    expect(enter('×3', success('1÷3', '0.333')).expression).toBe('0.333×3')
    expect(enter('^2', success('5−8', '-3')).expression).toBe('(-3)^2')
    expect(enter('%', success('2+3', '5')).expression).toBe('5%')
  })
  it('keeps the original expression and result on arrows until an actual edit', () => {
    const result = success('2+3', '5')
    const moved = reduce(result, { type: 'move', direction: -1 })
    expect(moved.feedback).toEqual({ kind: 'result', value: '5' })
    expect(enter('4', moved).expression).toBe('2+43')
    expect(enter('4', moved).feedback.kind).toBe('editing')
    expect(reduce(result, { type: 'backspace' }).expression).toBe('2+')
    expect(enter(')', result).expression).toBe('2+3)')
  })
  it('suppresses repeated equals after success and during loading', () => {
    const result = success('2+3', '5')
    expect(reduce(result, { type: 'calculate' })).toBe(result)
    const pending = reduce(enter('2+3'), { type: 'calculate' })
    expect(reduce(pending, { type: 'calculate' })).toBe(pending)
  })
  it.each(['success', 'failure'] as const)(
    'ignores stale %s after editing, reset, and a new request',
    (type) => {
      const pending = reduce(enter('2+3'), { type: 'calculate' })
      const stale =
        type === 'success'
          ? { type, id: pending.requestId, result: '5' }
          : { type, id: pending.requestId, message: 'Old error' }
      const edited = enter('4', pending)
      expect(reduce(edited, stale)).toBe(edited)
      const cleared = reduce(pending, { type: 'clear' })
      expect(reduce(cleared, stale)).toBe(cleared)
      const newer = reduce(enter('9', cleared), { type: 'calculate' })
      expect(reduce(newer, stale)).toBe(newer)
    },
  )
  it('preserves errors on cursor movement and permits retry', () => {
    const pending = reduce(enter('2'), { type: 'calculate' })
    const failed = reduce(pending, {
      type: 'failure',
      id: pending.requestId,
      message: 'Could not calculate. Try again.',
    })
    expect(reduce(failed, { type: 'move', direction: -1 }).feedback).toEqual(
      failed.feedback,
    )
    expect(reduce(failed, { type: 'calculate' }).feedback.kind).toBe('pending')
    expect(enter('3', failed).feedback.kind).toBe('editing')
  })
})

describe('submission syntax', () => {
  it.each([
    '2+3×4',
    '(2+3)×4',
    '2^3^2',
    '−2÷3',
    '√(9+5)',
    '√-9',
    '200×10%',
    '2^-3',
    '01.',
    '0.',
    '2--3',
    '(2)%',
    '√9%',
    '1/0',
    '0^0',
  ])('accepts %s for backend evaluation', (expression) => {
    expect(validateExpression(expression)).toBeNull()
  })
  it.each([
    '2+',
    '2(3)',
    '10%%',
    '--2',
    '√√9',
    '√-(2)',
    '.5',
    '1..2',
    '1e3',
    '2 3',
    '2**3',
  ])('rejects invalid syntax %s', (expression) => {
    expect(validateExpression(expression)).toBe('Check the expression.')
  })
  it('reports empty input and parentheses specifically', () => {
    expect(validateExpression('')).toBe('Enter a calculation.')
    expect(validateExpression('(2+')).toBe('Check the parentheses.')
    expect(validateExpression(')2(')).toBe('Check the parentheses.')
  })
  it('bounds recursive nesting but accepts long flat expressions', () => {
    expect(
      validateExpression('('.repeat(128) + '1' + ')'.repeat(128)),
    ).toBeNull()
    expect(validateExpression('('.repeat(129) + '1' + ')'.repeat(129))).toBe(
      'Check the expression.',
    )
    expect(validateExpression('1^'.repeat(129) + '1')).toBe(
      'Check the expression.',
    )
    expect(validateExpression('1+'.repeat(2000) + '1')).toBeNull()
  })
})
