import { ref, onUnmounted } from 'vue'

export interface WSMessage {
  type: string
  data: any
  timestamp: string
}

export function useWebSocket() {
  const connected = ref(false)
  const lastMessage = ref<WSMessage | null>(null)
  const handlers = new Map<string, Set<(data: any) => void>>()
  let ws: WebSocket | null = null
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null

  const connect = () => {
    const token = localStorage.getItem('token')
    if (!token) {
      // No token — retry after delay (user might not be logged in yet)
      reconnectTimer = setTimeout(connect, 3000)
      return
    }
    const proto = location.protocol === 'https:' ? 'wss:' : 'ws:'
    const url = `${proto}//${location.host}/ws?token=${encodeURIComponent(token)}`
    ws = new WebSocket(url)

    ws.onopen = () => {
      connected.value = true
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
      reconnectTimer = setTimeout(connect, 3000)
    }

    ws.onerror = () => {
      ws?.close()
    }
  }

  const on = (type: string, handler: (data: any) => void) => {
    if (!handlers.has(type)) handlers.set(type, new Set())
    handlers.get(type)!.add(handler)
  }

  const off = (type: string, handler: (data: any) => void) => {
    handlers.get(type)?.delete(handler)
  }

  const disconnect = () => {
    if (reconnectTimer) clearTimeout(reconnectTimer)
    ws?.close()
    ws = null
    connected.value = false
  }

  connect()

  onUnmounted(disconnect)

  return { connected, lastMessage, on, off, disconnect }
}
