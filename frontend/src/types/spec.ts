export interface SpecFileNode {
  name: string
  path: string
  is_dir: boolean
  children?: SpecFileNode[]
  size?: number
  mod_time?: string
}

export interface LintIssue {
  line: number
  column?: number
  endLine?: number
  endColumn?: number
  message: string
  severity: 'error' | 'warning' | 'info'
  ruleId: string
  ruleName?: string
  source?: 'rule' | 'ai'
  quickFix?: string
}

export interface LintResult {
  file?: string
  issues: LintIssue[]
  stats: {
    errorCount: number
    warningCount: number
    infoCount: number
  }
}

export type LintMode = 'rule_only' | 'rule_and_ai' | 'ai_only'

export type RuleCategory = 'required' | 'style' | 'best-practice' | 'custom'

export interface LintRule {
  id: string
  name: string
  description: string
  category: RuleCategory
  enabled: boolean
  severity: 'error' | 'warning' | 'info'
  pattern?: string
}

export interface SaveRequest {
  content: string
  message?: string
}

export interface CommitRequest {
  message: string
  content?: string
}

export interface CommitResponse {
  commit_hash: string
  message: string
}

export interface EditorState {
  currentFile: string | null
  content: string
  originalContent: string
  isDirty: boolean
  cursorPosition: { line: number; column: number } | null
}

export interface FormatOptions {
  curlify?: boolean
  removeClean?: boolean
  removeBuildRoot?: boolean
  removeGroup?: boolean
  licenseSpdx?: boolean
  sortDeps?: boolean
  tabToSpaces?: boolean
  indentSize?: number
  preambleOrder?: boolean
  alignValues?: boolean
  pathMacros?: boolean
  utilMacros?: boolean
  commonCleanup?: boolean
  conditionalTrim?: boolean
}

export interface FormatChange {
  line: number
  type: string
  before: string
  after: string
  reason: string
}

export interface FormatResult {
  content: string
  changes: FormatChange[]
}
