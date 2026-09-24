import { useLayoutEffect, useRef } from 'react'
import type { CalculatorState } from '../types/calculator'

/** Present state as read-only text and scroll the insertion point into view. */
export function Display({ state }: { state: CalculatorState }) {
  const caret = useRef<HTMLSpanElement>(null)
  const expressionRegion = useRef<HTMLDivElement>(null)
  const { expression, cursor, feedback, continueResult } = state
  const editing = !continueResult
  useLayoutEffect(() => {
    const region = expressionRegion.current
    const marker = caret.current
    if (!region || !marker) return
    const position = marker.getBoundingClientRect()
    const viewport = region.getBoundingClientRect()
    if (position.right > viewport.right)
      region.scrollLeft += position.right - viewport.right + 4
    if (position.left < viewport.left)
      region.scrollLeft -= viewport.left - position.left + 4
  }, [expression, cursor, editing])
  const cursorDescription = `Expression: ${expression || 'empty'}. Cursor at position ${cursor} of ${expression.length}.`
  return (
    <div className="calculator-display">
      <div id="expression-label" className="display-label">
        Expression
      </div>
      <div
        ref={expressionRegion}
        className="expression-scroll"
        role="region"
        aria-label={cursorDescription}
      >
        <div className="expression-text" aria-hidden="true">
          {Array.from({ length: expression.length + 1 }, (_, index) => (
            <span key={index}>
              {index === cursor && editing && (
                <span ref={caret} className="insertion-caret" />
              )}
              {index < expression.length && (
                <span
                  className={
                    /[+−\-×÷^]/.test(expression[index])
                      ? 'expression-operator'
                      : undefined
                  }
                >
                  {expression[index] === '-' ? '−' : expression[index]}
                </span>
              )}
            </span>
          ))}
        </div>
      </div>
      <div className="sr-only" aria-live="polite" aria-atomic="true">
        {cursorDescription}
      </div>
      <div id="result-label" className="display-label mt-4">
        Result
      </div>
      <div
        className="result-area"
        aria-labelledby="result-label"
        role="region"
        tabIndex={0}
      >
        {feedback.kind === 'error' ? (
          <p className="result-error">{feedback.message}</p>
        ) : (
          <p className="result-number">
            {feedback.kind === 'result'
              ? feedback.value
              : expression
                ? '\u00a0'
                : '0'}
          </p>
        )}
      </div>
      <div className="sr-only" aria-live="polite" aria-atomic="true">
        {feedback.kind === 'result'
          ? `Result: ${feedback.value}`
          : feedback.kind === 'error'
            ? feedback.message
            : ''}
      </div>
    </div>
  )
}
