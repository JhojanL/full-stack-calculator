import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { calculate, requestFailure } from '../../src/lib/api'

beforeEach(() => {
  vi.stubEnv('PROD', false)
  vi.stubEnv('VITE_API_BASE_URL', undefined)
})

afterEach(() => {
  vi.unstubAllEnvs()
  vi.unstubAllGlobals()
  vi.useRealTimers()
})
/** Stub fetch with JSON-encoded body and HTTP status (200 by default); return the spy. */
function respond(body: unknown, status = 200) {
  const fetch = vi
    .fn<typeof globalThis.fetch>()
    .mockResolvedValue(new Response(JSON.stringify(body), { status }))
  vi.stubGlobal('fetch', fetch)
  return fetch
}

describe('API boundary', () => {
  it.each([
    [true, 'https://calculator-api.jhojanlerma.dev'],
    [true, 'https://calculator-api.jhojanlerma.dev/'],
    [true, undefined],
    [false, 'https://calculator-api.jhojanlerma.dev'],
  ])(
    'routes requests with production=%s and base=%s',
    async (production, base) => {
      vi.stubEnv('PROD', production)
      vi.stubEnv('VITE_API_BASE_URL', base)
      const fetch = respond({ result: '2' })
      await expect(
        calculate('1+1', new AbortController().signal),
      ).resolves.toBe('2')
      expect(fetch).toHaveBeenCalledWith(
        production && base
          ? 'https://calculator-api.jhojanlerma.dev/calculate'
          : '/calculate',
        expect.objectContaining({
          method: 'POST',
          body: '{"expression":"1+1"}',
        }),
      )
    },
  )

  it('posts JSON and preserves large decimal results exactly', async () => {
    const result = '123456789012345678901234567890'
    const fetch = respond({ result })
    await expect(
      calculate('2^100', new AbortController().signal),
    ).resolves.toBe(result)
    expect(fetch).toHaveBeenCalledWith(
      '/calculate',
      expect.objectContaining({
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: '{"expression":"2^100"}',
      }),
    )
  })
  it.each([
    { result: 14 },
    { result: '1e3' },
    { result: '-0' },
    { result: '1.230' },
    { result: '01' },
    { result: '1.2345' },
    { result: '14', extra: true },
    {},
  ])('rejects malformed success %j', async (body) => {
    respond(body)
    await expect(calculate('2', new AbortController().signal)).rejects.toThrow(
      requestFailure,
    )
  })
  it.each([
    [422, 'DIVISION_BY_ZERO', 'Cannot divide by zero.'],
    [429, 'RATE_LIMIT_EXCEEDED', 'Rate limit exceeded. Try again shortly.'],
    [500, 'INTERNAL_ERROR', requestFailure],
  ])('handles contract error status %i', async (status, code, message) => {
    respond({ error: { code, message } }, Number(status))
    await expect(
      calculate('1÷0', new AbortController().signal),
    ).rejects.toThrow(String(message))
  })
  it.each([
    [
      200,
      {
        error: { code: 'DIVISION_BY_ZERO', message: 'Cannot divide by zero.' },
      },
    ],
    [
      500,
      {
        error: { code: 'DIVISION_BY_ZERO', message: 'Cannot divide by zero.' },
      },
    ],
    [422, { error: { code: 'UNKNOWN', message: 'Bad' } }],
    [
      422,
      { error: { code: 'DIVISION_BY_ZERO', message: 'Unexpected message' } },
    ],
    [201, { result: '1' }],
  ])('rejects unexpected envelopes/statuses %j', async (status, body) => {
    respond(body, Number(status))
    await expect(calculate('1', new AbortController().signal)).rejects.toThrow(
      requestFailure,
    )
  })
  it('handles transport failures', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new TypeError('offline')))
    await expect(calculate('1', new AbortController().signal)).rejects.toThrow(
      requestFailure,
    )
  })
  it.each(['timeout', 'cancel'])('aborts on %s', async (reason) => {
    vi.useFakeTimers()
    let aborted = false
    vi.stubGlobal(
      'fetch',
      vi.fn(
        (_url: string, init: RequestInit) =>
          new Promise((_resolve, reject) => {
            init.signal?.addEventListener('abort', () => {
              aborted = true
              reject(new DOMException('Aborted', 'AbortError'))
            })
          }),
      ),
    )
    const controller = new AbortController()
    const assertion = expect(calculate('1', controller.signal)).rejects.toThrow(
      requestFailure,
    )
    if (reason === 'timeout') await vi.advanceTimersByTimeAsync(10_000)
    else controller.abort()
    await assertion
    expect(aborted).toBe(true)
    expect(vi.getTimerCount()).toBe(0)
  })
})
