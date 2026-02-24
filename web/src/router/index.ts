import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('../views/Login.vue'),
      meta: { requiresAuth: false },
    },
    {
      path: '/',
      redirect: '/dashboard',
    },
    {
      path: '/dashboard',
      name: 'Dashboard',
      component: () => import('../views/Dashboard.vue'),
      meta: { title: '仪表板' },
    },
    {
      path: '/devices',
      name: 'Devices',
      component: () => import('../views/Devices.vue'),
      meta: { title: '设备管理' },
    },
    {
      path: '/devices/:id',
      name: 'DeviceDetail',
      component: () => import('../views/DeviceDetail.vue'),
      meta: { title: '设备详情' },
    },
    {
      path: '/templates',
      name: 'Templates',
      component: () => import('../views/Templates.vue'),
      meta: { title: '配置模板' },
    },
    {
      path: '/vpn',
      name: 'VPN',
      component: () => import('../views/VPN.vue'),
      meta: { title: 'VPN 管理' },
    },
    {
      path: '/firewall',
      name: 'Firewall',
      component: () => import('../views/Firewall.vue'),
      meta: { title: '防火墙' },
    },
    {
      path: '/monitoring',
      name: 'Monitoring',
      component: () => import('../views/Monitoring.vue'),
      meta: { title: '监控中心' },
    },
    {
      path: '/users',
      name: 'Users',
      component: () => import('../views/Users.vue'),
      meta: { title: '用户管理' },
    },
    {
      path: '/audit',
      name: 'Audit',
      component: () => import('../views/Audit.vue'),
      meta: { title: '审计日志' },
    },
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
  next()
})

export default router
