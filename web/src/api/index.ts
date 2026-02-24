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

export default api
