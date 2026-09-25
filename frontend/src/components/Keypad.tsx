import {
  ArrowLeft,
  ArrowRight,
  Delete,
  Divide,
  Equal,
  Minus,
  Percent,
  Plus,
  Radical,
  RotateCcw,
  X,
} from 'lucide-react'
import type { ReactNode } from 'react'
import type { CalculatorAction } from '../types/calculator'
import { ExponentIcon } from './ExponentIcon'

interface Key {
  label: string
  face: ReactNode
  action: CalculatorAction
  variant?: 'primary' | 'secondary'
}
const keys: Key[] = [
  {
    label: 'Move cursor left',
    face: <ArrowLeft />,
    action: { type: 'move', direction: -1 },
  },
  {
    label: 'Move cursor right',
    face: <ArrowRight />,
    action: { type: 'move', direction: 1 },
  },
  {
    label: 'Backspace',
    face: <Delete />,
    action: { type: 'backspace' },
    variant: 'secondary',
  },
  {
    label: 'Clear all',
    face: <RotateCcw />,
    action: { type: 'clear' },
    variant: 'secondary',
  },
  {
    label: 'Open parenthesis',
    face: '(',
    action: { type: 'insert', value: '(' },
  },
  {
    label: 'Close parenthesis',
    face: ')',
    action: { type: 'insert', value: ')' },
  },
  {
    label: 'Square root',
    face: <Radical />,
    action: { type: 'insert', value: '√' },
  },
  {
    label: 'Raise to a power',
    face: <ExponentIcon />,
    action: { type: 'insert', value: '^' },
  },
  ...['7', '8', '9'].map(digit),
  {
    label: 'Divide',
    face: <Divide />,
    action: { type: 'insert', value: '÷' },
    variant: 'primary',
  },
  ...['4', '5', '6'].map(digit),
  {
    label: 'Multiply',
    face: <X />,
    action: { type: 'insert', value: '×' },
    variant: 'primary',
  },
  ...['1', '2', '3'].map(digit),
  {
    label: 'Subtract or enter a negative value',
    face: <Minus />,
    action: { type: 'insert', value: '−' },
    variant: 'primary',
  },
  {
    label: 'Percentage',
    face: <Percent />,
    action: { type: 'insert', value: '%' },
  },
  digit('0'),
  { label: 'Decimal point', face: '.', action: { type: 'insert', value: '.' } },
  {
    label: 'Add',
    face: <Plus />,
    action: { type: 'insert', value: '+' },
    variant: 'primary',
  },
  {
    label: 'Calculate',
    face: <Equal />,
    action: { type: 'calculate' },
    variant: 'primary',
  },
]

/** Create a text key whose label and inserted value are the supplied single decimal digit. */
function digit(value: string): Key {
  return { label: value, face: value, action: { type: 'insert', value } }
}

/**
 * Render ordered keys; onAction receives the activated key's action.
 * pending marks equals as aria-disabled while keeping it focusable. The caller
 * must suppress duplicate calculations because button clicks still dispatch.
 */
export function Keypad({
  onAction,
  pending,
}: {
  onAction: (action: CalculatorAction) => void
  pending: boolean
}) {
  return (
    <div className="grid grid-cols-4 gap-2">
      {keys.map((key, index) => (
        <button
          key={key.label}
          type="button"
          aria-label={key.label}
          aria-disabled={
            key.action.type === 'calculate' && pending ? true : undefined
          }
          className={`calculator-key ${key.variant ? `key-${key.variant}` : ''} ${index < 4 ? 'key-edit' : ''} ${key.action.type === 'calculate' ? 'key-equals col-span-4' : ''}`}
          onClick={() => onAction(key.action)}
        >
          <span aria-hidden="true">{key.face}</span>
        </button>
      ))}
    </div>
  )
}
