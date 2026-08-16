import type { ReviewFindingDTO } from '@/api/modules/review'

export interface DiffLine {
  type: 'hunk' | 'add' | 'del' | 'ctx'
  content: string
  oldNum: string | number
  newNum: string | number
}

export interface DiffFile {
  file_path: string
  lines: DiffLine[]
}

export interface VisibleLinesResult {
  lines: DiffFile['lines']
  fmap: Map<number, ReviewFindingDTO[]>
}

export function parseDiff(rawDiff: string): DiffFile[] {
  if (!rawDiff) return []
  const files: DiffFile[] = []
  let current_file: DiffFile | null = null
  let oldNum = 0
  let newNum = 0
  let currentPath = ''

  for (const line of rawDiff.split('\n')) {
    if (line.startsWith('diff --git')) {
      const m = line.match(/diff --git a\/(.+?) b\/(.+)/)
      if (m) currentPath = m[2] || ''
      if (current_file) files.push(current_file)
      current_file = { file_path: currentPath, lines: [] }
      oldNum = 0
      newNum = 0
      continue
    }
    if (!current_file) continue
    if (line.startsWith('@@')) {
      const m = line.match(/@@ -(\d+)(?:,\d+)? \+(\d+)(?:,\d+)? @@/)
      if (m) { oldNum = parseInt(m[1] || '0'); newNum = parseInt(m[2] || '0') }
      current_file.lines.push({ type: 'hunk', content: line, oldNum: '', newNum: '' })
    } else if (line.startsWith('+')) {
      current_file.lines.push({ type: 'add', content: line.slice(1), oldNum: '', newNum: newNum++ })
    } else if (line.startsWith('-')) {
      current_file.lines.push({ type: 'del', content: line.slice(1), oldNum: oldNum++, newNum: '' })
    } else if (line.startsWith(' ') || line === '') {
      const content = line.startsWith(' ') ? line.slice(1) : ''
      current_file.lines.push({ type: 'ctx', content, oldNum: oldNum++, newNum: newNum++ })
    } else if (line.startsWith('---') || line.startsWith('+++') || line.startsWith('index') || line.startsWith('new file') || line.startsWith('deleted file') || line.startsWith('rename')) {
      continue
    }
  }
  if (current_file) files.push(current_file)
  return files
}

export function pathMatches(findingPath: string, diffPath: string): boolean {
  if (findingPath === diffPath) return true
  const fb = findingPath.split('/').pop() || ''
  const db = diffPath.split('/').pop() || ''
  if (fb && db && fb === db) return true
  if (diffPath.endsWith('/' + findingPath) || findingPath.endsWith('/' + diffPath)) return true
  return false
}

export function fileFindings(findings: ReviewFindingDTO[], file_path: string): ReviewFindingDTO[] {
  return findings.filter(f => f.file_path && pathMatches(f.file_path, file_path))
}

export function fileLevelFindings(findings: ReviewFindingDTO[], file_path: string): ReviewFindingDTO[] {
  return findings.filter(f => f.file_path && pathMatches(f.file_path, file_path) && (!f.new_line || f.new_line === 0))
}

export function globalFindings(findings: ReviewFindingDTO[]): ReviewFindingDTO[] {
  return findings.filter(f => !f.file_path)
}

export function fmapGet(fmap: Map<number, ReviewFindingDTO[]>, lineNum: number | string): ReviewFindingDTO[] {
  if (!lineNum || lineNum === '') return []
  const n = typeof lineNum === 'string' ? parseInt(lineNum) : lineNum
  if (isNaN(n)) return []
  return fmap.get(n) || []
}

export function buildFindingLineMap(file: DiffFile, findings: ReviewFindingDTO[]): Map<number, ReviewFindingDTO[]> {
  const ff = fileFindings(findings, file.file_path).filter(f => f.new_line && f.new_line > 0)
  if (ff.length === 0) return new Map()

  const codeLines: { idx: number; newNum: number; content: string; type: string }[] = []
  for (let i = 0; i < file.lines.length; i++) {
    const l = file.lines[i]
    if (!l || l.type === 'hunk') continue
    const n = typeof l.newNum === 'number' ? l.newNum : parseInt(String(l.newNum))
    if (!isNaN(n) && n > 0) {
      codeLines.push({ idx: i, newNum: n, content: l.content.toLowerCase(), type: l.type })
    }
  }

  const result = new Map<number, ReviewFindingDTO[]>()
  for (const f of ff) {
    const ln = f.new_line
    const exact = codeLines.find(c => c.newNum === ln)
    if (exact) {
      const arr = result.get(exact.newNum) || []
      arr.push(f)
      result.set(exact.newNum, arr)
      continue
    }

    const matched = findBestCodeLine(f, codeLines)
    if (matched) {
      const arr = result.get(matched.newNum) || []
      arr.push(f)
      result.set(matched.newNum, arr)
    }
  }
  return result
}

