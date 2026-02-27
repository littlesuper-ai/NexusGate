import axios from 'axios'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 15000,
})

// Parse JWT expiration without external library
function getTokenExp(token: string): number | null {
  try {
    const payload = JSON.parse(atob(token.split('.')[1]))
    return payload.exp || null
  } catch { return null }
}

// Proactive token refresh: refresh when <20% lifetime remains
let refreshTimer: ReturnType<typeof setTimeout> | null = null

function scheduleTokenRefresh() {
  if (refreshTimer) clearTimeout(refreshTimer)
  const token = localStorage.getItem('token')
  if (!token) return
  const exp = getTokenExp(token)
  if (!exp) return
  const now = Math.floor(Date.now() / 1000)
  const remaining = exp - now
  // Refresh when 80% of lifetime has passed (at 20% remaining)
  const refreshAt = Math.max(remaining * 0.8, 60) * 1000
  refreshTimer = setTimeout(async () => {
    try {
      const { data } = await api.post('/auth/refresh')
      localStorage.setItem('token', data.token)
      if (data.user) {
        localStorage.setItem('role', data.user.role)
      }
      scheduleTokenRefresh()
    } catch {
      // Refresh failed — token may have expired, let 401 interceptor handle it
    }
  }, refreshAt)
}

// Start refresh timer on load if token exists
scheduleTokenRefresh()

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      if (refreshTimer) clearTimeout(refreshTimer)
      localStorage.removeItem('token')
      window.location.href = '/login'
    } else if (err.response?.status === 403) {
      import('element-plus').then(({ ElMessage }) => {
        ElMessage.error('权限不足，无法执行此操作')
      })
    }
    return Promise.reject(err)
  }
)

// Auth
export const login = (username: string, password: string) =>
  api.post('/auth/login', { username, password }).then((res) => {
    // Schedule refresh after successful login
    if (res.data.token) {
      localStorage.setItem('token', res.data.token)
      scheduleTokenRefresh()
    }
    return res
  })

export const refreshToken = () => api.post('/auth/refresh')

export const getMe = () => api.get('/auth/me')

export const changePassword = (oldPassword: string, newPassword: string) =>
  api.put('/auth/password', { old_password: oldPassword, new_password: newPassword })

// Devices
export const getDevices = (params?: Record<string, string>) =>
  api.get('/devices', { params })

export const getDevice = (id: number) =>
  api.get(`/devices/${id}`)

export const updateDevice = (id: number, data: Record<string, string>) =>
  api.put(`/devices/${id}`, data)

export const deleteDevice = (id: number) =>
  api.delete(`/devices/${id}`)

export const rebootDevice = (id: number) =>
  api.post(`/devices/${id}/reboot`)

export const bulkDeleteDevices = (ids: number[]) =>
  api.post('/devices/bulk/delete', { ids })

export const bulkRebootDevices = (ids: number[]) =>
  api.post('/devices/bulk/reboot', { ids })

export const getDeviceMetrics = (id: number) =>
  api.get(`/devices/${id}/metrics`)

export const exportDevicesCSV = (params?: Record<string, string>) =>
  api.get('/devices/export', { params, responseType: 'blob' })

// Config
export const getTemplates = (category?: string) =>
  api.get('/templates', { params: category ? { category } : {} })

export const createTemplate = (data: Record<string, string>) =>
  api.post('/templates', data)

export const updateTemplate = (id: number, data: Record<string, string>) =>
  api.put(`/templates/${id}`, data)

export const deleteTemplate = (id: number) =>
  api.delete(`/templates/${id}`)

export const pushConfig = (deviceId: number, data: { template_id?: number; content?: string }) =>
  api.post(`/devices/${deviceId}/config/push`, data)

export const getConfigHistory = (deviceId: number) =>
  api.get(`/devices/${deviceId}/config/history`)

// Dashboard
export const getDashboardSummary = () =>
  api.get('/dashboard/summary')

// Users
export const getUsers = () => api.get('/users')
export const createUser = (data: { username: string; password: string; role: string; email?: string }) =>
  api.post('/users', data)
export const updateUser = (id: number, data: { role?: string; email?: string; password?: string }) =>
  api.put(`/users/${id}`, data)
export const deleteUser = (id: number) => api.delete(`/users/${id}`)

// Audit
export const getAuditLogs = (params?: Record<string, string | number>) => api.get('/audit-logs', { params })

// Firewall
export const getFirewallZones = (deviceId?: number) =>
  api.get('/firewall/zones', { params: deviceId ? { device_id: deviceId } : {} })
export const createFirewallZone = (data: any) => api.post('/firewall/zones', data)
export const updateFirewallZone = (id: number, data: any) => api.put(`/firewall/zones/${id}`, data)
export const deleteFirewallZone = (id: number) => api.delete(`/firewall/zones/${id}`)
export const getFirewallRules = (deviceId?: number) =>
  api.get('/firewall/rules', { params: deviceId ? { device_id: deviceId } : {} })
export const createFirewallRule = (data: any) => api.post('/firewall/rules', data)
export const updateFirewallRule = (id: number, data: any) => api.put(`/firewall/rules/${id}`, data)
export const deleteFirewallRule = (id: number) => api.delete(`/firewall/rules/${id}`)
export const applyFirewall = (deviceId: number) => api.post(`/firewall/apply/${deviceId}`)

// VPN
export const getVPNInterfaces = (deviceId?: number) =>
  api.get('/vpn/interfaces', { params: deviceId ? { device_id: deviceId } : {} })
