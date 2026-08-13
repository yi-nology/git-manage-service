import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { SpecFileNode, LintIssue, LintRule } from '@/types/spec'

export const useSpecStore = defineStore('spec', () => {
  const fileTree = ref<SpecFileNode[]>([])
  const current_file = ref<string | null>(null)
  const content = ref('')
  const original_content = ref('')
  const is_dirty = ref(false)
  const lintIssues = ref<LintIssue[]>([])
  const rules = ref<LintRule[]>([])
  const loading = ref(false)
  const editorReady = ref(false)

  function setFileTree(tree: SpecFileNode[]) {
    fileTree.value = tree
  }

  function setCurrentFile(path: string | null) {
    current_file.value = path
  }

  function setContent(newContent: string) {
    content.value = newContent
    is_dirty.value = newContent !== original_content.value
  }

  function setOriginalContent(original: string) {
    original_content.value = original
    is_dirty.value = content.value !== original
  }

  function resetContent() {
    content.value = original_content.value
    is_dirty.value = false
  }

  function setLintIssues(issues: LintIssue[]) {
    lintIssues.value = issues
  }

  function clearLintIssues() {
    lintIssues.value = []
  }

  function setRules(newRules: LintRule[]) {
    rules.value = newRules
  }

  function update_rule(id: string, data: Partial<LintRule>) {
    const index = rules.value.findIndex((r) => r.id === id)
    if (index !== -1 && rules.value[index]) {
      const rule = rules.value[index]
      Object.assign(rule, data)
    }
  }

  function setLoading(value: boolean) {
    loading.value = value
  }

  function setEditorReady(value: boolean) {
    editorReady.value = value
  }

  function getEnabledRuleIds() {
    return rules.value.filter((r) => r.enabled).map((r) => r.id)
  }

  return {
    fileTree,
    current_file,
    content,
    original_content,
    is_dirty,
    lintIssues,
    rules,
    loading,
    editorReady,
    setFileTree,
    setCurrentFile,
    setContent,
    setOriginalContent,
    resetContent,
    setLintIssues,
    clearLintIssues,
    setRules,
    update_rule,
    setLoading,
    setEditorReady,
    getEnabledRuleIds,
  }
})
