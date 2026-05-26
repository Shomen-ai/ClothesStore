import { defineStore } from 'pinia'
import { ref } from 'vue'

const STORAGE_KEY = 'store-theme'
const DEFAULT_THEME = 'light'

function applyToDom(value) {
  if (value === 'dark') {
    document.documentElement.setAttribute('data-theme', 'dark')
  } else {
    document.documentElement.setAttribute('data-theme', 'light')
  }
}

export const useThemeStore = defineStore('theme', () => {
  const theme = ref(DEFAULT_THEME)

  function init() {
    const saved = localStorage.getItem(STORAGE_KEY)
    theme.value = saved === 'dark' || saved === 'light' ? saved : DEFAULT_THEME
    applyToDom(theme.value)
  }

  function set(value) {
    theme.value = value === 'dark' ? 'dark' : 'light'
    localStorage.setItem(STORAGE_KEY, theme.value)
    applyToDom(theme.value)
  }

  function toggle() {
    set(theme.value === 'dark' ? 'light' : 'dark')
  }

  return { theme, init, set, toggle }
})
