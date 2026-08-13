import { ref, computed } from 'vue'
import {
  getFileTree,
  type TreeEntry,
} from '@/api/modules/file'

export interface FlatTreeItem {
  name: string
  path: string
  isDir: boolean
  depth: number
}

export function useFileTree(repo_key: string) {
  const entries = ref<TreeEntry[]>([])
  const treeLoading = ref(false)
  const expandedDirs = ref(new Set<string>())
  const subTreeCache = ref<Record<string, TreeEntry[]>>({})

  const flatTreeItems = computed<FlatTreeItem[]>(() => {
    const items: FlatTreeItem[] = []
    function walk(list: TreeEntry[], depth: number) {
      const sorted = [...list].sort((a, b) => {
        if (a.type !== b.type) return a.type === 'dir' ? -1 : 1
        return a.name.localeCompare(b.name)
      })
      for (const e of sorted) {
        items.push({ name: e.name, path: e.path, isDir: e.type === 'dir', depth })
        if (e.type === 'dir' && expandedDirs.value.has(e.path)) {
          const children = subTreeCache.value[e.path]
          if (children) walk(children, depth + 1)
        }
      }
    }
    walk(entries.value, 0)
    return items
  })

  async function loadTree(ref: string) {
    treeLoading.value = true
    try {
      const res = await getFileTree(repo_key, {
        ref: ref || undefined,
      })
      entries.value = res.entries || []
    } catch {
      entries.value = []
    } finally {
      treeLoading.value = false
    }
  }

  async function loadSubTree(path: string, ref: string) {
    try {
      const res = await getFileTree(repo_key, {
        ref: ref || undefined,
        path,
      })
      subTreeCache.value = { ...subTreeCache.value, [path]: res.entries || [] }
    } catch { /* ignore */ }
  }

  async function toggleDir(path: string) {
    const s = new Set(expandedDirs.value)
    if (s.has(path)) {
      s.delete(path)
    } else {
      s.add(path)
      if (!subTreeCache.value[path]) {
        await loadSubTree(path, '')
      }
    }
    expandedDirs.value = s
  }

  function collapseAll() {
    expandedDirs.value = new Set()
    subTreeCache.value = {}
  }

  function resetExpanded() {
    expandedDirs.value = new Set()
    subTreeCache.value = {}
  }

  return {
    entries,
    treeLoading,
    expandedDirs,
    subTreeCache,
    flatTreeItems,
    loadTree,
    toggleDir,
    collapseAll,
    resetExpanded,
  }
}
