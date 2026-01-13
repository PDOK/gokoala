export function safeGetCurrentUrl(urlInput: string | undefined) {
  let url = new URL(globalThis.location.href)
  if (urlInput != undefined) {
    url = new URL(urlInput)
  }
  return url
}

/**
 * Safe, strictly-typed replacement for history.replaceState
 * Angular-compatible, strict-mode safe
 */

export function safeReplaceState<TState = unknown>(state: TState, title: string = '', url?: string | null): void {
  if (url == null || url === '') {
    globalThis.history.replaceState(state, title)
    return
  }

  let target: URL

  try {
    target = new URL(url, globalThis.location.href)
  } catch {
    // Invalid URL → state-only update
    globalThis.history.replaceState(state, title)
    return
  }

  const current = globalThis.location

  const isSameOrigin = target.protocol === current.protocol && target.hostname === current.hostname && target.port === current.port

  if (isSameOrigin) {
    const safePath = target.pathname + target.search + target.hash

    globalThis.history.replaceState(state, title, safePath)
  } else {
    // Cross-origin → no URL update
    globalThis.history.replaceState(state, title)
  }
}

/**
 * Safe, strictly-typed replacement for history.pushState
 * Angular-compatible, strict-mode safe
 */
export function safePushState<TState = unknown>(state: TState, title: string = '', url?: string | null): void {
  if (url == null || url === '') {
    globalThis.history.pushState(state, title)
    return
  }

  let target: URL

  try {
    target = new URL(url, globalThis.location.href)
  } catch {
    globalThis.history.pushState(state, title)
    return
  }

  const current = globalThis.location

  const isSameOrigin = target.protocol === current.protocol && target.hostname === current.hostname && target.port === current.port

  if (isSameOrigin) {
    const safePath = target.pathname + target.search + target.hash

    globalThis.history.pushState(state, title, safePath)
  } else {
    // Cross-origin → state only
    globalThis.history.pushState(state, title)
  }
}
