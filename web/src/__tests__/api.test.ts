import { describe, it, expect, vi, beforeEach } from 'vitest'
import axios from 'axios'

// Mock element-plus to avoid CSS import issues
vi.mock('element-plus', () => ({
  ElMessage: { error: vi.fn(), success: vi.fn(), warning: vi.fn() },
}))

// We test the exported utility functions and API endpoint wiring
// by importing them after mocking axios.
vi.mock('axios', () => {
  const instance = {
    get: vi.fn().mockResolvedValue({ data: {} }),
    post: vi.fn().mockResolvedValue({ data: {} }),
    put: vi.fn().mockResolvedValue({ data: {} }),
    delete: vi.fn().mockResolvedValue({ data: {} }),
    interceptors: {
      request: { use: vi.fn() },
      response: { use: vi.fn() },
    },
  }
  return {
    default: { create: vi.fn(() => instance) },
  }
})

// Import after mock setup
const apiModule = await import('@/api/index')

// Get the mocked instance
const mockAxios = (axios.create as any)()

describe('API Client', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('apiErr()', () => {
    it('extracts error message from axios error response', () => {
      const err = { response: { data: { error: 'invalid credentials' } } }
      expect(apiModule.apiErr(err, 'fallback')).toBe('invalid credentials')
    })

    it('returns fallback when no response data', () => {
      expect(apiModule.apiErr({}, 'fallback msg')).toBe('fallback msg')
    })

    it('returns fallback for null error', () => {
      expect(apiModule.apiErr(null, 'default')).toBe('default')
    })

    it('returns fallback when error field is missing', () => {
      const err = { response: { data: {} } }
      expect(apiModule.apiErr(err, 'default')).toBe('default')
    })
  })

  describe('Device API', () => {
    it('getDevices calls GET /devices with params', async () => {
      await apiModule.getDevices({ status: 'online' })
      expect(mockAxios.get).toHaveBeenCalledWith('/devices', { params: { status: 'online' } })
    })

    it('getDevice calls GET /devices/:id', async () => {
      await apiModule.getDevice(42)
      expect(mockAxios.get).toHaveBeenCalledWith('/devices/42')
    })

    it('updateDevice calls PUT /devices/:id', async () => {
      await apiModule.updateDevice(1, { name: 'router-1' })
      expect(mockAxios.put).toHaveBeenCalledWith('/devices/1', { name: 'router-1' })
    })

    it('deleteDevice calls DELETE /devices/:id', async () => {
      await apiModule.deleteDevice(5)
      expect(mockAxios.delete).toHaveBeenCalledWith('/devices/5')
    })

    it('rebootDevice calls POST /devices/:id/reboot', async () => {
      await apiModule.rebootDevice(3)
      expect(mockAxios.post).toHaveBeenCalledWith('/devices/3/reboot')
    })

    it('bulkDeleteDevices calls POST with ids', async () => {
      await apiModule.bulkDeleteDevices([1, 2, 3])
      expect(mockAxios.post).toHaveBeenCalledWith('/devices/bulk/delete', { ids: [1, 2, 3] })
    })

    it('bulkRebootDevices calls POST with ids', async () => {
      await apiModule.bulkRebootDevices([4, 5])
      expect(mockAxios.post).toHaveBeenCalledWith('/devices/bulk/reboot', { ids: [4, 5] })
    })

    it('exportDevicesCSV requests blob response', async () => {
      await apiModule.exportDevicesCSV({ status: 'online' })
      expect(mockAxios.get).toHaveBeenCalledWith('/devices/export', {
        params: { status: 'online' },
        responseType: 'blob',
      })
    })
  })

  describe('Auth API', () => {
    it('login calls POST /auth/login', async () => {
      mockAxios.post.mockResolvedValueOnce({ data: { token: 'abc', user: { role: 'admin' } } })
      await apiModule.login('admin', 'Pass1234')
      expect(mockAxios.post).toHaveBeenCalledWith('/auth/login', { username: 'admin', password: 'Pass1234' })
    })

    it('changePassword calls PUT /auth/password', async () => {
      await apiModule.changePassword('OldPass1', 'NewPass2')
      expect(mockAxios.put).toHaveBeenCalledWith('/auth/password', {
        old_password: 'OldPass1',
        new_password: 'NewPass2',
      })
    })

    it('getMe calls GET /auth/me', async () => {
      await apiModule.getMe()
      expect(mockAxios.get).toHaveBeenCalledWith('/auth/me')
    })
  })

  describe('Config API', () => {
    it('getTemplates with category', async () => {
      await apiModule.getTemplates('network')
      expect(mockAxios.get).toHaveBeenCalledWith('/templates', { params: { category: 'network' } })
    })

    it('getTemplates without category', async () => {
      await apiModule.getTemplates()
      expect(mockAxios.get).toHaveBeenCalledWith('/templates', { params: {} })
    })

    it('pushConfig sends template_id', async () => {
      await apiModule.pushConfig(1, { template_id: 5 })
      expect(mockAxios.post).toHaveBeenCalledWith('/devices/1/config/push', { template_id: 5 })
    })
  })

  describe('Firewall API', () => {
    it('getFirewallZones with device filter', async () => {
      await apiModule.getFirewallZones(10)
      expect(mockAxios.get).toHaveBeenCalledWith('/firewall/zones', { params: { device_id: 10 } })
    })

    it('getFirewallZones without filter', async () => {
      await apiModule.getFirewallZones()
      expect(mockAxios.get).toHaveBeenCalledWith('/firewall/zones', { params: {} })
    })

    it('applyFirewall calls POST', async () => {
      await apiModule.applyFirewall(7)
      expect(mockAxios.post).toHaveBeenCalledWith('/firewall/apply/7')
    })
  })

  describe('User API', () => {
    it('getUsers calls GET /users', async () => {
      await apiModule.getUsers()
      expect(mockAxios.get).toHaveBeenCalledWith('/users')
    })

    it('createUser calls POST /users', async () => {
      await apiModule.createUser({ username: 'test', password: 'Pass1234', role: 'viewer' })
      expect(mockAxios.post).toHaveBeenCalledWith('/users', { username: 'test', password: 'Pass1234', role: 'viewer' })
    })

    it('deleteUser calls DELETE /users/:id', async () => {
      await apiModule.deleteUser(3)
      expect(mockAxios.delete).toHaveBeenCalledWith('/users/3')
    })
  })

  describe('Settings API', () => {
    it('batchUpsertSettings sends array', async () => {
      const items = [{ key: 'k1', value: 'v1' }]
      await apiModule.batchUpsertSettings(items)
      expect(mockAxios.post).toHaveBeenCalledWith('/settings/batch', items)
    })
  })

  describe('Alert API', () => {
    it('resolveAlert calls POST /alerts/:id/resolve', async () => {
      await apiModule.resolveAlert(99)
      expect(mockAxios.post).toHaveBeenCalledWith('/alerts/99/resolve')
    })
  })

  describe('clearTokenRefresh()', () => {
    it('is a callable function', () => {
      expect(typeof apiModule.clearTokenRefresh).toBe('function')
      // Should not throw
      apiModule.clearTokenRefresh()
    })
  })
})
