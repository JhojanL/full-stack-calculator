/** Keypad edits; insert carries one digit or symbol and move advances one position. */
export type EditAction =
  | { type: 'insert'; value: string }
  | { type: 'move'; direction: -1 | 1 }
  | { type: 'backspace' }
  | { type: 'clear' }

/** Display feedback; pending retains the submitted expression and its request identity. */
export type Feedback =
  | { kind: 'editing' }
  | { kind: 'result'; value: string }
  | { kind: 'error'; message: string }
  | { kind: 'pending'; id: number; expression: string }

/** Editor and request state owned by the calculator reducer. */
export interface CalculatorState {
  expression: string
  /** UTF-16 insertion offset, from zero through expression.length. */
  cursor: number
  feedback: Feedback
  /** The next eligible key can reuse the displayed result until editing resumes. */
  continueResult: boolean
  /** Submission sequence retained across clear to reject stale completions. */
  requestId: number
}

/** Editing/submission intent or a completion tagged with the originating pending ID. */
export type CalculatorAction =
  | EditAction
  | { type: 'calculate' }
  | { type: 'success'; id: number; result: string }
  | { type: 'failure'; id: number; message: string }
