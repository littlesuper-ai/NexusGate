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
        <el-menu-item index="/alerts">
          <el-icon><Bell /></el-icon>
          <span>告警中心</span>
        </el-menu-item>
        <el-menu-item v-if="isAdmin" index="/users">
          <el-icon><User /></el-icon>
          <span>用户管理</span>
        </el-menu-item>
        <el-menu-item v-if="isAdmin" index="/audit">
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
            {{ username }}
            <el-tag size="small" :type="role === 'admin' ? 'danger' : role === 'operator' ? '' : 'info'" style="margin-left: 6px">
              {{ role }}
            </el-tag>
            <el-icon><ArrowDown /></el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="password">修改密码</el-dropdown-item>
              <el-dropdown-item command="logout" divided>退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </el-header>
      <el-main>
        <slot />
      </el-main>
    </el-container>

    <el-dialog v-model="showPasswordDialog" title="修改密码" width="400">
      <el-form :model="pwForm" label-width="80px">
        <el-form-item label="旧密码">
          <el-input v-model="pwForm.oldPassword" type="password" show-password />
        </el-form-item>
        <el-form-item label="新密码">
          <el-input v-model="pwForm.newPassword" type="password" show-password />
        </el-form-item>
        <el-form-item label="确认密码">
          <el-input v-model="pwForm.confirmPassword" type="password" show-password />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showPasswordDialog = false">取消</el-button>
        <el-button type="primary" @click="handleChangePassword">确认</el-button>
      </template>
    </el-dialog>
  </el-container>
</template>

<script setup lang="ts">
import { computed, ref, reactive } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  Monitor, Connection, Document, Setting,
  DataLine, User, List, ArrowDown, Upload, Bell,
} from '@element-plus/icons-vue'
import { changePassword, clearTokenRefresh } from '../api'

const route = useRoute()
const router = useRouter()
const username = computed(() => localStorage.getItem('username') || 'Admin')
const role = computed(() => localStorage.getItem('role') || 'viewer')
const isAdmin = computed(() => role.value === 'admin')

const showPasswordDialog = ref(false)
const pwForm = reactive({ oldPassword: '', newPassword: '', confirmPassword: '' })

const handleCommand = (cmd: string) => {
  if (cmd === 'logout') {
    clearTokenRefresh()
    localStorage.removeItem('token')
    localStorage.removeItem('username')
    localStorage.removeItem('role')
    router.push('/login')
  } else if (cmd === 'password') {
    pwForm.oldPassword = ''
    pwForm.newPassword = ''
    pwForm.confirmPassword = ''
    showPasswordDialog.value = true
  }
}

const handleChangePassword = async () => {
  if (!pwForm.oldPassword || !pwForm.newPassword) {
    ElMessage.warning('请填写所有字段')
    return
  }
  if (pwForm.newPassword.length < 8) {
    ElMessage.warning('新密码至少8个字符')
    return
  }
  if (pwForm.newPassword !== pwForm.confirmPassword) {
    ElMessage.warning('两次输入的密码不一致')
    return
  }
  try {
    await changePassword(pwForm.oldPassword, pwForm.newPassword)
    ElMessage.success('密码已修改')
    showPasswordDialog.value = false
  } catch (err: any) {
    ElMessage.error(err.response?.data?.error || '修改失败')
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
