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
  end_line?: number
  end_column?: number
  message: string
  severity: 'error' | 'warning' | 'info'
  rule_id: string
  rule_name?: string
  source?: 'rule' | 'ai'
  quick_fix?: string
}

export interface LintResult {
  file?: string
  issues: LintIssue[]
  stats: {
    error_count: number
    warning_count: number
    info_count: number
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
  priority?: number
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
  current_file: string | null
  content: string
  original_content: string
  is_dirty: boolean
  cursor_position: { line: number; column: number } | null
}

export interface FormatOptions {
  curlify?: boolean
  remove_clean?: boolean
  remove_build_root?: boolean
  remove_group?: boolean
  license_spdx?: boolean
  sort_deps?: boolean
  tab_to_spaces?: boolean
  indent_size?: number
  preamble_order?: boolean
  align_values?: boolean
  path_macros?: boolean
  util_macros?: boolean
  common_cleanup?: boolean
  conditional_trim?: boolean
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