export function findBestCodeLine(
  f: ReviewFindingDTO,
  codeLines: { idx: number; newNum: number; content: string; type: string }[]
): { idx: number; newNum: number; content: string; type: string } | null {
  const keywords = extractKeywords(f.title + ' ' + f.message + ' ' + (f.suggestion || ''))
  if (keywords.length === 0) {
    return codeLines.length > 0 ? codeLines[0]! : null
  }

  let bestLine: typeof codeLines[0] | null = null
  let bestScore = -1

  for (const cl of codeLines) {
    if (cl.type === 'del') continue
    let score = 0
    for (const kw of keywords) {
      if (cl.content.includes(kw.toLowerCase())) score += kw.length
    }
    if (cl.type === 'add') score += 1
    if (score > bestScore) {
      bestScore = score
      bestLine = cl
    }
  }

  return bestLine
}

export function extractKeywords(text: string): string[] {
  const stopWords = new Set([
    'the', 'a', 'an', 'is', 'are', 'was', 'were', 'be', 'been', 'being',
    'have', 'has', 'had', 'do', 'does', 'did', 'will', 'would', 'could',
    'should', 'may', 'might', 'can', 'shall', 'to', 'of', 'in', 'for',
    'on', 'with', 'at', 'by', 'from', 'as', 'into', 'through', 'during',
    'before', 'after', 'above', 'below', 'between', 'out', 'off', 'over',
    'under', 'again', 'further', 'then', 'once', 'here', 'there', 'when',
    'where', 'why', 'how', 'all', 'both', 'each', 'few', 'more', 'most',
    'other', 'some', 'such', 'no', 'nor', 'not', 'only', 'own', 'same',
    'so', 'than', 'too', 'very', 'just', 'because', 'but', 'and', 'or',
    'if', 'while', 'this', 'that', 'these', 'those', 'it', 'its',
    'code', 'file', 'line', 'use', 'used', 'using', 'also', 'which',
    'about', 'up', 'like', 'what', 'get', 'set', 'new', 'add', 'make',
  ])
  const words = text.match(/[a-zA-Z_][a-zA-Z0-9_.]{2,}/g) || []
  const unique = [...new Set(words.filter(w => !stopWords.has(w.toLowerCase())))]
  return unique.filter(w => w.length >= 3).slice(0, 15)
}

export function fileIcon(path: string): string {
  if (path.endsWith('.go')) return '🔷'
  if (path.endsWith('.ts') || path.endsWith('.tsx')) return '🔷'
  if (path.endsWith('.js') || path.endsWith('.jsx')) return '🟨'
  if (path.endsWith('.vue')) return '💚'
  if (path.endsWith('.py')) return '🐍'
  if (path.endsWith('.rs')) return '🦀'
  if (path.endsWith('.css') || path.endsWith('.scss')) return '🎨'
  if (path.endsWith('.html')) return '📄'
  if (path.endsWith('.json') || path.endsWith('.yaml') || path.endsWith('.yml')) return '📋'
  if (path.endsWith('.md')) return '📝'
  if (path.endsWith('.sql')) return '🗃️'
  return '📄'
}

export function fileAddCount(file: DiffFile): number {
  return file.lines.filter(l => l.type === 'add').length
}

export function fileDelCount(file: DiffFile): number {
  return file.lines.filter(l => l.type === 'del').length
}

export function severityText(s: string): string {
  const m: Record<string, string> = { critical: '严重', high: '高危', medium: '中等', low: '低危', info: '提示' }
  return m[s] || s
}

export function statusLabel(s: string): string {
  const m: Record<string, string> = { pending: '等待中', running: '运行中', success: '通过', failed: '失败', blocked: '已阻止' }
  return m[s] || s
}

export function riskLabel(r: string): string {
  const m: Record<string, string> = { critical: '严重', high: '高危', medium: '中等', low: '低危', info: '提示' }
  return m[r] || r
}

export function triggerLabel(t: string): string {
  const m: Record<string, string> = { manual: '手动', webhook: 'Webhook', api: 'API' }
  return m[t] || t
}

export function shortSha(sha: string): string {
  return sha ? sha.substring(0, 7) : '—'
}

export function timeAgo(d: string): string {
  if (!d) return '—'
  const diff = Date.now() - new Date(d).getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return '刚刚'
  if (mins < 60) return `${mins} 分钟前`
  const h = Math.floor(mins / 60)
  if (h < 24) return `${h} 小时前`
  return `${Math.floor(h / 24)} 天前`
}

export type StatusVariant = 'success' | 'danger' | 'warning' | 'info' | 'running' | 'default'

export function statusVariant(s: string): StatusVariant {
  const m: Record<string, StatusVariant> = { pending: 'warning', running: 'running', success: 'success', failed: 'danger', blocked: 'danger' }
  return m[s] || 'default'
}

export function riskVariant(r: string): StatusVariant {
  const m: Record<string, StatusVariant> = { critical: 'danger', high: 'warning', medium: 'warning', low: 'info', info: 'default' }
  return m[r] || 'default'
}

export function severityVariant(s: string): StatusVariant {
  const m: Record<string, StatusVariant> = { critical: 'danger', high: 'warning', medium: 'warning', low: 'info', info: 'default' }
  return m[s] || 'default'
}
