const themeStorageKey = 'eu5.theme'

export type AppTheme = 'dark' | 'light'

function normalizeTheme(value: string | null): AppTheme {
  return value === 'light' ? 'light' : 'dark'
}

export function applyTheme(theme: AppTheme): void {
  if (typeof document === 'undefined') {
    return
  }
  document.documentElement.setAttribute('data-theme', theme)
  window.localStorage.setItem(themeStorageKey, theme)
}

export function initializeTheme(): AppTheme {
  const stored = window.localStorage.getItem(themeStorageKey)
  const next = normalizeTheme(stored)
  applyTheme(next)
  return next
}

