import { Display } from './components/Display'
import { Keypad } from './components/Keypad'
import { useCalculator } from './hooks/useCalculator'

export default function App() {
  const { state, dispatch } = useCalculator()
  const pending = state.feedback.kind === 'pending'
  return (
    <main className="page">
      <h1 className="sr-only">Calculator</h1>
      <section className="calculator" aria-label="Calculator">
        <Display state={state} />
        <div className="calculation-status" role="status">
          {pending ? 'Calculating…' : ''}
        </div>
        <Keypad onAction={dispatch} pending={pending} />
      </section>
    </main>
  )
}
