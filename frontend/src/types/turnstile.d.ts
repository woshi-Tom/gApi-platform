type TurnstileTheme = 'auto' | 'light' | 'dark'
type TurnstileSize = 'normal' | 'compact' | 'flexible'

interface TurnstileRenderOptions {
  sitekey: string
  theme?: TurnstileTheme
  size?: TurnstileSize
  callback?: (token: string) => void
  'expired-callback'?: () => void
  'error-callback'?: () => void
  'timeout-callback'?: () => void
}

interface TurnstileAPI {
  render: (container: HTMLElement, options: TurnstileRenderOptions) => string
  reset: (widgetId?: string) => void
  remove: (widgetId: string) => void
}

interface Window {
  turnstile?: TurnstileAPI
}

interface ImportMetaEnv {
  readonly VITE_TURNSTILE_SITE_KEY?: string
}