export const createVPNInterface = (data: any) => api.post('/vpn/interfaces', data)
export const updateVPNInterface = (id: number, data: any) => api.put(`/vpn/interfaces/${id}`, data)
export const deleteVPNInterface = (id: number) => api.delete(`/vpn/interfaces/${id}`)
export const getVPNPeers = (interfaceId?: number) =>
  api.get('/vpn/peers', { params: interfaceId ? { interface_id: interfaceId } : {} })
export const createVPNPeer = (data: any) => api.post('/vpn/peers', data)
export const updateVPNPeer = (id: number, data: any) => api.put(`/vpn/peers/${id}`, data)
export const deleteVPNPeer = (id: number) => api.delete(`/vpn/peers/${id}`)
export const applyVPN = (deviceId: number) => api.post(`/vpn/apply/${deviceId}`)

// Firmware
export const getFirmwares = (target?: string) =>
  api.get('/firmware', { params: target ? { target } : {} })
export const uploadFirmware = (formData: FormData) =>
  api.post('/firmware/upload', formData, { headers: { 'Content-Type': 'multipart/form-data' } })
export const deleteFirmware = (id: number) => api.delete(`/firmware/${id}`)
export const markFirmwareStable = (id: number) => api.post(`/firmware/${id}/stable`)
export const pushFirmwareUpgrade = (data: { device_id: number; firmware_id: number }) =>
  api.post('/firmware/upgrade', data)
export const batchFirmwareUpgrade = (data: { firmware_id: number; group?: string; model?: string }) =>
  api.post('/firmware/upgrade/batch', data)
export const getUpgradeHistory = (deviceId?: number) =>
  api.get('/firmware/upgrades', { params: deviceId ? { device_id: deviceId } : {} })

// Multi-WAN
export const getWANInterfaces = (deviceId?: number) =>
  api.get('/network/wan', { params: deviceId ? { device_id: deviceId } : {} })
export const createWANInterface = (data: any) => api.post('/network/wan', data)
export const updateWANInterface = (id: number, data: any) => api.put(`/network/wan/${id}`, data)
export const deleteWANInterface = (id: number) => api.delete(`/network/wan/${id}`)
export const getMWANPolicies = (deviceId?: number) =>
  api.get('/network/mwan/policies', { params: deviceId ? { device_id: deviceId } : {} })
export const createMWANPolicy = (data: any) => api.post('/network/mwan/policies', data)
export const updateMWANPolicy = (id: number, data: any) => api.put(`/network/mwan/policies/${id}`, data)
export const deleteMWANPolicy = (id: number) => api.delete(`/network/mwan/policies/${id}`)
export const getMWANRules = (deviceId?: number) =>
  api.get('/network/mwan/rules', { params: deviceId ? { device_id: deviceId } : {} })
export const createMWANRule = (data: any) => api.post('/network/mwan/rules', data)
export const updateMWANRule = (id: number, data: any) => api.put(`/network/mwan/rules/${id}`, data)
export const deleteMWANRule = (id: number) => api.delete(`/network/mwan/rules/${id}`)
export const applyMWAN = (deviceId: number) => api.post(`/network/mwan/apply/${deviceId}`)

// DHCP
export const getDHCPPools = (deviceId?: number) =>
  api.get('/network/dhcp/pools', { params: deviceId ? { device_id: deviceId } : {} })
export const createDHCPPool = (data: any) => api.post('/network/dhcp/pools', data)
export const updateDHCPPool = (id: number, data: any) => api.put(`/network/dhcp/pools/${id}`, data)
export const deleteDHCPPool = (id: number) => api.delete(`/network/dhcp/pools/${id}`)
export const getStaticLeases = (deviceId?: number) =>
  api.get('/network/dhcp/leases', { params: deviceId ? { device_id: deviceId } : {} })
export const createStaticLease = (data: any) => api.post('/network/dhcp/leases', data)
export const updateStaticLease = (id: number, data: any) => api.put(`/network/dhcp/leases/${id}`, data)
export const deleteStaticLease = (id: number) => api.delete(`/network/dhcp/leases/${id}`)
export const applyDHCP = (deviceId: number) => api.post(`/network/dhcp/apply/${deviceId}`)

// VLAN
export const getVLANs = (deviceId?: number) =>
  api.get('/network/vlans', { params: deviceId ? { device_id: deviceId } : {} })
export const createVLAN = (data: any) => api.post('/network/vlans', data)
export const updateVLAN = (id: number, data: any) => api.put(`/network/vlans/${id}`, data)
export const deleteVLAN = (id: number) => api.delete(`/network/vlans/${id}`)
export const applyVLAN = (deviceId: number) => api.post(`/network/vlans/apply/${deviceId}`)

// Settings
export const getSettings = (category?: string) =>
  api.get('/settings', { params: category ? { category } : {} })
export const getSetting = (key: string) => api.get(`/settings/${key}`)
export const upsertSetting = (data: { key: string; value: string; category?: string }) =>
  api.post('/settings', data)
export const batchUpsertSettings = (items: { key: string; value: string; category?: string }[]) =>
  api.post('/settings/batch', items)
export const deleteSetting = (key: string) => api.delete(`/settings/${key}`)

// Alerts
export const getAlerts = (params?: Record<string, string>) =>
  api.get('/alerts', { params })
export const getAlertSummary = () => api.get('/alerts/summary')
export const resolveAlert = (id: number) => api.post(`/alerts/${id}/resolve`)

/** Clear the proactive token refresh timer (call on explicit logout). */
export function clearTokenRefresh() {
  if (refreshTimer) { clearTimeout(refreshTimer); refreshTimer = null }
}

/** Extract error message from Axios error response, with fallback. */
export function apiErr(e: any, fallback: string): string {
  return e?.response?.data?.error || fallback
}

export default api
