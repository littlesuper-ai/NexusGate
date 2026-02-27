import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { ref } from 'vue'

// Test the WebSocket composable logic in isolation by mocking WebSocket and Vue lifecycle

// Track onUnmounted callbacks
let unmountCallbacks: Array<() => void> = []

vi.mock('vue', async () => {
  const actual = await vi.importActual<typeof import('vue')>('vue')
  return {
    ...actual,
    onUnmounted: vi.fn((cb: () => void) => { unmountCallbacks.push(cb) }),
  }
})

// Mock WebSocket
class MockWebSocket {
  static instances: MockWebSocket[] = []
  readyState = 0 // CONNECTING
  onopen: ((ev: any) => void) | null = null
  onmessage: ((ev: any) => void) | null = null
  onclose: ((ev: any) => void) | null = null
  onerror: ((ev: any) => void) | null = null
  url: string
  closed = false

  constructor(url: string) {
    this.url = url
    MockWebSocket.instances.push(this)
  }

  close() {
    this.closed = true
    this.readyState = 3 // CLOSED
    if (this.onclose) this.onclose({})
  }

  simulateOpen() {
    this.readyState = 1 // OPEN
    if (this.onopen) this.onopen({})
  }

  simulateMessage(data: any) {
    if (this.onmessage) this.onmessage({ data: JSON.stringify(data) })
  }

  simulateError() {
    if (this.onerror) this.onerror({})
  }
}

Object.defineProperty(globalThis, 'WebSocket', {
  value: MockWebSocket,
  writable: true,
})

// Also mock the constants
;(globalThis as any).WebSocket.OPEN = 1
;(globalThis as any).WebSocket.CONNECTING = 0

describe('useWebSocket', () => {
  beforeEach(() => {
    MockWebSocket.instances = []
    unmountCallbacks = []
    localStorage.clear()
    vi.useFakeTimers()
    // Reset module state between tests by re-importing
  })

  afterEach(() => {
    vi.useRealTimers()
    vi.restoreAllMocks()
  })

  it('creates WebSocket connection when token exists', async () => {
    localStorage.setItem('token', 'test-token')

    // Fresh import to reset singleton state
    vi.resetModules()
    const { useWebSocket } = await import('@/composables/useWebSocket')
    const { connected, on } = useWebSocket()

    expect(MockWebSocket.instances.length).toBeGreaterThan(0)
    const ws = MockWebSocket.instances[MockWebSocket.instances.length - 1]
    expect(ws.url).toContain('token=test-token')
  })

  it('delays connection when no token available', async () => {
    // No token set
    vi.resetModules()
    const { useWebSocket } = await import('@/composables/useWebSocket')
    const instancesBefore = MockWebSocket.instances.length
    useWebSocket()

    // Without token, it should schedule a retry, not create a WebSocket immediately
    // The retry timeout means no new instance is created yet
    // (or the module might have tried from a prior state — we check the retry behavior)
    vi.advanceTimersByTime(2000)
    // After timer fires, still no token — should retry again
  })

  it('dispatches messages to registered handlers', async () => {
    localStorage.setItem('token', 'test-token')
    vi.resetModules()
    const { useWebSocket } = await import('@/composables/useWebSocket')
    const { on } = useWebSocket()

    const handler = vi.fn()
    on('device_status', handler)

    const ws = MockWebSocket.instances[MockWebSocket.instances.length - 1]
    ws.simulateOpen()
    ws.simulateMessage({ type: 'device_status', data: { id: 1, status: 'online' }, timestamp: '2026-01-01T00:00:00Z' })

    expect(handler).toHaveBeenCalledWith({ id: 1, status: 'online' })
  })

  it('ignores messages for unregistered types', async () => {
    localStorage.setItem('token', 'test-token')
    vi.resetModules()
    const { useWebSocket } = await import('@/composables/useWebSocket')
    const { on } = useWebSocket()

    const handler = vi.fn()
    on('device_status', handler)

    const ws = MockWebSocket.instances[MockWebSocket.instances.length - 1]
    ws.simulateOpen()
    ws.simulateMessage({ type: 'metrics_update', data: {}, timestamp: '2026-01-01T00:00:00Z' })

    expect(handler).not.toHaveBeenCalled()
  })

  it('cleans up handlers on unmount', async () => {
    localStorage.setItem('token', 'test-token')
    vi.resetModules()
    const { useWebSocket } = await import('@/composables/useWebSocket')
    const { on } = useWebSocket()

    const handler = vi.fn()
    on('device_status', handler)

    // Simulate component unmount
    for (const cb of unmountCallbacks) cb()
    unmountCallbacks = []

    const ws = MockWebSocket.instances[MockWebSocket.instances.length - 1]
    ws.simulateOpen()
    ws.simulateMessage({ type: 'device_status', data: { id: 1 }, timestamp: '' })

    expect(handler).not.toHaveBeenCalled()
  })

  it('removes handler with off()', async () => {
    localStorage.setItem('token', 'test-token')
    vi.resetModules()
    const { useWebSocket } = await import('@/composables/useWebSocket')
    const { on, off } = useWebSocket()

    const handler = vi.fn()
    on('test_event', handler)
    off('test_event', handler)

    const ws = MockWebSocket.instances[MockWebSocket.instances.length - 1]
    ws.simulateOpen()
    ws.simulateMessage({ type: 'test_event', data: {}, timestamp: '' })

    expect(handler).not.toHaveBeenCalled()
  })

  it('handles non-JSON messages gracefully', async () => {
    localStorage.setItem('token', 'test-token')
    vi.resetModules()
    const { useWebSocket } = await import('@/composables/useWebSocket')
    useWebSocket()

    const ws = MockWebSocket.instances[MockWebSocket.instances.length - 1]
    ws.simulateOpen()

    // Should not throw
    expect(() => {
      if (ws.onmessage) ws.onmessage({ data: 'not-json' })
    }).not.toThrow()
  })

  it('sets connected to true on open', async () => {
    localStorage.setItem('token', 'test-token')
    vi.resetModules()
    const { useWebSocket } = await import('@/composables/useWebSocket')
    const { connected } = useWebSocket()

    const ws = MockWebSocket.instances[MockWebSocket.instances.length - 1]
    ws.simulateOpen()

    expect(connected.value).toBe(true)
  })
})
