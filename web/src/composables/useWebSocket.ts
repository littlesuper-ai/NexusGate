import { ref, onUnmounted } from 'vue'

export interface WSMessage {
  type: string
  data: any
  timestamp: string
}

// Singleton state — shared across all components
const connected = ref(false)
const lastMessage = ref<WSMessage | null>(null)
const handlers = new Map<string, Set<(data: any) => void>>()
let ws: WebSocket | null = null
let reconnectTimer: ReturnType<typeof setTimeout> | null = null
let refCount = 0
let reconnectAttempts = 0
const MAX_RECONNECT_ATTEMPTS = 30
const BASE_RECONNECT_DELAY = 2000

function connect() {
  if (ws && (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING)) return

  const token = localStorage.getItem('token')
  if (!token) {
    const delay = Math.min(BASE_RECONNECT_DELAY * Math.pow(1.5, reconnectAttempts), 30000)
    reconnectTimer = setTimeout(connect, delay)
    return
  }
  const proto = location.protocol === 'https:' ? 'wss:' : 'ws:'
  const url = `${proto}//${location.host}/ws?token=${encodeURIComponent(token)}`
  ws = new WebSocket(url)

  ws.onopen = () => {
    connected.value = true
    reconnectAttempts = 0
    if (reconnectTimer) { clearTimeout(reconnectTimer); reconnectTimer = null }
  }

  ws.onmessage = (event) => {
    try {
      const msg: WSMessage = JSON.parse(event.data)
      lastMessage.value = msg
      const fns = handlers.get(msg.type)
      if (fns) fns.forEach(fn => fn(msg.data))
    } catch { /* ignore non-JSON */ }
  }

  ws.onclose = () => {
    connected.value = false
    ws = null
    if (refCount > 0 && reconnectAttempts < MAX_RECONNECT_ATTEMPTS) {
      reconnectAttempts++
      const delay = Math.min(BASE_RECONNECT_DELAY * Math.pow(1.5, reconnectAttempts), 30000)
      reconnectTimer = setTimeout(connect, delay)
    }
  }

  ws.onerror = () => {
    ws?.close()
  }
}

function disconnect() {
  if (reconnectTimer) clearTimeout(reconnectTimer)
  reconnectTimer = null
  ws?.close()
  ws = null
  connected.value = false
}

/**
 * Shared WebSocket composable (singleton connection).
 * Multiple components can call useWebSocket() — only one connection is maintained.
 * Handlers registered with `on()` are automatically cleaned up on unmount.
 */
export function useWebSocket() {
  const localHandlers: Array<{ type: string; fn: (data: any) => void }> = []

  refCount++
  connect()

  const on = (type: string, handler: (data: any) => void) => {
    if (!handlers.has(type)) handlers.set(type, new Set())
    handlers.get(type)!.add(handler)
    localHandlers.push({ type, fn: handler })
  }

  const off = (type: string, handler: (data: any) => void) => {
    handlers.get(type)?.delete(handler)
  }

  onUnmounted(() => {
    // Clean up handlers registered by this component
    for (const { type, fn } of localHandlers) {
      handlers.get(type)?.delete(fn)
    }
    refCount--
    if (refCount <= 0) {
      refCount = 0
      disconnect()
    }
  })

  return { connected, lastMessage, on, off }
}
