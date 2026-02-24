import axios from 'axios'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 15000,
})

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
      localStorage.removeItem('token')
      window.location.href = '/login'
    }
    return Promise.reject(err)
  }
)

// Auth
export const login = (username: string, password: string) =>
  api.post('/auth/login', { username, password })

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

export const getDeviceMetrics = (id: number) =>
  api.get(`/devices/${id}/metrics`)

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
export const deleteUser = (id: number) => api.delete(`/users/${id}`)

// Audit
export const getAuditLogs = () => api.get('/audit-logs')

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
export const deleteWANInterface = (id: number) => api.delete(`/network/wan/${id}`)
export const getMWANPolicies = (deviceId?: number) =>
  api.get('/network/mwan/policies', { params: deviceId ? { device_id: deviceId } : {} })
export const createMWANPolicy = (data: any) => api.post('/network/mwan/policies', data)
export const deleteMWANPolicy = (id: number) => api.delete(`/network/mwan/policies/${id}`)
export const getMWANRules = (deviceId?: number) =>
  api.get('/network/mwan/rules', { params: deviceId ? { device_id: deviceId } : {} })
export const createMWANRule = (data: any) => api.post('/network/mwan/rules', data)
export const deleteMWANRule = (id: number) => api.delete(`/network/mwan/rules/${id}`)
export const applyMWAN = (deviceId: number) => api.post(`/network/mwan/apply/${deviceId}`)

// DHCP
export const getDHCPPools = (deviceId?: number) =>
  api.get('/network/dhcp/pools', { params: deviceId ? { device_id: deviceId } : {} })
export const createDHCPPool = (data: any) => api.post('/network/dhcp/pools', data)
export const deleteDHCPPool = (id: number) => api.delete(`/network/dhcp/pools/${id}`)
export const getStaticLeases = (deviceId?: number) =>
  api.get('/network/dhcp/leases', { params: deviceId ? { device_id: deviceId } : {} })
export const createStaticLease = (data: any) => api.post('/network/dhcp/leases', data)
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

export default api
