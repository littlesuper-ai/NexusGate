import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'

// Hoisted mocks — available before vi.mock factory runs
const { mockLogin, mockElMessage } = vi.hoisted(() => ({
  mockLogin: vi.fn(),
  mockElMessage: { error: vi.fn(), warning: vi.fn(), success: vi.fn(), info: vi.fn() },
}))

vi.mock('../api', () => ({
  login: (...args: any[]) => mockLogin(...args),
  apiErr: (e: any, fallback: string) => e?.response?.data?.error || fallback,
}))

vi.mock('element-plus', async (importOriginal) => {
  const actual = await importOriginal<typeof import('element-plus')>()
  return { ...actual, ElMessage: mockElMessage }
})

import Login from '../views/Login.vue'

function createTestRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', redirect: '/login' },
      { path: '/login', name: 'Login', component: Login },
      { path: '/dashboard', name: 'Dashboard', component: { template: '<div>Dashboard</div>' } },
    ],
  })
}

describe('Login.vue', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
  })

  const mountLogin = async () => {
    const router = createTestRouter()
    await router.push('/login')
    await router.isReady()
    return mount(Login, { global: { plugins: [router] } })
  }

  it('renders login form with title', async () => {
    const wrapper = await mountLogin()
    expect(wrapper.text()).toContain('NexusGate')
    expect(wrapper.text()).toContain('企业级路由网关管理平台')
  })

  it('shows warning when fields are empty', async () => {
    const wrapper = await mountLogin()

    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(mockElMessage.warning).toHaveBeenCalledWith('请输入用户名和密码')
    expect(mockLogin).not.toHaveBeenCalled()
  })

  it('calls login API with credentials', async () => {
    mockLogin.mockResolvedValueOnce({
      data: { token: 'jwt-token', user: { username: 'admin', role: 'admin' } },
    })

    const wrapper = await mountLogin()
    const inputs = wrapper.findAll('input')
    await inputs[0].setValue('admin')
    await inputs[1].setValue('Pass1234')

    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(mockLogin).toHaveBeenCalledWith('admin', 'Pass1234')
  })

  it('stores token and role on successful login', async () => {
    mockLogin.mockResolvedValueOnce({
      data: { token: 'jwt-token-123', user: { username: 'admin', role: 'admin' } },
    })

    const wrapper = await mountLogin()
    const inputs = wrapper.findAll('input')
    await inputs[0].setValue('admin')
    await inputs[1].setValue('Pass1234')

    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(localStorage.getItem('token')).toBe('jwt-token-123')
    expect(localStorage.getItem('role')).toBe('admin')
    expect(localStorage.getItem('username')).toBe('admin')
  })

  it('shows error on login failure', async () => {
    mockLogin.mockRejectedValueOnce({
      response: { data: { error: 'invalid credentials' } },
    })

    const wrapper = await mountLogin()
    const inputs = wrapper.findAll('input')
    await inputs[0].setValue('admin')
    await inputs[1].setValue('wrongpass')

    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(mockElMessage.error).toHaveBeenCalledWith('invalid credentials')
  })

  it('shows fallback error message on network error', async () => {
    mockLogin.mockRejectedValueOnce(new Error('network error'))

    const wrapper = await mountLogin()
    const inputs = wrapper.findAll('input')
    await inputs[0].setValue('admin')
    await inputs[1].setValue('password')

    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(mockElMessage.error).toHaveBeenCalledWith('登录失败，请检查用户名和密码')
  })
})
