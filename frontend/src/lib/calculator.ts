import type { CalculatorAction, CalculatorState } from '../types/calculator'
import { validateExpression } from './validation'

/** Empty editor state; reducer callers must treat this shared value as immutable. */
export const initialState: CalculatorState = {
  expression: '',
  cursor: 0,
  feedback: { kind: 'editing' },
  continueResult: false,
  requestId: 0,
}

/**
 * Apply action to state without mutation or arithmetic evaluation.
 * Insert actions carry one keypad digit or symbol; request completions carry the
 * pending ID. Ignored actions may return state unchanged. A valid calculate
 * action records a pending expression for the hook to submit.
 */
export function calculatorReducer(
  state: CalculatorState,
  action: CalculatorAction,
): CalculatorState {
  // Preserve the sequence so a response from before reset cannot match a new request.
  if (action.type === 'clear')
    return { ...initialState, requestId: state.requestId }
  // Cancellation can race with completion; accept only the currently pending ID.
  if (action.type === 'success' || action.type === 'failure') {
    if (state.feedback.kind !== 'pending' || state.feedback.id !== action.id)
      return state
    return {
      ...state,
      continueResult: action.type === 'success',
      feedback:
        action.type === 'success'
          ? { kind: 'result', value: action.result }
          : { kind: 'error', message: action.message },
    }
  }
  if (action.type === 'calculate') {
    if (
      state.feedback.kind === 'pending' ||
      (state.feedback.kind === 'result' && state.continueResult)
    )
      return state
    const error = validateExpression(state.expression)
    if (error) return { ...state, feedback: { kind: 'error', message: error } }
    const id = state.requestId + 1
    return {
      ...state,
      requestId: id,
      feedback: { kind: 'pending', id, expression: state.expression },
    }
  }
  if (action.type === 'move')
    return {
      ...state,
      cursor: Math.max(
        0,
        Math.min(state.expression.length, state.cursor + action.direction),
      ),
      continueResult: false,
    }
  let { expression, cursor } = state
  if (action.type === 'backspace') {
    if (!cursor) return { ...state, continueResult: false }
    expression = expression.slice(0, cursor - 1) + expression.slice(cursor)
    cursor--
  } else {
    if (state.continueResult && state.feedback.kind === 'result') {
      if (/^[+−\-×÷^%]$/.test(action.value)) {
        const result = state.feedback.value
        expression = result.startsWith('-') ? `(${result})` : result
      } else if (/^[\d.(√]$/.test(action.value)) expression = ''
      cursor = expression.length
    }
    let value = action.value
    if (value === '.') {
      const left = expression.slice(0, cursor).match(/[\d.]*$/)?.[0] ?? ''
      const right = expression.slice(cursor).match(/^[\d.]*/)?.[0] ?? ''
      if ((left + right).includes('.')) return state
      if (!left) value = '0.'
    }
    expression = expression.slice(0, cursor) + value + expression.slice(cursor)
    cursor += value.length
  }
  return {
    ...state,
    expression,
    cursor,
    continueResult: false,
    feedback: { kind: 'editing' },
  }
}
