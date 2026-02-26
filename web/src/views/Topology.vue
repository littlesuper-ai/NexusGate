<template>
  <div>
    <el-card>
      <template #header>
        <el-row justify="space-between" align="middle">
          <span>网络拓扑</span>
          <el-button @click="fetchAndRender" :icon="Refresh">刷新</el-button>
        </el-row>
      </template>
      <div ref="chartRef" style="height: 600px"></div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, markRaw } from 'vue'
import { Refresh } from '@element-plus/icons-vue'
import * as echarts from 'echarts'
import { ElMessage } from 'element-plus'
import { getDevices } from '../api'
import { useWebSocket } from '../composables/useWebSocket'

const chartRef = ref<HTMLElement>()
let chart: echarts.ECharts | null = null
let deviceMap = new Map<string, Device>()

const { on: wsOn } = useWebSocket()

// Real-time status updates — update node colors
wsOn('device_status', (data: any) => {
  if (!chart) return
  const deviceId = data.device_id
  const status = data.status || 'online'
  const nodeId = `device-${deviceId}`
  const color = status === 'online' ? '#67C23A' : status === 'offline' ? '#F56C6C' : '#909399'

  const option = chart.getOption() as any
  if (option?.series?.[0]?.nodes) {
    const node = option.series[0].nodes.find((n: any) => n.id === nodeId)
    if (node) {
      node.itemStyle = { ...node.itemStyle, color }
    }
    // Also update link style
    for (const link of option.series[0].links || []) {
      if (link.target === nodeId) {
        link.lineStyle = {
          ...link.lineStyle,
          color: status === 'online' ? '#67C23A' : '#ddd',
          type: status === 'online' ? 'solid' : 'dashed',
        }
      }
    }
    chart.setOption(option, true)
  }
})

interface Device {
  id: number; name: string; ip_address: string; status: string
  group: string; model: string; tags: string
}

const fetchAndRender = async () => {
  let data: Device[]
  try {
    const res = await getDevices()
    data = res.data
  } catch {
    ElMessage.error('获取拓扑数据失败')
    return
  }
  const devices: Device[] = data

  // Build topology: Internet -> groups -> devices
  const nodes: any[] = []
  const links: any[] = []
  const categories = [
    { name: 'Internet' },
    { name: '核心网关' },
    { name: '分支网关' },
    { name: 'IoT/其他' },
  ]

  // Internet node
  nodes.push({
    id: 'internet',
    name: 'Internet',
    symbol: 'roundRect',
    symbolSize: [80, 40],
    category: 0,
    itemStyle: { color: '#409EFF' },
    label: { fontSize: 14, fontWeight: 'bold' },
  })

  // Group devices
  const groups = new Map<string, Device[]>()
  for (const d of devices) {
    const g = d.group || '未分组'
    if (!groups.has(g)) groups.set(g, [])
    groups.get(g)!.push(d)
  }

  let idx = 0
  for (const [groupName, groupDevices] of groups) {
    // group hub node
    const hubId = `group-${idx}`
    nodes.push({
      id: hubId,
      name: groupName,
      symbol: 'diamond',
      symbolSize: 35,
      category: 1,
      itemStyle: { color: '#E6A23C' },
      label: { fontSize: 12, fontWeight: 'bold' },
    })
    links.push({ source: 'internet', target: hubId, lineStyle: { width: 3, color: '#409EFF' } })

    for (const d of groupDevices) {
      const isCore = d.tags?.includes('core')
      const isIoT = d.tags?.includes('iot')
      const cat = isCore ? 1 : isIoT ? 3 : 2

      nodes.push({
        id: `device-${d.id}`,
        name: `${d.name}\n${d.ip_address}`,
        symbol: 'circle',
        symbolSize: isCore ? 40 : 28,
        category: cat,
        itemStyle: {
          color: d.status === 'online' ? '#67C23A' : d.status === 'offline' ? '#F56C6C' : '#909399',
          borderColor: isCore ? '#E6A23C' : undefined,
          borderWidth: isCore ? 3 : 0,
        },
        label: { fontSize: 10 },
      })
      links.push({
        source: hubId,
        target: `device-${d.id}`,
        lineStyle: {
          width: isCore ? 2.5 : 1.5,
          color: d.status === 'online' ? '#67C23A' : '#ddd',
          type: d.status === 'online' ? 'solid' : 'dashed',
        },
      })
    }
    idx++
  }

  // VPN links between branch gateways (if tags contain 'vpn')
  const vpnDevices = devices.filter(d => d.tags?.includes('vpn'))
  const coreDevice = devices.find(d => d.tags?.includes('core'))
  if (coreDevice) {
    for (const d of vpnDevices) {
      if (d.id !== coreDevice.id) {
        links.push({
          source: `device-${coreDevice.id}`,
          target: `device-${d.id}`,
          lineStyle: { width: 1.5, color: '#9b59b6', type: 'dashed', curveness: 0.3 },
          label: { show: true, formatter: 'VPN', fontSize: 9, color: '#9b59b6' },
        })
      }
    }
  }

  if (!chartRef.value) return
  chart?.dispose()
  chart = markRaw(echarts.init(chartRef.value))

  chart.setOption({
    tooltip: {},
    legend: { data: categories.map(c => c.name), bottom: 10 },
    series: [{
      type: 'graph',
      layout: 'force',
      roam: true,
      draggable: true,
      categories,
      nodes,
      links,
      force: {
        repulsion: 400,
        edgeLength: [100, 200],
        gravity: 0.1,
      },
      label: { show: true, position: 'bottom' },
      lineStyle: { opacity: 0.8 },
      emphasis: {
        focus: 'adjacency',
        lineStyle: { width: 4 },
      },
    }],
  })
}

onMounted(fetchAndRender)

onUnmounted(() => {
  chart?.dispose()
  chart = null
})
</script>
