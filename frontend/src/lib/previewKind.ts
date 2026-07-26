export type PreviewKind = 'image' | 'video' | 'audio' | 'pdf' | 'code' | 'text' | 'unsupported'

const IMAGE_EXT = new Set(['png', 'jpg', 'jpeg', 'gif', 'webp', 'svg', 'bmp', 'ico'])
const VIDEO_EXT = new Set(['mp4', 'webm', 'mov', 'm4v'])
const AUDIO_EXT = new Set(['mp3', 'wav', 'ogg', 'flac', 'm4a'])
const TEXT_EXT = new Set(['txt', 'log', 'cfg', 'ini', 'env'])

const CODE_LANG: Record<string, string> = {
  js: 'javascript',
  jsx: 'jsx',
  ts: 'typescript',
  tsx: 'tsx',
  go: 'go',
  py: 'python',
  rb: 'ruby',
  java: 'java',
  c: 'c',
  cpp: 'cpp',
  h: 'c',
  rs: 'rust',
  json: 'json',
  yaml: 'yaml',
  yml: 'yaml',
  toml: 'toml',
  md: 'markdown',
  sh: 'shell',
  html: 'html',
  css: 'css',
  scss: 'scss',
  sql: 'sql',
  php: 'php',
  swift: 'swift',
  kt: 'kotlin',
}

export function extOf(name: string): string {
  const i = name.lastIndexOf('.')
  return i <= 0 ? '' : name.slice(i + 1).toLowerCase()
}

// Classifies a file by extension into the preview types spec §5.4 covers.
// Extensionless files (READMEs, Makefiles, dotfiles) default to plain text.
export function previewKindOf(name: string): PreviewKind {
  const ext = extOf(name)
  if (IMAGE_EXT.has(ext)) return 'image'
  if (VIDEO_EXT.has(ext)) return 'video'
  if (AUDIO_EXT.has(ext)) return 'audio'
  if (ext === 'pdf') return 'pdf'
  if (ext in CODE_LANG) return 'code'
  if (TEXT_EXT.has(ext) || ext === '') return 'text'
  return 'unsupported'
}

// Maps a code file's extension to a Shiki language id.
export function codeLanguageOf(name: string): string {
  return CODE_LANG[extOf(name)] ?? 'text'
}
