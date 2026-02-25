<template>
  <div>
    <el-row justify="space-between" align="middle" style="margin-bottom: 16px">
      <el-col :span="12">
        <el-input v-model="search" placeholder="搜索设备名称/MAC/IP" clearable style="width: 300px" />
      </el-col>
      <el-col :span="12" style="text-align: right">
        <el-radio-group v-model="statusFilter" @change="fetchDevices">
          <el-radio-button value="">全部</el-radio-button>
          <el-radio-button value="online">在线</el-radio-button>
          <el-radio-button value="offline">离线</el-radio-button>
        </el-radio-group>
      </el-col>
    </el-row>

    <el-table :data="filteredDevices" v-loading="loading" stripe @selection-change="handleSelectionChange">
      <el-table-column v-if="canWrite" type="selection" width="45" />
      <el-table-column prop="name" label="设备名称" />
      <el-table-column prop="mac" label="MAC 地址" width="180" />
      <el-table-column prop="ip_address" label="IP 地址" width="150" />
      <el-table-column prop="model" label="型号" width="160" />
      <el-table-column prop="firmware" label="固件版本" width="120" />
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 'online' ? 'success' : row.status === 'offline' ? 'danger' : 'info'" size="small">
            {{ row.status === 'online' ? '在线' : row.status === 'offline' ? '离线' : '未知' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="group" label="分组" width="120" />
      <el-table-column label="CPU" width="80">
        <template #default="{ row }">{{ row.cpu_usage?.toFixed(1) }}%</template>
      </el-table-column>
      <el-table-column label="内存" width="80">
        <template #default="{ row }">{{ row.mem_usage?.toFixed(1) }}%</template>
      </el-table-column>
      <el-table-column label="操作" :width="canWrite ? 200 : 80" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="$router.push(`/devices/${row.id}`)">详情</el-button>
          <el-button v-if="canWrite" size="small" type="warning" @click="handleReboot(row)">重启</el-button>
          <el-button v-if="canWrite" size="small" type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- Bulk actions & pagination -->
    <el-row justify="space-between" align="middle" style="margin-top: 16px">
      <el-col :span="12">
        <template v-if="canWrite">
          <el-button type="warning" size="small" :disabled="selectedRows.length === 0" @click="handleBulkReboot">
            批量重启 ({{ selectedRows.length }})
          </el-button>
          <el-button type="danger" size="small" :disabled="selectedRows.length === 0" @click="handleBulkDelete">
            批量删除 ({{ selectedRows.length }})
          </el-button>
        </template>
      </el-col>
      <el-col :span="12" style="text-align: right">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :page-sizes="[20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next"
          @size-change="fetchDevices"
          @current-change="fetchDevices"
        />
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getDevices, rebootDevice, deleteDevice, bulkDeleteDevices, bulkRebootDevices } from '../api'

interface Device {
  id: number
  name: string
  mac: string
  ip_address: string
  model: string
  firmware: string
  status: string
  group: string
  cpu_usage: number
  mem_usage: number
}

const devices = ref<Device[]>([])
const loading = ref(false)
const search = ref('')
const statusFilter = ref('')
const selectedRows = ref<Device[]>([])
const page = ref(1)
const pageSize = ref(50)
const total = ref(0)
const canWrite = computed(() => {
  const role = localStorage.getItem('role')
  return role === 'admin' || role === 'operator'
})

const filteredDevices = computed(() => {
  const q = search.value.toLowerCase()
  if (!q) return devices.value
  return devices.value.filter(
    (d) => d.name.toLowerCase().includes(q) || d.mac.toLowerCase().includes(q) || d.ip_address?.includes(q)
  )
})

const handleSelectionChange = (rows: Device[]) => {
  selectedRows.value = rows
}

const fetchDevices = async () => {
  loading.value = true
  try {
    const params: Record<string, string | number> = { page: page.value, page_size: pageSize.value }
    if (statusFilter.value) params.status = statusFilter.value
    const { data } = await getDevices(params as any)
    if (data.data) {
      devices.value = data.data
      total.value = data.total
    } else {
      devices.value = data
      total.value = data.length
    }
  } catch {
    ElMessage.error('获取设备列表失败')
  } finally {
    loading.value = false
  }
}

const handleReboot = async (device: Device) => {
  await ElMessageBox.confirm(`确认重启设备 "${device.name}"？`, '重启确认', { type: 'warning' })
  await rebootDevice(device.id)
  ElMessage.success('重启指令已发送')
}

const handleDelete = async (device: Device) => {
  await ElMessageBox.confirm(`确认删除设备 "${device.name}"？此操作不可恢复`, '删除确认', { type: 'warning' })
  await deleteDevice(device.id)
  ElMessage.success('已删除')
  fetchDevices()
}

const handleBulkReboot = async () => {
  await ElMessageBox.confirm(`确认批量重启 ${selectedRows.value.length} 台设备？`, '批量重启', { type: 'warning' })
  const ids = selectedRows.value.map((d) => d.id)
  await bulkRebootDevices(ids)
  ElMessage.success(`已发送重启指令到 ${selectedRows.value.length} 台设备`)
}

const handleBulkDelete = async () => {
  await ElMessageBox.confirm(`确认批量删除 ${selectedRows.value.length} 台设备？此操作不可恢复`, '批量删除', { type: 'warning' })
  const ids = selectedRows.value.map((d) => d.id)
  await bulkDeleteDevices(ids)
  ElMessage.success('已删除')
  fetchDevices()
}

onMounted(fetchDevices)
</script>
