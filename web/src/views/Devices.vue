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

    <el-table :data="filteredDevices" v-loading="loading" stripe>
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
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="$router.push(`/devices/${row.id}`)">详情</el-button>
          <el-button size="small" type="warning" @click="handleReboot(row)">重启</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getDevices, rebootDevice, deleteDevice } from '../api'

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

const filteredDevices = computed(() => {
  const q = search.value.toLowerCase()
  return devices.value.filter(
    (d) => d.name.toLowerCase().includes(q) || d.mac.toLowerCase().includes(q) || d.ip_address?.includes(q)
  )
})

const fetchDevices = async () => {
  loading.value = true
  try {
    const params: Record<string, string> = {}
    if (statusFilter.value) params.status = statusFilter.value
    const { data } = await getDevices(params)
    devices.value = data
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

onMounted(fetchDevices)
</script>
