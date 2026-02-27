import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'

// Hoisted mocks — available before vi.mock factory runs
const { mockGetDevices, mockElMessage } = vi.hoisted(() => ({
  mockGetDevices: vi.fn(),
  mockElMessage: { error: vi.fn(), warning: vi.fn(), success: vi.fn(), info: vi.fn() },
}))

vi.mock('../api', () => ({
  getDevices: (...args: any[]) => mockGetDevices(...args),
  rebootDevice: vi.fn().mockResolvedValue({}),
  deleteDevice: vi.fn().mockResolvedValue({}),
  bulkDeleteDevices: vi.fn().mockResolvedValue({}),
  bulkRebootDevices: vi.fn().mockResolvedValue({}),
  exportDevicesCSV: vi.fn().mockResolvedValue({ data: new Blob() }),
  apiErr: (e: any, fallback: string) => e?.response?.data?.error || fallback,
}))

vi.mock('element-plus', async (importOriginal) => {
  const actual = await importOriginal<typeof import('element-plus')>()
  return {
    ...actual,
    ElMessage: mockElMessage,
    ElMessageBox: { confirm: vi.fn().mockResolvedValue(true) },
  }
})

import Devices from '../views/Devices.vue'

const sampleDevices = [
  { id: 1, name: 'Router-A', mac: 'AA:BB:CC:DD:EE:01', ip_address: '192.168.1.1', model: 'X86-64', firmware: '23.05.5', status: 'online', group: 'office', cpu_usage: 25.5, mem_usage: 40.0 },
  { id: 2, name: 'Router-B', mac: 'AA:BB:CC:DD:EE:02', ip_address: '192.168.1.2', model: 'MT7621', firmware: '23.05.5', status: 'offline', group: 'branch', cpu_usage: 0, mem_usage: 0 },
  { id: 3, name: 'Gateway-C', mac: 'AA:BB:CC:DD:EE:03', ip_address: '10.0.0.1', model: 'IPQ8074', firmware: '23.05.4', status: 'online', group: 'office', cpu_usage: 60.2, mem_usage: 72.1 },
]

function createTestRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', redirect: '/devices' },
      { path: '/devices', name: 'Devices', component: Devices },
      { path: '/devices/:id', name: 'DeviceDetail', component: { template: '<div></div>' } },
    ],
  })
}

describe('Devices.vue', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
    mockGetDevices.mockResolvedValue({
      data: { data: sampleDevices, total: 3, page: 1, page_size: 50 },
    })
  })

  const mountDevices = async (role = 'admin') => {
    localStorage.setItem('token', 'test-token')
    localStorage.setItem('role', role)
    const router = createTestRouter()
    await router.push('/devices')
    await router.isReady()
    const wrapper = mount(Devices, { global: { plugins: [router] } })
    await flushPromises()
    return wrapper
  }

  it('fetches devices on mount', async () => {
    await mountDevices()
    expect(mockGetDevices).toHaveBeenCalled()
  })

  it('displays device names', async () => {
    const wrapper = await mountDevices()
    expect(wrapper.text()).toContain('Router-A')
    expect(wrapper.text()).toContain('Router-B')
    expect(wrapper.text()).toContain('Gateway-C')
  })

  it('displays MAC addresses', async () => {
    const wrapper = await mountDevices()
    expect(wrapper.text()).toContain('AA:BB:CC:DD:EE:01')
    expect(wrapper.text()).toContain('AA:BB:CC:DD:EE:02')
  })

  it('displays IP addresses', async () => {
    const wrapper = await mountDevices()
    expect(wrapper.text()).toContain('192.168.1.1')
    expect(wrapper.text()).toContain('10.0.0.1')
  })

  it('shows status filter buttons', async () => {
    const wrapper = await mountDevices()
    expect(wrapper.text()).toContain('全部')
    expect(wrapper.text()).toContain('在线')
    expect(wrapper.text()).toContain('离线')
  })

  it('shows CSV export button', async () => {
    const wrapper = await mountDevices()
    expect(wrapper.text()).toContain('导出 CSV')
  })

  it('shows bulk actions for admin', async () => {
    const wrapper = await mountDevices('admin')
    expect(wrapper.text()).toContain('批量重启')
    expect(wrapper.text()).toContain('批量删除')
  })

  it('shows bulk actions for operator', async () => {
    const wrapper = await mountDevices('operator')
    expect(wrapper.text()).toContain('批量重启')
    expect(wrapper.text()).toContain('批量删除')
  })

  it('hides bulk actions for viewer', async () => {
    const wrapper = await mountDevices('viewer')
    expect(wrapper.text()).not.toContain('批量重启')
    expect(wrapper.text()).not.toContain('批量删除')
  })

  it('handles API error gracefully', async () => {
    mockGetDevices.mockRejectedValueOnce({
      response: { data: { error: 'database error' } },
    })
    await mountDevices()
    expect(mockElMessage.error).toHaveBeenCalledWith('database error')
  })
})
