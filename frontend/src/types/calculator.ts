export type EditAction =
  | { type: 'insert'; value: string }
  | { type: 'move'; direction: -1 | 1 }
  | { type: 'backspace' }
  | { type: 'clear' }

export type Feedback =
  | { kind: 'editing' }
  | { kind: 'result'; value: string }
  | { kind: 'error'; message: string }
  | { kind: 'pending'; id: number; expression: string }

export interface CalculatorState {
  expression: string
  cursor: number
  feedback: Feedback
  continueResult: boolean
  requestId: number
}

export type CalculatorAction =
  | EditAction
  | { type: 'calculate' }
  | { type: 'success'; id: number; result: string }
  | { type: 'failure'; id: number; message: string }
