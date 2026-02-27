import { describe, it, expect, beforeEach } from 'vitest'
import { createRouter, createWebHistory, createMemoryHistory } from 'vue-router'

// Re-create the router logic for testing (the actual router module has side effects
// from dynamic imports, so we test the guard logic directly)

const roleLevel: Record<string, number> = { viewer: 1, operator: 2, admin: 3 }

function createTestRouter() {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/login', name: 'Login', component: { template: '<div>Login</div>' }, meta: { requiresAuth: false } },
      { path: '/', redirect: '/dashboard' },
      { path: '/dashboard', name: 'Dashboard', component: { template: '<div>Dashboard</div>' }, meta: { title: '仪表板' } },
      { path: '/devices', name: 'Devices', component: { template: '<div>Devices</div>' }, meta: { title: '设备管理' } },
      { path: '/users', name: 'Users', component: { template: '<div>Users</div>' }, meta: { title: '用户管理', requiresAdmin: true } },
      { path: '/audit', name: 'Audit', component: { template: '<div>Audit</div>' }, meta: { title: '审计日志', requiresAdmin: true } },
      { path: '/settings', name: 'Settings', component: { template: '<div>Settings</div>' }, meta: { title: '系统设置', minRole: 'operator' } },
      { path: '/:pathMatch(.*)*', name: 'NotFound', component: { template: '<div>404</div>' } },
    ],
  })

  router.beforeEach((to, _from, next) => {
    if (to.meta.requiresAuth === false) {
      next()
      return
    }
    const token = localStorage.getItem('token')
    if (!token) {
      next('/login')
      return
    }
    const userRole = localStorage.getItem('role') || 'viewer'
    if (to.meta.requiresAdmin && userRole !== 'admin') {
      next('/dashboard')
      return
    }
    const minRole = to.meta.minRole as string | undefined
    if (minRole && (roleLevel[userRole] || 0) < (roleLevel[minRole] || 0)) {
      next('/dashboard')
      return
    }
    next()
  })

  return router
}

describe('Router Guards', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('allows access to login page without token', async () => {
    const router = createTestRouter()
    await router.push('/login')
    expect(router.currentRoute.value.name).toBe('Login')
  })

  it('redirects to login when no token', async () => {
    const router = createTestRouter()
    await router.push('/dashboard')
    expect(router.currentRoute.value.path).toBe('/login')
  })

  it('allows authenticated user to access dashboard', async () => {
    localStorage.setItem('token', 'fake-token')
    localStorage.setItem('role', 'viewer')
    const router = createTestRouter()
    await router.push('/dashboard')
    expect(router.currentRoute.value.name).toBe('Dashboard')
  })

  it('allows authenticated user to access devices', async () => {
    localStorage.setItem('token', 'fake-token')
    localStorage.setItem('role', 'operator')
    const router = createTestRouter()
    await router.push('/devices')
    expect(router.currentRoute.value.name).toBe('Devices')
  })

  it('redirects non-admin from admin-only routes', async () => {
    localStorage.setItem('token', 'fake-token')
    localStorage.setItem('role', 'viewer')
    const router = createTestRouter()
    await router.push('/users')
    expect(router.currentRoute.value.path).toBe('/dashboard')
  })

  it('allows admin to access admin routes', async () => {
    localStorage.setItem('token', 'fake-token')
    localStorage.setItem('role', 'admin')
    const router = createTestRouter()
    await router.push('/users')
    expect(router.currentRoute.value.name).toBe('Users')
  })

  it('allows admin to access audit logs', async () => {
    localStorage.setItem('token', 'fake-token')
    localStorage.setItem('role', 'admin')
    const router = createTestRouter()
    await router.push('/audit')
    expect(router.currentRoute.value.name).toBe('Audit')
  })

  it('redirects viewer from operator-only routes (minRole)', async () => {
    localStorage.setItem('token', 'fake-token')
    localStorage.setItem('role', 'viewer')
    const router = createTestRouter()
    await router.push('/settings')
    expect(router.currentRoute.value.path).toBe('/dashboard')
  })

  it('allows operator to access settings (minRole: operator)', async () => {
    localStorage.setItem('token', 'fake-token')
    localStorage.setItem('role', 'operator')
    const router = createTestRouter()
    await router.push('/settings')
    expect(router.currentRoute.value.name).toBe('Settings')
  })

  it('allows admin to access settings (minRole: operator)', async () => {
    localStorage.setItem('token', 'fake-token')
    localStorage.setItem('role', 'admin')
    const router = createTestRouter()
    await router.push('/settings')
    expect(router.currentRoute.value.name).toBe('Settings')
  })

  it('redirects / to /dashboard', async () => {
    localStorage.setItem('token', 'fake-token')
    const router = createTestRouter()
    await router.push('/')
    expect(router.currentRoute.value.path).toBe('/dashboard')
  })

  it('handles 404 routes', async () => {
    localStorage.setItem('token', 'fake-token')
    const router = createTestRouter()
    await router.push('/non-existent-page')
    expect(router.currentRoute.value.name).toBe('NotFound')
  })

  it('defaults to viewer role when role not set', async () => {
    localStorage.setItem('token', 'fake-token')
    // No role set — should default to viewer
    const router = createTestRouter()
    await router.push('/users')
    // viewer cannot access admin-only route
    expect(router.currentRoute.value.path).toBe('/dashboard')
  })
})
