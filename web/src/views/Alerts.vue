<template>
  <div>
    <!-- Summary cards -->
    <el-row :gutter="16" style="margin-bottom: 16px">
      <el-col :span="6">
        <el-statistic title="未解决告警" :value="summary.unresolved" />
      </el-col>
      <el-col :span="6">
        <el-statistic title="严重告警">
          <template #default>
            <span :style="{ color: summary.critical > 0 ? '#F56C6C' : undefined }">{{ summary.critical }}</span>
          </template>
        </el-statistic>
      </el-col>
      <el-col :span="6">
        <el-statistic title="警告" :value="summary.warning" />
      </el-col>
      <el-col :span="6">
        <el-statistic title="历史总计" :value="summary.total" />
      </el-col>
    </el-row>

    <!-- Filters -->
    <el-card style="margin-bottom: 16px">
      <el-row :gutter="16" align="middle">
        <el-col :span="4">
          <el-select v-model="filter.resolved" placeholder="状态" clearable style="width: 100%" @change="fetchAlerts">
            <el-option label="未解决" value="false" />
            <el-option label="已解决" value="true" />
          </el-select>
        </el-col>
        <el-col :span="4">
          <el-select v-model="filter.severity" placeholder="级别" clearable style="width: 100%" @change="fetchAlerts">
            <el-option label="严重" value="critical" />
            <el-option label="警告" value="warning" />
          </el-select>
        </el-col>
      </el-row>
    </el-card>

    <!-- Alert table -->
    <el-table :data="alerts" v-loading="loading" stripe border size="small">
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column prop="device_name" label="设备" width="140" />
      <el-table-column prop="metric" label="指标" width="100">
        <template #default="{ row }">
          {{ metricLabel(row.metric) }}
        </template>
      </el-table-column>
      <el-table-column label="当前值" width="100">
        <template #default="{ row }">
          {{ row.metric === 'conntrack' ? row.value : row.value.toFixed(1) + '%' }}
        </template>
      </el-table-column>
      <el-table-column label="阈值" width="100">
        <template #default="{ row }">
          {{ row.metric === 'conntrack' ? row.threshold : row.threshold.toFixed(0) + '%' }}
        </template>
      </el-table-column>
      <el-table-column prop="severity" label="级别" width="80">
        <template #default="{ row }">
          <el-tag :type="row.severity === 'critical' ? 'danger' : 'warning'" size="small">
            {{ row.severity === 'critical' ? '严重' : '警告' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="80">
        <template #default="{ row }">
          <el-tag :type="row.resolved ? 'success' : 'danger'" size="small">
            {{ row.resolved ? '已解决' : '活跃' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="触发时间" width="180">
        <template #default="{ row }">
          {{ new Date(row.created_at).toLocaleString() }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="80">
        <template #default="{ row }">
          <el-button v-if="!row.resolved" size="small" type="success" link @click="handleResolve(row.id)">解决</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-empty v-if="alerts.length === 0" description="暂无告警" />

    <el-row justify="end" style="margin-top: 16px">
      <el-pagination
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :page-sizes="[20, 50, 100]"
        :total="total"
        layout="total, sizes, prev, pager, next"
        @size-change="fetchAlerts"
        @current-change="fetchAlerts"
      />
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getAlerts, getAlertSummary, resolveAlert } from '../api'
import { useWebSocket } from '../composables/useWebSocket'

const alerts = ref<any[]>([])
const summary = ref({ total: 0, unresolved: 0, warning: 0, critical: 0 })
const filter = reactive({ resolved: 'false', severity: '' })
const page = ref(1)
const pageSize = ref(50)
const total = ref(0)
const loading = ref(false)

const { on: wsOn } = useWebSocket()

const metricLabel = (m: string) => {
  const map: Record<string, string> = { cpu: 'CPU', memory: '内存', conntrack: '连接数' }
  return map[m] || m
}

const fetchAlerts = async () => {
  loading.value = true
  try {
    const params: Record<string, string | number> = { page: page.value, page_size: pageSize.value }
    if (filter.resolved) params.resolved = filter.resolved
    if (filter.severity) params.severity = filter.severity
    const [alertRes, sumRes] = await Promise.all([getAlerts(params as any), getAlertSummary()])
    if (alertRes.data.data) {
      alerts.value = alertRes.data.data
      total.value = alertRes.data.total
    } else {
      alerts.value = alertRes.data
      total.value = alertRes.data.length
    }
    summary.value = sumRes.data
  } catch {
    ElMessage.error('获取告警数据失败')
  } finally {
    loading.value = false
  }
}

const handleResolve = async (id: number) => {
  await resolveAlert(id)
  ElMessage.success('已解决')
  fetchAlerts()
}

// Real-time new alert events
wsOn('alert', (data: any) => {
  alerts.value.unshift(data)
  summary.value.unresolved++
  if (data.severity === 'critical') summary.value.critical++
  else summary.value.warning++
})

onMounted(fetchAlerts)
</script>
