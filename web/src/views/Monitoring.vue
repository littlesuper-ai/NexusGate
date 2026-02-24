<template>
  <div>
    <el-row justify="space-between" align="middle" style="margin-bottom: 16px">
      <el-col :span="12">
        <span style="margin-right: 12px">选择设备：</span>
        <el-select v-model="deviceId" placeholder="选择设备" @change="fetchMetrics" style="width: 260px">
          <el-option v-for="d in devices" :key="d.id" :label="`${d.name} (${d.ip_address})`" :value="d.id" />
        </el-select>
      </el-col>
      <el-col :span="12" style="text-align: right">
        <el-tag :type="wsConnected ? 'success' : 'info'" size="small" style="margin-right: 12px">
          {{ wsConnected ? '实时' : '离线' }}
        </el-tag>
        <el-button :icon="Refresh" @click="fetchMetrics" :disabled="!deviceId">刷新</el-button>
      </el-col>
    </el-row>

    <div v-if="metrics.length > 0">
      <el-row :gutter="20">
        <el-col :span="12">
          <el-card>
            <template #header>CPU 使用率 (%)</template>
            <div ref="cpuChartRef" style="height: 300px"></div>
          </el-card>
        </el-col>
        <el-col :span="12">
          <el-card>
            <template #header>内存使用率 (%)</template>
            <div ref="memChartRef" style="height: 300px"></div>
          </el-card>
        </el-col>
      </el-row>

      <el-row :gutter="20" style="margin-top: 20px">
        <el-col :span="12">
          <el-card>
            <template #header>网络流量 (MB)</template>
            <div ref="netChartRef" style="height: 300px"></div>
          </el-card>
        </el-col>
        <el-col :span="12">
          <el-card>
            <template #header>连接追踪数</template>
            <div ref="connChartRef" style="height: 300px"></div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <el-empty v-else-if="deviceId" description="暂无监控数据" />
    <el-empty v-else description="请选择一台设备查看监控数据" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick, markRaw, watch } from 'vue'
import { Refresh } from '@element-plus/icons-vue'
import * as echarts from 'echarts'
import { getDevices, getDeviceMetrics } from '../api'
import { useWebSocket } from '../composables/useWebSocket'

const devices = ref<any[]>([])
const deviceId = ref<number | null>(null)
const metrics = ref<any[]>([])
const { connected: wsConnected, on: wsOn } = useWebSocket()

// Handle real-time metrics updates via WebSocket
wsOn('device_status', (data: any) => {
  if (!deviceId.value) return
  const selectedDev = devices.value.find((d: any) => d.id === deviceId.value)
  if (!selectedDev) return
  if (data.mac !== selectedDev.mac && data.device_id !== deviceId.value) return

  // Append new data point to metrics and re-render charts
  const point = {
    cpu_usage: data.cpu_usage,
    mem_usage: data.mem_usage,
    rx_bytes: data.rx_bytes || 0,
    tx_bytes: data.tx_bytes || 0,
    conntrack: data.conntrack || 0,
    collected_at: new Date().toISOString(),
  }
  metrics.value.unshift(point)
  if (metrics.value.length > 500) metrics.value.pop()
  renderCharts()
})

const cpuChartRef = ref<HTMLElement>()
const memChartRef = ref<HTMLElement>()
const netChartRef = ref<HTMLElement>()
const connChartRef = ref<HTMLElement>()

let cpuChart: echarts.ECharts | null = null
let memChart: echarts.ECharts | null = null
let netChart: echarts.ECharts | null = null
let connChart: echarts.ECharts | null = null

const fetchMetrics = async () => {
  if (!deviceId.value) return
  const { data } = await getDeviceMetrics(deviceId.value)
  metrics.value = data
  await nextTick()
  renderCharts()
}

const renderCharts = () => {
  const sorted = [...metrics.value].reverse()
  const times = sorted.map((m: any) => {
    const d = new Date(m.collected_at)
    return `${d.getHours().toString().padStart(2, '0')}:${d.getMinutes().toString().padStart(2, '0')}`
  })

  const commonOptions = {
    tooltip: { trigger: 'axis' as const },
    grid: { left: 50, right: 20, top: 20, bottom: 30 },
    xAxis: { type: 'category' as const, data: times, axisLabel: { interval: Math.floor(times.length / 8) } },
  }

  // CPU
  if (cpuChartRef.value) {
    cpuChart?.dispose()
    cpuChart = markRaw(echarts.init(cpuChartRef.value))
    cpuChart.setOption({
      ...commonOptions,
      yAxis: { type: 'value', max: 100 },
      series: [{ data: sorted.map((m: any) => m.cpu_usage?.toFixed(1)), type: 'line', smooth: true, areaStyle: { opacity: 0.3 }, itemStyle: { color: '#409EFF' } }],
    })
  }

  // Memory
  if (memChartRef.value) {
    memChart?.dispose()
    memChart = markRaw(echarts.init(memChartRef.value))
    memChart.setOption({
      ...commonOptions,
      yAxis: { type: 'value', max: 100 },
      series: [{ data: sorted.map((m: any) => m.mem_usage?.toFixed(1)), type: 'line', smooth: true, areaStyle: { opacity: 0.3 }, itemStyle: { color: '#E6A23C' } }],
    })
  }

  // Network
  if (netChartRef.value) {
    netChart?.dispose()
    netChart = markRaw(echarts.init(netChartRef.value))
    netChart.setOption({
      ...commonOptions,
      yAxis: { type: 'value' },
      legend: { data: ['RX', 'TX'], bottom: 0 },
      grid: { ...commonOptions.grid, bottom: 50 },
      series: [
        { name: 'RX', data: sorted.map((m: any) => (m.rx_bytes / 1048576).toFixed(1)), type: 'line', smooth: true, itemStyle: { color: '#67C23A' } },
        { name: 'TX', data: sorted.map((m: any) => (m.tx_bytes / 1048576).toFixed(1)), type: 'line', smooth: true, itemStyle: { color: '#F56C6C' } },
      ],
    })
  }

  // Conntrack
  if (connChartRef.value) {
    connChart?.dispose()
    connChart = markRaw(echarts.init(connChartRef.value))
    connChart.setOption({
      ...commonOptions,
      yAxis: { type: 'value' },
      series: [{ data: sorted.map((m: any) => m.conntrack), type: 'bar', itemStyle: { color: '#909399' } }],
    })
  }
}

onMounted(async () => {
  const { data } = await getDevices()
  devices.value = data
})
</script>
