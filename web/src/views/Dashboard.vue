<template>
  <div>
    <el-row :gutter="16" class="stat-cards">
      <el-col :span="4">
        <el-card shadow="hover">
          <el-statistic title="设备总数" :value="summary.total_devices" />
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card shadow="hover">
          <el-statistic title="在线设备" :value="summary.online_devices">
            <template #suffix><span style="color: #67c23a; font-size: 14px">台</span></template>
          </el-statistic>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card shadow="hover">
          <el-statistic title="离线设备" :value="summary.offline_devices">
            <template #suffix><span style="color: #f56c6c; font-size: 14px">台</span></template>
          </el-statistic>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card shadow="hover">
          <el-statistic title="未知设备" :value="summary.unknown_devices">
            <template #suffix><span style="color: #909399; font-size: 14px">台</span></template>
          </el-statistic>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card shadow="hover" :body-style="{ cursor: 'pointer' }" @click="$router.push('/alerts')">
          <el-statistic title="未处理告警" :value="alertSummary.unresolved">
            <template #suffix>
              <span v-if="alertSummary.critical > 0" style="color: #f56c6c; font-size: 12px">
                ({{ alertSummary.critical }} 严重)
              </span>
            </template>
          </el-statistic>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card shadow="hover">
          <el-statistic title="配置模板" :value="templateCount" />
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="16">
        <el-card>
          <template #header>
            <el-row justify="space-between" align="middle">
              <span>设备列表</span>
              <el-tag :type="wsConnected ? 'success' : 'info'" size="small">
                {{ wsConnected ? '实时更新' : '离线' }}
              </el-tag>
            </el-row>
          </template>
          <el-table :data="devices" stripe size="small" max-height="420">
            <el-table-column prop="name" label="名称" />
            <el-table-column prop="ip_address" label="IP" width="130" />
            <el-table-column prop="group" label="分组" width="100" />
            <el-table-column prop="status" label="状态" width="80">
              <template #default="{ row }">
                <el-tag :type="row.status === 'online' ? 'success' : row.status === 'offline' ? 'danger' : 'info'" size="small">
                  {{ row.status === 'online' ? '在线' : row.status === 'offline' ? '离线' : '未知' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="CPU" width="70">
              <template #default="{ row }">{{ row.cpu_usage ? row.cpu_usage.toFixed(0) + '%' : '-' }}</template>
            </el-table-column>
            <el-table-column label="内存" width="70">
              <template #default="{ row }">{{ row.mem_usage ? row.mem_usage.toFixed(0) + '%' : '-' }}</template>
            </el-table-column>
            <el-table-column prop="firmware" label="固件" width="90" />
          </el-table>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card>
          <template #header>最近操作</template>
          <el-timeline>
            <el-timeline-item
              v-for="log in auditLogs"
              :key="log.id"
              :timestamp="formatTime(log.created_at)"
              placement="top"
              :type="logType(log.action)"
              size="small"
            >
              <strong>{{ log.username }}</strong> {{ log.action }} {{ log.resource }}
              <div v-if="log.detail" style="color: #999; font-size: 12px">{{ log.detail }}</div>
            </el-timeline-item>
          </el-timeline>
          <el-empty v-if="auditLogs.length === 0" description="暂无操作记录" />
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getDashboardSummary, getDevices, getTemplates, getAuditLogs, getAlertSummary } from '../api'
import { useWebSocket } from '../composables/useWebSocket'

const summary = ref({ total_devices: 0, online_devices: 0, offline_devices: 0, unknown_devices: 0 })
const alertSummary = ref({ total: 0, unresolved: 0, warning: 0, critical: 0 })
const devices = ref<any[]>([])
const templateCount = ref(0)
const auditLogs = ref<any[]>([])

const { connected: wsConnected, on: wsOn } = useWebSocket()

// Refresh summary counts from API
const refreshSummary = async () => {
  const [s, al] = await Promise.all([
    getDashboardSummary().catch(() => ({ data: summary.value })),
    getAlertSummary().catch(() => ({ data: alertSummary.value })),
  ])
  summary.value = s.data
  alertSummary.value = al.data
}

// Handle real-time device status updates
wsOn('device_status', (data: any) => {
  const idx = devices.value.findIndex((d: any) => d.mac === data.mac || d.id === data.device_id)
  if (idx !== -1) {
    devices.value[idx].cpu_usage = data.cpu_usage
    devices.value[idx].mem_usage = data.mem_usage
    devices.value[idx].status = data.status
    devices.value[idx].uptime_secs = data.uptime_secs
  }
  // Refresh summary when device status changes
  refreshSummary()
})

// Handle real-time alert events
wsOn('alert', () => {
  refreshSummary()
})

const formatTime = (t: string) => {
  const d = new Date(t)
  return `${d.getMonth() + 1}/${d.getDate()} ${d.getHours().toString().padStart(2, '0')}:${d.getMinutes().toString().padStart(2, '0')}`
}

const logType = (action: string) => {
  if (action === 'login') return 'primary'
  if (action === 'create') return 'success'
  if (action === 'delete' || action === 'reboot') return 'danger'
  return 'warning'
}

onMounted(async () => {
  const [s, d, t, a, al] = await Promise.all([
    getDashboardSummary().catch(() => ({ data: { total_devices: 0, online_devices: 0, offline_devices: 0, unknown_devices: 0 } })),
    getDevices().catch(() => ({ data: [] })),
    getTemplates().catch(() => ({ data: [] })),
    getAuditLogs().catch(() => ({ data: [] })),
    getAlertSummary().catch(() => ({ data: { total: 0, unresolved: 0, warning: 0, critical: 0 } })),
  ])
  summary.value = s.data
  devices.value = d.data
  templateCount.value = t.data.length
  auditLogs.value = a.data.slice(0, 10)
  alertSummary.value = al.data
})
</script>

<style scoped>
.stat-cards .el-card { text-align: center; }
</style>
