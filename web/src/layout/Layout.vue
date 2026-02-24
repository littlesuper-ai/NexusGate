<template>
  <el-container class="layout">
    <el-aside width="220px" class="sidebar">
      <div class="logo">
        <h2>NexusGate</h2>
      </div>
      <el-menu
        :default-active="route.path"
        router
        background-color="#001529"
        text-color="#ffffffa6"
        active-text-color="#fff"
      >
        <el-menu-item index="/dashboard">
          <el-icon><Monitor /></el-icon>
          <span>仪表板</span>
        </el-menu-item>
        <el-menu-item index="/devices">
          <el-icon><Connection /></el-icon>
          <span>设备管理</span>
        </el-menu-item>
        <el-menu-item index="/templates">
          <el-icon><Document /></el-icon>
          <span>配置模板</span>
        </el-menu-item>
        <el-sub-menu index="network">
          <template #title>
            <el-icon><Setting /></el-icon>
            <span>网络管理</span>
          </template>
          <el-menu-item index="/firewall">防火墙</el-menu-item>
          <el-menu-item index="/vpn">VPN</el-menu-item>
          <el-menu-item index="/mwan">多线负载</el-menu-item>
          <el-menu-item index="/dhcp">DHCP</el-menu-item>
          <el-menu-item index="/vlan">VLAN</el-menu-item>
          <el-menu-item index="/topology">网络拓扑</el-menu-item>
        </el-sub-menu>
        <el-menu-item index="/firmware">
          <el-icon><Upload /></el-icon>
          <span>固件管理</span>
        </el-menu-item>
        <el-menu-item index="/monitoring">
          <el-icon><DataLine /></el-icon>
          <span>监控中心</span>
        </el-menu-item>
        <el-menu-item index="/users">
          <el-icon><User /></el-icon>
          <span>用户管理</span>
        </el-menu-item>
        <el-menu-item index="/audit">
          <el-icon><List /></el-icon>
          <span>审计日志</span>
        </el-menu-item>
        <el-menu-item index="/settings">
          <el-icon><Setting /></el-icon>
          <span>系统设置</span>
        </el-menu-item>
      </el-menu>
    </el-aside>

    <el-container>
      <el-header class="header">
        <span class="page-title">{{ route.meta.title }}</span>
        <el-dropdown @command="handleCommand">
          <span class="user-info">
            {{ username }} <el-icon><ArrowDown /></el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="logout">退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </el-header>
      <el-main>
        <slot />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  Monitor, Connection, Document, Setting,
  DataLine, User, List, ArrowDown, Upload,
} from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const username = computed(() => localStorage.getItem('username') || 'Admin')

const handleCommand = (cmd: string) => {
  if (cmd === 'logout') {
    localStorage.removeItem('token')
    localStorage.removeItem('username')
    router.push('/login')
  }
}
</script>

<style scoped>
.layout {
  height: 100vh;
}
.sidebar {
  background: #001529;
  overflow-y: auto;
}
.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
}
.logo h2 {
  margin: 0;
  font-size: 20px;
}
.header {
  background: #fff;
  display: flex;
  align-items: center;
  justify-content: space-between;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);
}
.page-title {
  font-size: 16px;
  font-weight: 600;
}
.user-info {
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 4px;
}
</style>
