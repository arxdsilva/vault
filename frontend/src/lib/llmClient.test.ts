import { afterEach, describe, expect, it, vi } from 'vitest'
import { sendMessage } from './llmClient'

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), { status })
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('sendMessage — anthropic', () => {
  it('returns the final text when the model replies without using a tool', async () => {
    const fetchMock = vi.fn(async () =>
      jsonResponse({ stop_reason: 'end_turn', content: [{ type: 'text', text: 'Hi there' }] }),
    )
    vi.stubGlobal('fetch', fetchMock)

    const result = await sendMessage({
      provider: 'anthropic',
      apiKey: 'sk-ant-test',
      history: [],
      userText: 'hello',
    })

    expect(result).toEqual({ role: 'assistant', text: 'Hi there', toolCalls: [] })
    expect(fetchMock).toHaveBeenCalledTimes(1)
    const [url, init] = fetchMock.mock.calls[0] as unknown as [string, RequestInit]
    expect(url).toBe('https://api.anthropic.com/v1/messages')
    expect((init.headers as Record<string, string>)['x-api-key']).toBe('sk-ant-test')
  })

  it('executes a requested tool, then continues the loop for the final reply', async () => {
    const execute = vi.fn(async () => ({ entries: [] }))
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        jsonResponse({
          stop_reason: 'tool_use',
          content: [{ type: 'tool_use', id: 'call_1', name: 'list_directory', input: { path: '/home' } }],
        }),
      )
      .mockResolvedValueOnce(
        jsonResponse({ stop_reason: 'end_turn', content: [{ type: 'text', text: 'It is empty' }] }),
      )
    vi.stubGlobal('fetch', fetchMock)

    const result = await sendMessage({
      provider: 'anthropic',
      apiKey: 'sk-ant-test',
      history: [],
      userText: "what's in my home folder?",
      execute,
    })

    expect(execute).toHaveBeenCalledWith('list_directory', { path: '/home' })
    expect(fetchMock).toHaveBeenCalledTimes(2)
    expect(result.text).toBe('It is empty')
    expect(result.toolCalls).toEqual([{ name: 'list_directory', input: { path: '/home' }, result: { entries: [] } }])
  })

  it('records a tool error instead of throwing, and keeps going', async () => {
    const execute = vi.fn(async () => {
      throw new Error('boom')
    })
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        jsonResponse({
          stop_reason: 'tool_use',
          content: [{ type: 'tool_use', id: 'call_1', name: 'delete_file', input: { path: '/x' } }],
        }),
      )
      .mockResolvedValueOnce(jsonResponse({ stop_reason: 'end_turn', content: [{ type: 'text', text: 'failed' }] }))
    vi.stubGlobal('fetch', fetchMock)

    const result = await sendMessage({
      provider: 'anthropic',
      apiKey: 'k',
      history: [],
      userText: 'delete /x',
      execute,
    })

    expect(result.toolCalls?.[0].result).toEqual({ error: 'Error: boom' })
  })

  it('throws on a non-OK HTTP response', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => jsonResponse({}, 401)))

    await expect(
      sendMessage({ provider: 'anthropic', apiKey: 'bad', history: [], userText: 'hi' }),
    ).rejects.toThrow('HTTP 401')
  })

  it('gives up after too many tool-call rounds', async () => {
    const fetchMock = vi.fn(async () =>
      jsonResponse({
        stop_reason: 'tool_use',
        content: [{ type: 'tool_use', id: 'call_1', name: 'list_directory', input: {} }],
      }),
    )
    vi.stubGlobal('fetch', fetchMock)

    await expect(
      sendMessage({
        provider: 'anthropic',
        apiKey: 'k',
        history: [],
        userText: 'loop forever',
        execute: vi.fn(async () => ({})),
      }),
    ).rejects.toThrow('too many tool-call rounds')
  })
})

describe('sendMessage — openai', () => {
  it('returns the final text when the model replies without a tool call', async () => {
    const fetchMock = vi.fn(async () =>
      jsonResponse({ choices: [{ message: { role: 'assistant', content: 'Hi there' } }] }),
    )
    vi.stubGlobal('fetch', fetchMock)

    const result = await sendMessage({ provider: 'openai', apiKey: 'sk-oai-test', history: [], userText: 'hello' })

    expect(result).toEqual({ role: 'assistant', text: 'Hi there', toolCalls: [] })
    const [url, init] = fetchMock.mock.calls[0] as unknown as [string, RequestInit]
    expect(url).toBe('https://api.openai.com/v1/chat/completions')
    expect((init.headers as Record<string, string>).Authorization).toBe('Bearer sk-oai-test')
  })

  it('executes a requested tool call, then continues for the final reply', async () => {
    const execute = vi.fn(async () => ({ ok: true }))
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        jsonResponse({
          choices: [
            {
              message: {
                role: 'assistant',
                content: null,
                tool_calls: [{ id: 'call_1', function: { name: 'delete_file', arguments: '{"path":"/x"}' } }],
              },
            },
          ],
        }),
      )
      .mockResolvedValueOnce(jsonResponse({ choices: [{ message: { role: 'assistant', content: 'done' } }] }))
    vi.stubGlobal('fetch', fetchMock)

    const result = await sendMessage({
      provider: 'openai',
      apiKey: 'k',
      history: [],
      userText: 'delete /x',
      execute,
    })

    expect(execute).toHaveBeenCalledWith('delete_file', { path: '/x' })
    expect(result.text).toBe('done')
    expect(result.toolCalls).toEqual([{ name: 'delete_file', input: { path: '/x' }, result: { ok: true } }])
  })

  it('throws on a non-OK HTTP response', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => jsonResponse({}, 500)))

    await expect(sendMessage({ provider: 'openai', apiKey: 'k', history: [], userText: 'hi' })).rejects.toThrow(
      'HTTP 500',
    )
  })
})
