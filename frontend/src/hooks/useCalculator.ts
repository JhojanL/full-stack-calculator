import { useEffect, useReducer } from 'react'
import { calculatorReducer, initialState } from '../lib/calculator'
import { calculate, requestFailure } from '../lib/api'

/** Own keypad state and cancel superseded or unmounted requests. */
export function useCalculator() {
  const [state, dispatch] = useReducer(calculatorReducer, initialState)
  const pending = state.feedback.kind === 'pending' ? state.feedback : null
  useEffect(() => {
    if (!pending) return
    const controller = new AbortController()
    void calculate(pending.expression, controller.signal).then(
      (result) => dispatch({ type: 'success', id: pending.id, result }),
      (error: unknown) =>
        dispatch({
          type: 'failure',
          id: pending.id,
          message: error instanceof Error ? error.message : requestFailure,
        }),
    )
    return () => controller.abort()
  }, [pending])
  return { state, dispatch }
}
