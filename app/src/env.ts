/** API origin for browser calls. Production = same-origin (empty). Development = local Go server. */
export function getApiBaseUrl(): string {
  const configured = import.meta.env.VITE_API_BASE_URL?.replace(/\/+$/, '')
  if (configured) return configured
  if (import.meta.env.MODE === 'development') return 'http://localhost:8080'
  return ''
}

/** Pastebin API (separate service in Docker). Dev default matches server/cmd/pastebin :8081. */
export function getPastebinApiBaseUrl(): string {
  const configured = import.meta.env.VITE_PASTEBIN_API_BASE_URL?.replace(/\/+$/, '')
  if (configured) return configured
  if (import.meta.env.MODE === 'development') return 'http://localhost:8081'
  return ''
}
