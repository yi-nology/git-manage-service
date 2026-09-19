import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

/**
 * UI 全局状态管理
 */
export const useUIStore = defineStore('ui', () => {
  // 深色模式
  const isDarkMode = ref(false)

  // 从 localStorage 恢复状态
  const restoreState = () => {
    const savedDarkMode = localStorage.getItem('ui:darkMode')
    if (savedDarkMode !== null) {
      isDarkMode.value = savedDarkMode === 'true'
    }
  }

  // 切换深色模式
  const toggleDarkMode = () => {
    isDarkMode.value = !isDarkMode.value
    applyDarkMode()
  }

  // 应用深色模式
  const applyDarkMode = () => {
    const html = document.documentElement
    if (isDarkMode.value) {
      html.classList.add('dark')
    } else {
      html.classList.remove('dark')
    }
    localStorage.setItem('ui:darkMode', String(isDarkMode.value))
  }

  // 监听深色模式变化
  watch(isDarkMode, () => {
    applyDarkMode()
  })

  return {
    // 状态
    isDarkMode,

    // 方法
    toggleDarkMode,
    restoreState,
    applyDarkMode,
  }
})
