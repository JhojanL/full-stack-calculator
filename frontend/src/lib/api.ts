import { z } from 'zod'

/** User-facing fallback for transport failures and responses outside the API contract. */
export const requestFailure = 'Could not calculate. Try again.'
const messages = {
  INVALID_REQUEST:
    'Request must be a JSON object containing only a string expression.',
  UNSUPPORTED_MEDIA_TYPE: 'Content-Type must be application/json.',
  EMPTY_EXPRESSION: 'Enter a calculation.',
  INVALID_EXPRESSION: 'Check the expression.',
  INVALID_OPERAND: 'Check the expression.',
  UNMATCHED_PARENTHESES: 'Check the parentheses.',
  UNSUPPORTED_OPERATION: 'Check the expression.',
  DIVISION_BY_ZERO: 'Cannot divide by zero.',
  NEGATIVE_SQUARE_ROOT: 'Square root requires a nonnegative value.',
  UNSUPPORTED_POWER: 'This power cannot be calculated.',
  NUMERIC_OUT_OF_RANGE: 'The result is outside the supported range.',
  RATE_LIMIT_EXCEEDED: 'Rate limit exceeded. Try again shortly.',
  INTERNAL_ERROR: requestFailure,
} as const
const successSchema = z.strictObject({
  result: z
    .string()
    .regex(/^(?:0|-?[1-9][0-9]*|-?(?:0|[1-9][0-9]*)\.[0-9]{0,2}[1-9])$/),
})
const errorSchema = z.strictObject({
  error: z.strictObject({
    code: z.enum(
      Object.keys(messages) as [
        keyof typeof messages,
        ...(keyof typeof messages)[],
      ],
    ),
    message: z.string().min(1),
  }),
})
const statuses: Record<string, number[]> = {
  INVALID_REQUEST: [400, 413],
  UNSUPPORTED_MEDIA_TYPE: [415],
  RATE_LIMIT_EXCEEDED: [429],
  INTERNAL_ERROR: [500],
}

/**
 * Submit expression verbatim to the same-origin API and return its decimal string.
 * signal cancels the request; an independent 10-second deadline also aborts it.
 * Reject with a contractual error message, or requestFailure for invalid responses,
 * network failures, timeout, and cancellation. Always release timers/listeners.
 */
export async function calculate(
  expression: string,
  signal: AbortSignal,
): Promise<string> {
  const controller = new AbortController()
  const cancel = () => controller.abort()
  signal.addEventListener('abort', cancel, { once: true })
  if (signal.aborted) cancel()
  const timeout = setTimeout(cancel, 10_000)
  try {
    const response = await fetch('/calculate', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ expression }),
      signal: controller.signal,
    })
    const body: unknown = await response.json()
    if (response.status === 200) {
      const parsed = successSchema.safeParse(body)
      if (parsed.success) return parsed.data.result
    } else {
      const parsed = errorSchema.safeParse(body)
      if (parsed.success) {
        const { code, message } = parsed.data.error
        if (
          (statuses[code] ?? [422]).includes(response.status) &&
          message === messages[code]
        )
          throw new Error(message)
      }
    }
    throw new Error(requestFailure)
  } catch (error) {
    if (
      error instanceof Error &&
      Object.values(messages).some((message) => message === error.message)
    )
      throw error
    throw new Error(requestFailure, { cause: error })
  } finally {
    clearTimeout(timeout)
    signal.removeEventListener('abort', cancel)
  }
}
