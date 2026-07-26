import { CHAT_TOOLS, executeTool, type ToolSpec } from './chatTools'

export type Provider = 'anthropic' | 'openai'
export type ChatRole = 'user' | 'assistant'

export interface ToolCallRecord {
  name: string
  input: Record<string, unknown>
  result: unknown
}

export interface ChatMessage {
  role: ChatRole
  text: string
  toolCalls?: ToolCallRecord[]
}

const DEFAULT_ANTHROPIC_MODEL = 'claude-sonnet-5'
const DEFAULT_OPENAI_MODEL = 'gpt-4o'

// Caps how many tool round-trips a single turn can take before giving up,
// so a confused model can't loop forever burning the user's API quota.
const MAX_TOOL_ROUNDS = 8

export function defaultModelFor(provider: Provider): string {
  return provider === 'anthropic' ? DEFAULT_ANTHROPIC_MODEL : DEFAULT_OPENAI_MODEL
}

type ExecuteTool = (name: string, input: Record<string, unknown>) => Promise<unknown>

export interface SendOptions {
  provider: Provider
  apiKey: string
  model?: string
  history: ChatMessage[] // prior turns, oldest first, tool calls not replayed
  userText: string
  tools?: ToolSpec[]
  execute?: ExecuteTool
}

export async function sendMessage(opts: SendOptions): Promise<ChatMessage> {
  const tools = opts.tools ?? CHAT_TOOLS
  const execute = opts.execute ?? executeTool
  const model = opts.model || defaultModelFor(opts.provider)

  return opts.provider === 'anthropic'
    ? sendAnthropic(opts.apiKey, model, opts.history, opts.userText, tools, execute)
    : sendOpenAI(opts.apiKey, model, opts.history, opts.userText, tools, execute)
}

async function runTool(
  name: string,
  input: Record<string, unknown>,
  execute: ExecuteTool,
  toolCalls: ToolCallRecord[],
): Promise<unknown> {
  try {
    const result = await execute(name, input)
    toolCalls.push({ name, input, result })
    return result
  } catch (err) {
    const result = { error: String(err) }
    toolCalls.push({ name, input, result })
    return result
  }
}

interface AnthropicContentBlock {
  type: 'text' | 'tool_use' | 'tool_result'
  text?: string
  id?: string
  name?: string
  input?: Record<string, unknown>
  tool_use_id?: string
  content?: string
}

interface AnthropicMessage {
  role: 'user' | 'assistant'
  content: string | AnthropicContentBlock[]
}

async function sendAnthropic(
  apiKey: string,
  model: string,
  history: ChatMessage[],
  userText: string,
  tools: ToolSpec[],
  execute: ExecuteTool,
): Promise<ChatMessage> {
  const messages: AnthropicMessage[] = [
    ...history.map((m) => ({ role: m.role, content: m.text })),
    { role: 'user' as const, content: userText },
  ]
  const anthropicTools = tools.map((t) => ({
    name: t.name,
    description: t.description,
    input_schema: t.inputSchema,
  }))
  const toolCalls: ToolCallRecord[] = []

  for (let round = 0; round < MAX_TOOL_ROUNDS; round++) {
    const res = await fetch('https://api.anthropic.com/v1/messages', {
      method: 'POST',
      headers: {
        'x-api-key': apiKey,
        'anthropic-version': '2023-06-01',
        'content-type': 'application/json',
      },
      body: JSON.stringify({ model, max_tokens: 1024, messages, tools: anthropicTools }),
    })
    if (!res.ok) throw new Error(`Anthropic API error: HTTP ${res.status}`)
    const data = await res.json()
    const content: AnthropicContentBlock[] = data.content ?? []

    if (data.stop_reason !== 'tool_use') {
      const text = content
        .filter((b) => b.type === 'text')
        .map((b) => b.text ?? '')
        .join('')
      return { role: 'assistant', text, toolCalls }
    }

    messages.push({ role: 'assistant', content })
    const toolResults: AnthropicContentBlock[] = []
    for (const block of content) {
      if (block.type !== 'tool_use') continue
      const result = await runTool(block.name ?? '', block.input ?? {}, execute, toolCalls)
      toolResults.push({ type: 'tool_result', tool_use_id: block.id, content: JSON.stringify(result) })
    }
    messages.push({ role: 'user', content: toolResults })
  }

  throw new Error('too many tool-call rounds without a final response')
}

interface OpenAIToolCall {
  id: string
  function: { name: string; arguments: string }
}

interface OpenAIMessage {
  role: 'user' | 'assistant' | 'tool'
  content?: string | null
  tool_calls?: OpenAIToolCall[]
  tool_call_id?: string
}

async function sendOpenAI(
  apiKey: string,
  model: string,
  history: ChatMessage[],
  userText: string,
  tools: ToolSpec[],
  execute: ExecuteTool,
): Promise<ChatMessage> {
  const messages: OpenAIMessage[] = [
    ...history.map((m) => ({ role: m.role, content: m.text })),
    { role: 'user' as const, content: userText },
  ]
  const openaiTools = tools.map((t) => ({
    type: 'function',
    function: { name: t.name, description: t.description, parameters: t.inputSchema },
  }))
  const toolCalls: ToolCallRecord[] = []

  for (let round = 0; round < MAX_TOOL_ROUNDS; round++) {
    const res = await fetch('https://api.openai.com/v1/chat/completions', {
      method: 'POST',
      headers: { Authorization: `Bearer ${apiKey}`, 'content-type': 'application/json' },
      body: JSON.stringify({ model, messages, tools: openaiTools }),
    })
    if (!res.ok) throw new Error(`OpenAI API error: HTTP ${res.status}`)
    const data = await res.json()
    const message = data.choices?.[0]?.message ?? {}

    if (!message.tool_calls || message.tool_calls.length === 0) {
      return { role: 'assistant', text: message.content ?? '', toolCalls }
    }

    messages.push({ role: 'assistant', content: message.content ?? null, tool_calls: message.tool_calls })
    for (const call of message.tool_calls as OpenAIToolCall[]) {
      const input = call.function.arguments ? JSON.parse(call.function.arguments) : {}
      const result = await runTool(call.function.name, input, execute, toolCalls)
      messages.push({ role: 'tool', tool_call_id: call.id, content: JSON.stringify(result) })
    }
  }

  throw new Error('too many tool-call rounds without a final response')
}
