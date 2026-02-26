<template>
  <div v-loading="loading">
    <el-page-header @back="$router.push('/devices')" :content="device?.name || ''" />

    <el-row :gutter="20" style="margin-top: 20px">
      <!-- Left: device info card -->
      <el-col :span="8">
        <el-card>
          <template #header>设备信息</template>
          <el-descriptions :column="1" border size="small">
            <el-descriptions-item label="名称">{{ device?.name }}</el-descriptions-item>
            <el-descriptions-item label="MAC">
              <span style="font-family: monospace">{{ device?.mac }}</span>
            </el-descriptions-item>
            <el-descriptions-item label="IP 地址">{{ device?.ip_address }}</el-descriptions-item>
            <el-descriptions-item label="型号">{{ device?.model }}</el-descriptions-item>
            <el-descriptions-item label="固件">{{ device?.firmware }}</el-descriptions-item>
            <el-descriptions-item label="分组">
              <el-tag v-if="device?.group" size="small">{{ device?.group }}</el-tag>
              <span v-else>-</span>
            </el-descriptions-item>
            <el-descriptions-item label="标签">{{ device?.tags || '-' }}</el-descriptions-item>
            <el-descriptions-item label="状态">
              <el-tag :type="device?.status === 'online' ? 'success' : 'danger'" size="small">
                {{ device?.status === 'online' ? '在线' : '离线' }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="运行时间">{{ formatUptime(device?.uptime_secs) }}</el-descriptions-item>
            <el-descriptions-item label="最后在线">{{ device?.last_seen_at ? new Date(device.last_seen_at).toLocaleString() : '-' }}</el-descriptions-item>
          </el-descriptions>
        </el-card>

        <el-card style="margin-top: 16px">
          <template #header>操作</template>
          <el-space wrap>
            <el-button type="warning" @click="handleReboot">重启设备</el-button>
            <el-button type="primary" @click="showPushConfig = true">下发配置</el-button>
            <el-button @click="handleEditInfo">编辑信息</el-button>
          </el-space>
        </el-card>
      </el-col>

      <!-- Right: tabs -->
      <el-col :span="16">
        <el-card>
          <el-tabs v-model="activeTab">
            <!-- Real-time status -->
            <el-tab-pane label="实时状态" name="status">
              <el-row :gutter="16">
                <el-col :span="12">
                  <div class="metric-label">CPU 使用率</div>
                  <el-progress :percentage="device?.cpu_usage || 0" :stroke-width="20" striped :color="progressColor(device?.cpu_usage)" />
                </el-col>
                <el-col :span="12">
                  <div class="metric-label">内存使用率</div>
                  <el-progress :percentage="device?.mem_usage || 0" :stroke-width="20" striped :color="progressColor(device?.mem_usage)" />
                </el-col>
              </el-row>
              <div ref="miniCpuChart" style="height: 200px; margin-top: 20px"></div>
            </el-tab-pane>

            <!-- Metrics history -->
            <el-tab-pane label="历史指标" name="metrics">
              <el-row :gutter="16">
                <el-col :span="12">
                  <div ref="cpuChartRef" style="height: 250px"></div>
                </el-col>
                <el-col :span="12">
                  <div ref="memChartRef" style="height: 250px"></div>
                </el-col>
              </el-row>
              <el-row :gutter="16" style="margin-top: 12px">
                <el-col :span="12">
                  <div ref="netChartRef" style="height: 250px"></div>
                </el-col>
                <el-col :span="12">
                  <div ref="connChartRef" style="height: 250px"></div>
                </el-col>
              </el-row>
            </el-tab-pane>

            <!-- Config history -->
            <el-tab-pane label="配置历史" name="config">
              <el-table :data="configs" stripe size="small">
                <el-table-column prop="id" label="ID" width="60" />
                <el-table-column prop="status" label="状态" width="100">
                  <template #default="{ row }">
                    <el-tag :type="row.status === 'applied' ? 'success' : row.status === 'failed' ? 'danger' : 'warning'" size="small">{{ row.status }}</el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="created_at" label="时间" width="180" />
                <el-table-column label="内容">
                  <template #default="{ row }">
                    <el-button size="small" link @click="previewConfig = row.content; showPreview = true">查看</el-button>
                  </template>
                </el-table-column>
              </el-table>
              <el-empty v-if="configs.length === 0" description="暂无配置记录" />
            </el-tab-pane>

            <!-- Upgrade history -->
            <el-tab-pane label="升级记录" name="upgrade">
              <el-table :data="upgrades" stripe size="small">
                <el-table-column prop="id" label="ID" width="60" />
                <el-table-column prop="firmware_id" label="固件 ID" width="80" />
                <el-table-column prop="status" label="状态" width="120">
                  <template #default="{ row }">
                    <el-tag :type="row.status === 'success' ? 'success' : row.status === 'failed' ? 'danger' : 'warning'" size="small">{{ row.status }}</el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="created_at" label="发起时间" />
                <el-table-column prop="finished_at" label="完成时间" />
              </el-table>
              <el-empty v-if="upgrades.length === 0" description="暂无升级记录" />
            </el-tab-pane>
          </el-tabs>
        </el-card>
      </el-col>
    </el-row>

    <!-- Config push dialog -->
    <el-dialog v-model="showPushConfig" title="下发配置" width="600">
      <el-form label-width="80px">
        <el-form-item label="配置模板">
          <el-select v-model="configForm.template_id" placeholder="选择模板（可选）" clearable style="width: 100%">
            <el-option v-for="t in templates" :key="t.id" :label="t.name" :value="t.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="UCI 配置">
          <el-input v-model="configForm.content" type="textarea" :rows="10" placeholder="直接输入 UCI 配置内容" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showPushConfig = false">取消</el-button>
        <el-button type="primary" @click="handlePushConfig">下发</el-button>
      </template>
    </el-dialog>

    <!-- Config preview dialog -->
    <el-dialog v-model="showPreview" title="配置内容" width="650">
      <pre style="background: #f5f7fa; padding: 16px; border-radius: 4px; max-height: 500px; overflow: auto; font-size: 13px; line-height: 1.6">{{ previewConfig }}</pre>
    </el-dialog>

    <!-- Edit info dialog -->
    <el-dialog v-model="showEdit" title="编辑设备信息" width="450">
      <el-form :model="editForm" label-width="80px">
        <el-form-item label="名称"><el-input v-model="editForm.name" /></el-form-item>
        <el-form-item label="分组"><el-input v-model="editForm.group" /></el-form-item>
        <el-form-item label="标签"><el-input v-model="editForm.tags" placeholder="逗号分隔" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showEdit = false">取消</el-button>
        <el-button type="primary" @click="saveEdit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted, nextTick, watch, markRaw } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import * as echarts from 'echarts'
import {
  getDevice, updateDevice, rebootDevice, pushConfig,
  getTemplates, getConfigHistory, getDeviceMetrics, getUpgradeHistory,
} from '../api'
import { useWebSocket } from '../composables/useWebSocket'

const route = useRoute()
const router = useRouter()
const deviceId = Number(route.params.id)

const device = ref<any>(null)
const templates = ref<any[]>([])
const configs = ref<any[]>([])
const upgrades = ref<any[]>([])
const metrics = ref<any[]>([])
const loading = ref(false)
const activeTab = ref('status')
const showPushConfig = ref(false)
const showPreview = ref(false)
const showEdit = ref(false)
const previewConfig = ref('')
const configForm = reactive({ template_id: null as number | null, content: '' })
const editForm = reactive({ name: '', group: '', tags: '' })

const cpuChartRef = ref<HTMLElement>()
const memChartRef = ref<HTMLElement>()
const netChartRef = ref<HTMLElement>()
const connChartRef = ref<HTMLElement>()
const miniCpuChart = ref<HTMLElement>()
const chartInstances: echarts.ECharts[] = []

const { on: wsOn } = useWebSocket()

// Real-time status updates
wsOn('device_status', (data: any) => {
  if (!device.value) return
  if (data.mac !== device.value.mac && data.device_id !== deviceId) return
  device.value.cpu_usage = data.cpu_usage
  device.value.mem_usage = data.mem_usage
  device.value.status = data.status
  device.value.uptime_secs = data.uptime_secs
})

// Real-time config ACK updates
wsOn('config_ack', (data: any) => {
  const cfg = configs.value.find((c: any) => c.id === data.config_id)
  if (cfg) cfg.status = data.status
})

// Real-time upgrade ACK updates
wsOn('upgrade_ack', (data: any) => {
  const upg = upgrades.value.find((u: any) => u.id === data.upgrade_id)
  if (upg) upg.status = data.status
})

const formatUptime = (secs?: number) => {
  if (!secs) return '-'
  const d = Math.floor(secs / 86400)
  const h = Math.floor((secs % 86400) / 3600)
  const m = Math.floor((secs % 3600) / 60)
  return `${d}天 ${h}小时 ${m}分钟`
}

const progressColor = (val?: number) => {
  if (!val) return '#409EFF'
  if (val > 85) return '#F56C6C'
  if (val > 60) return '#E6A23C'
  return '#67C23A'
}

const renderCharts = async () => {
  await nextTick()
  const sorted = [...metrics.value].reverse()
  const times = sorted.map((m: any) => {
    const d = new Date(m.collected_at)
    return `${d.getHours().toString().padStart(2, '0')}:${d.getMinutes().toString().padStart(2, '0')}`
  })
  const interval = Math.floor(times.length / 6)

  const mkOpts = (title: string, data: number[], color: string, isArea = true) => ({
    title: { text: title, left: 'center', textStyle: { fontSize: 13 } },
    tooltip: { trigger: 'axis' as const },
    grid: { left: 45, right: 15, top: 35, bottom: 25 },
    xAxis: { type: 'category' as const, data: times, axisLabel: { interval } },
    yAxis: { type: 'value' as const },
    series: [{ data, type: 'line', smooth: true, itemStyle: { color }, ...(isArea ? { areaStyle: { opacity: 0.2 } } : {}) }],
  })

  // Dispose previous chart instances before re-creating
  chartInstances.forEach(c => c.dispose())
  chartInstances.length = 0

  const initChart = (el: HTMLElement) => {
    const c = markRaw(echarts.init(el))
    chartInstances.push(c)
    return c
  }

  if (cpuChartRef.value) {
    initChart(cpuChartRef.value).setOption({ ...mkOpts('CPU %', sorted.map((m: any) => +m.cpu_usage?.toFixed(1)), '#409EFF'), yAxis: { type: 'value', max: 100 } })
  }
  if (memChartRef.value) {
    initChart(memChartRef.value).setOption({ ...mkOpts('内存 %', sorted.map((m: any) => +m.mem_usage?.toFixed(1)), '#E6A23C'), yAxis: { type: 'value', max: 100 } })
  }
  if (netChartRef.value) {
    initChart(netChartRef.value).setOption({
      title: { text: '网络流量 (MB)', left: 'center', textStyle: { fontSize: 13 } },
      tooltip: { trigger: 'axis' }, legend: { data: ['RX', 'TX'], bottom: 0 },
      grid: { left: 50, right: 15, top: 35, bottom: 40 },
      xAxis: { type: 'category', data: times, axisLabel: { interval } },
      yAxis: { type: 'value' },
      series: [
        { name: 'RX', data: sorted.map((m: any) => +(m.rx_bytes / 1048576).toFixed(1)), type: 'line', smooth: true, itemStyle: { color: '#67C23A' } },
        { name: 'TX', data: sorted.map((m: any) => +(m.tx_bytes / 1048576).toFixed(1)), type: 'line', smooth: true, itemStyle: { color: '#F56C6C' } },
      ],
    })
  }
  if (connChartRef.value) {
    initChart(connChartRef.value).setOption(mkOpts('连接追踪', sorted.map((m: any) => m.conntrack), '#909399', false))
  }

  // Mini CPU chart on status tab
  if (miniCpuChart.value && sorted.length > 0) {
    initChart(miniCpuChart.value).setOption({
      title: { text: 'CPU 趋势 (近2小时)', left: 'center', textStyle: { fontSize: 12 } },
      tooltip: { trigger: 'axis' }, grid: { left: 40, right: 15, top: 30, bottom: 25 },
      xAxis: { type: 'category', data: times, axisLabel: { interval } },
      yAxis: { type: 'value', max: 100 },
      series: [{ data: sorted.map((m: any) => +m.cpu_usage?.toFixed(1)), type: 'line', smooth: true, areaStyle: { opacity: 0.15 }, itemStyle: { color: '#409EFF' } }],
    })
  }
}

watch(activeTab, (tab) => {
  if (tab === 'metrics' && metrics.value.length > 0) nextTick(renderCharts)
})

const handleReboot = async () => {
  await ElMessageBox.confirm('确认重启该设备？', '重启确认', { type: 'warning' })
  await rebootDevice(deviceId)
  ElMessage.success('重启指令已发送')
}

const handlePushConfig = async () => {
  await pushConfig(deviceId, { template_id: configForm.template_id ?? undefined, content: configForm.content || undefined })
  ElMessage.success('配置已下发')
  showPushConfig.value = false
  const { data } = await getConfigHistory(deviceId)
  configs.value = data
}

const handleEditInfo = () => {
  editForm.name = device.value?.name || ''
  editForm.group = device.value?.group || ''
  editForm.tags = device.value?.tags || ''
  showEdit.value = true
}

const saveEdit = async () => {
  await updateDevice(deviceId, { ...editForm })
  ElMessage.success('已保存')
  showEdit.value = false
  const { data } = await getDevice(deviceId)
  device.value = data
}

onMounted(async () => {
  loading.value = true
  try {
    const [devRes, tplRes, cfgRes, metRes, upgRes] = await Promise.all([
      getDevice(deviceId),
      getTemplates(),
      getConfigHistory(deviceId),
      getDeviceMetrics(deviceId),
      getUpgradeHistory(deviceId).catch(() => ({ data: [] })),
    ])
    device.value = devRes.data
    templates.value = tplRes.data
    configs.value = cfgRes.data
    metrics.value = metRes.data
    upgrades.value = upgRes.data

    await nextTick()
    if (metrics.value.length > 0) renderCharts()
  } catch {
    ElMessage.error('获取设备信息失败')
    router.push('/devices')
  } finally {
    loading.value = false
  }
})

onUnmounted(() => {
  chartInstances.forEach(c => c.dispose())
  chartInstances.length = 0
})
</script>

<style scoped>
.metric-label {
  font-size: 13px;
  color: #606266;
  margin-bottom: 8px;
}
</style>
