<template>
  <div>
    <!-- Device selector -->
    <el-card style="margin-bottom: 16px">
      <el-row align="middle" :gutter="16">
        <el-col :span="6">
          <el-select v-model="selectedDevice" placeholder="选择设备" filterable clearable style="width: 100%" @change="fetchAll">
            <el-option v-for="d in devices" :key="d.id" :label="`${d.name} (${d.ip_address})`" :value="d.id" />
          </el-select>
        </el-col>
        <el-col :span="4">
          <el-button type="success" @click="handleApply" :disabled="!selectedDevice" :loading="applying">应用到设备</el-button>
        </el-col>
      </el-row>
    </el-card>

    <el-tabs v-model="activeTab">
      <!-- DHCP Pools -->
      <el-tab-pane label="地址池" name="pool">
        <el-row justify="end" style="margin-bottom: 12px">
          <el-button type="primary" @click="showPoolDialog = true" :disabled="!selectedDevice">添加地址池</el-button>
        </el-row>
        <el-table :data="pools" stripe border size="small">
          <el-table-column prop="id" label="ID" width="60" />
          <el-table-column prop="interface" label="接口" width="100" />
          <el-table-column prop="start" label="起始" width="80" />
          <el-table-column prop="limit" label="数量" width="80" />
          <el-table-column prop="lease_time" label="租期" width="80" />
          <el-table-column prop="dns" label="DNS 服务器" />
          <el-table-column prop="gateway" label="网关" width="140" />
          <el-table-column prop="enabled" label="启用" width="70">
            <template #default="{ row }">
              <el-tag :type="row.enabled ? 'success' : 'info'" size="small">{{ row.enabled ? '是' : '否' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="80">
            <template #default="{ row }">
              <el-button size="small" type="danger" link @click="handleDeletePool(row.id)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- Static Leases -->
      <el-tab-pane label="静态绑定" name="lease">
        <el-row justify="end" style="margin-bottom: 12px">
          <el-button type="primary" @click="showLeaseDialog = true" :disabled="!selectedDevice">添加绑定</el-button>
        </el-row>
        <el-table :data="leases" stripe border size="small">
          <el-table-column prop="id" label="ID" width="60" />
          <el-table-column prop="name" label="主机名" width="150" />
          <el-table-column prop="mac" label="MAC 地址" width="180">
            <template #default="{ row }">
              <span style="font-family: monospace">{{ row.mac }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="ip" label="IP 地址" width="150" />
          <el-table-column prop="created_at" label="创建时间" />
          <el-table-column label="操作" width="80">
            <template #default="{ row }">
              <el-button size="small" type="danger" link @click="handleDeleteLease(row.id)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>
    </el-tabs>

    <!-- Add Pool dialog -->
    <el-dialog v-model="showPoolDialog" title="添加 DHCP 地址池" width="480">
      <el-form :model="poolForm" label-width="80px">
        <el-form-item label="接口"><el-input v-model="poolForm.interface" placeholder="lan / guest / iot" /></el-form-item>
        <el-form-item label="起始地址"><el-input-number v-model="poolForm.start" :min="1" :max="254" /></el-form-item>
        <el-form-item label="地址数量"><el-input-number v-model="poolForm.limit" :min="1" :max="254" /></el-form-item>
        <el-form-item label="租期"><el-input v-model="poolForm.lease_time" placeholder="12h" /></el-form-item>
        <el-form-item label="DNS"><el-input v-model="poolForm.dns" placeholder="8.8.8.8,223.5.5.5" /></el-form-item>
        <el-form-item label="网关"><el-input v-model="poolForm.gateway" placeholder="192.168.1.1" /></el-form-item>
        <el-form-item label="启用"><el-switch v-model="poolForm.enabled" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showPoolDialog = false">取消</el-button>
        <el-button type="primary" @click="submitPool">确定</el-button>
      </template>
    </el-dialog>

    <!-- Add Lease dialog -->
    <el-dialog v-model="showLeaseDialog" title="添加静态绑定" width="450">
      <el-form :model="leaseForm" label-width="80px">
        <el-form-item label="主机名"><el-input v-model="leaseForm.name" placeholder="printer-1" /></el-form-item>
        <el-form-item label="MAC"><el-input v-model="leaseForm.mac" placeholder="AA:BB:CC:DD:EE:FF" /></el-form-item>
        <el-form-item label="IP"><el-input v-model="leaseForm.ip" placeholder="192.168.1.50" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showLeaseDialog = false">取消</el-button>
        <el-button type="primary" @click="submitLease">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getDevices, getDHCPPools, createDHCPPool, deleteDHCPPool,
  getStaticLeases, createStaticLease, deleteStaticLease, applyDHCP,
} from '../api'

const devices = ref<any[]>([])
const selectedDevice = ref<number | null>(null)
const activeTab = ref('pool')
const pools = ref<any[]>([])
const leases = ref<any[]>([])

const applying = ref(false)
const showPoolDialog = ref(false)
const showLeaseDialog = ref(false)

const poolForm = reactive({
  interface: 'lan', start: 100, limit: 150, lease_time: '12h',
  dns: '', gateway: '', enabled: true,
})
const leaseForm = reactive({ name: '', mac: '', ip: '' })

const fetchAll = async () => {
  if (!selectedDevice.value) { pools.value = []; leases.value = []; return }
  const [p, l] = await Promise.all([
    getDHCPPools(selectedDevice.value),
    getStaticLeases(selectedDevice.value),
  ])
  pools.value = p.data
  leases.value = l.data
}

const submitPool = async () => {
  await createDHCPPool({ ...poolForm, device_id: selectedDevice.value })
  ElMessage.success('已添加'); showPoolDialog.value = false
  Object.assign(poolForm, { interface: 'lan', start: 100, limit: 150, lease_time: '12h', dns: '', gateway: '', enabled: true })
  fetchAll()
}
const handleDeletePool = async (id: number) => {
  await ElMessageBox.confirm('确认删除此地址池？', '确认')
  await deleteDHCPPool(id); ElMessage.success('已删除'); fetchAll()
}
const submitLease = async () => {
  await createStaticLease({ ...leaseForm, device_id: selectedDevice.value })
  ElMessage.success('已添加'); showLeaseDialog.value = false
  Object.assign(leaseForm, { name: '', mac: '', ip: '' })
  fetchAll()
}
const handleDeleteLease = async (id: number) => {
  await ElMessageBox.confirm('确认删除此绑定？', '确认')
  await deleteStaticLease(id); ElMessage.success('已删除'); fetchAll()
}

const handleApply = async () => {
  if (!selectedDevice.value) return
  await ElMessageBox.confirm('确认将 DHCP 配置应用到该设备？', '应用配置')
  applying.value = true
  try {
    await applyDHCP(selectedDevice.value)
    ElMessage.success('DHCP 配置已推送')
  } finally {
    applying.value = false
  }
}

onMounted(async () => {
  const { data } = await getDevices()
  devices.value = data
})
</script>
