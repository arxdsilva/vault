import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('../lib/llmClient', async () => {
  const actual = await vi.importActual<typeof import('../lib/llmClient')>('../lib/llmClient')
  return { ...actual, sendMessage: vi.fn() }
})

import { sendMessage } from '../lib/llmClient'
import { useChat } from './chat'

const sendMessageMock = vi.mocked(sendMessage)

describe('useChat', () => {
  beforeEach(() => {
    localStorage.clear()
    sendMessageMock.mockReset()
    useChat.setState({
      provider: 'anthropic',
      apiKey: '',
      model: '',
      messages: [],
      sending: false,
      error: null,
    })
  })

  it('setApiKey updates state and persists to localStorage', () => {
    useChat.getState().setApiKey('sk-test')
    expect(useChat.getState().apiKey).toBe('sk-test')
    expect(JSON.parse(localStorage.getItem('vault:chat')!).apiKey).toBe('sk-test')
  })

  it('setProvider updates state and persists to localStorage', () => {
    useChat.getState().setProvider('openai')
    expect(useChat.getState().provider).toBe('openai')
    expect(JSON.parse(localStorage.getItem('vault:chat')!).provider).toBe('openai')
  })

  it('setModel updates state and persists to localStorage', () => {
    useChat.getState().setModel('gpt-4o-mini')
    expect(useChat.getState().model).toBe('gpt-4o-mini')
    expect(JSON.parse(localStorage.getItem('vault:chat')!).model).toBe('gpt-4o-mini')
  })

  it('send is a no-op without an API key', async () => {
    await useChat.getState().send('hello')
    expect(sendMessageMock).not.toHaveBeenCalled()
    expect(useChat.getState().messages).toEqual([])
  })

  it('send is a no-op for blank text', async () => {
    useChat.getState().setApiKey('sk-test')
    await useChat.getState().send('   ')
    expect(sendMessageMock).not.toHaveBeenCalled()
  })

  it('send appends the user message, then the assistant reply', async () => {
    useChat.getState().setApiKey('sk-test')
    sendMessageMock.mockResolvedValue({ role: 'assistant', text: 'hi back', toolCalls: [] })

    const promise = useChat.getState().send('hello')
    expect(useChat.getState().sending).toBe(true)
    expect(useChat.getState().messages).toEqual([{ role: 'user', text: 'hello' }])

    await promise

    expect(useChat.getState().sending).toBe(false)
    expect(useChat.getState().messages).toEqual([
      { role: 'user', text: 'hello' },
      { role: 'assistant', text: 'hi back', toolCalls: [] },
    ])
  })

  it('send records an error and stops sending, keeping the user message', async () => {
    useChat.getState().setApiKey('sk-test')
    sendMessageMock.mockRejectedValue(new Error('network down'))

    await useChat.getState().send('hello')

    expect(useChat.getState().sending).toBe(false)
    expect(useChat.getState().error).toContain('network down')
    expect(useChat.getState().messages).toEqual([{ role: 'user', text: 'hello' }])
  })

  it('clear resets messages and error', () => {
    useChat.setState({ messages: [{ role: 'user', text: 'x' }], error: 'oops' })
    useChat.getState().clear()
    expect(useChat.getState().messages).toEqual([])
    expect(useChat.getState().error).toBeNull()
  })
})
