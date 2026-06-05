import { defineStore } from 'pinia'
import { ref } from 'vue'

// localStorage key for the persisted theme and the fallback when none is saved.
const STORAGE_KEY = 'store-theme'
const DEFAULT_THEME = 'light'

// Reflect the active theme onto <html data-theme="..."> so CSS variables switch.
function applyToDom(value) {
  if (value === 'dark') {
    document.documentElement.setAttribute('data-theme', 'dark')
  } else {
    document.documentElement.setAttribute('data-theme', 'light')
  }
}

export const useThemeStore = defineStore('theme', () => {
  // Active theme ('light' | 'dark'), starting from the default until init runs.
  const theme = ref(DEFAULT_THEME)

  // Load the saved theme on app start, validate it, and apply it to the DOM.
  function init() {
    const saved = localStorage.getItem(STORAGE_KEY)
    theme.value = saved === 'dark' || saved === 'light' ? saved : DEFAULT_THEME
    applyToDom(theme.value)
  }

  // Set the theme (anything but 'dark' normalizes to 'light'), persist, and apply.
  function set(value) {
    theme.value = value === 'dark' ? 'dark' : 'light'
    localStorage.setItem(STORAGE_KEY, theme.value)
    applyToDom(theme.value)
  }

  // Flip between light and dark.
  function toggle() {
    set(theme.value === 'dark' ? 'light' : 'dark')
  }

  return { theme, init, set, toggle }
})
