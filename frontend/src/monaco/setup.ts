import editorWorker from 'monaco-editor/esm/vs/editor/editor.worker?worker'

let initialized = false

export function ensureMonacoEnvironment() {
  if (initialized) return

  ;(self as typeof globalThis & {
    MonacoEnvironment?: {
      getWorker: () => Worker
    }
  }).MonacoEnvironment = {
    getWorker() {
      return new editorWorker()
    },
  }

  initialized = true
}
