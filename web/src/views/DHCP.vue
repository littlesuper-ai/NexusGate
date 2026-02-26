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
          <el-button type="primary" @click="openPoolDialog()" :disabled="!selectedDevice">添加地址池</el-button>
        </el-row>
        <el-table :data="pools" stripe border size="small">
          <template #empty><el-empty description="暂无 DHCP 地址池" :image-size="60" /></template>
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
          <el-table-column label="操作" width="120">
            <template #default="{ row }">
              <el-button size="small" type="primary" link @click="openPoolDialog(row)">编辑</el-button>
              <el-button size="small" type="danger" link @click="handleDeletePool(row.id)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- Static Leases -->
      <el-tab-pane label="静态绑定" name="lease">
        <el-row justify="end" style="margin-bottom: 12px">
          <el-button type="primary" @click="openLeaseDialog()" :disabled="!selectedDevice">添加绑定</el-button>
        </el-row>
        <el-table :data="leases" stripe border size="small">
          <template #empty><el-empty description="暂无静态绑定" :image-size="60" /></template>
          <el-table-column prop="id" label="ID" width="60" />
          <el-table-column prop="name" label="主机名" width="150" />
          <el-table-column prop="mac" label="MAC 地址" width="180">
            <template #default="{ row }">
              <span style="font-family: monospace">{{ row.mac }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="ip" label="IP 地址" width="150" />
          <el-table-column prop="created_at" label="创建时间" />
          <el-table-column label="操作" width="120">
            <template #default="{ row }">
              <el-button size="small" type="primary" link @click="openLeaseDialog(row)">编辑</el-button>
              <el-button size="small" type="danger" link @click="handleDeleteLease(row.id)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>
    </el-tabs>

    <!-- Pool dialog (create/edit) -->
    <el-dialog v-model="showPoolDialog" :title="editingPoolId ? '编辑 DHCP 地址池' : '添加 DHCP 地址池'" width="480">
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
        <el-button type="primary" @click="submitPool" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>

    <!-- Lease dialog (create/edit) -->
    <el-dialog v-model="showLeaseDialog" :title="editingLeaseId ? '编辑静态绑定' : '添加静态绑定'" width="450">
      <el-form :model="leaseForm" label-width="80px">
        <el-form-item label="主机名"><el-input v-model="leaseForm.name" placeholder="printer-1" /></el-form-item>
        <el-form-item label="MAC"><el-input v-model="leaseForm.mac" placeholder="AA:BB:CC:DD:EE:FF" /></el-form-item>
        <el-form-item label="IP"><el-input v-model="leaseForm.ip" placeholder="192.168.1.50" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showLeaseDialog = false">取消</el-button>
        <el-button type="primary" @click="submitLease" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getDevices, getDHCPPools, createDHCPPool, updateDHCPPool, deleteDHCPPool,
  getStaticLeases, createStaticLease, updateStaticLease, deleteStaticLease, applyDHCP, apiErr,
} from '../api'

const devices = ref<any[]>([])
const selectedDevice = ref<number | null>(null)
const activeTab = ref('pool')
const pools = ref<any[]>([])
const leases = ref<any[]>([])

const applying = ref(false)
const showPoolDialog = ref(false)
const showLeaseDialog = ref(false)
const submitting = ref(false)
const editingPoolId = ref<number | null>(null)
const editingLeaseId = ref<number | null>(null)

const poolForm = reactive({
  interface: 'lan', start: 100, limit: 150, lease_time: '12h',
  dns: '', gateway: '', enabled: true,
})
const leaseForm = reactive({ name: '', mac: '', ip: '' })

const defaultPoolForm = { interface: 'lan', start: 100, limit: 150, lease_time: '12h', dns: '', gateway: '', enabled: true }
const defaultLeaseForm = { name: '', mac: '', ip: '' }

const openPoolDialog = (row?: any) => {
  if (row) {
    editingPoolId.value = row.id
    Object.assign(poolForm, { interface: row.interface, start: row.start, limit: row.limit, lease_time: row.lease_time, dns: row.dns || '', gateway: row.gateway || '', enabled: row.enabled })
  } else {
    editingPoolId.value = null
    Object.assign(poolForm, defaultPoolForm)
  }
  showPoolDialog.value = true
}

const openLeaseDialog = (row?: any) => {
  if (row) {
    editingLeaseId.value = row.id
    Object.assign(leaseForm, { name: row.name, mac: row.mac, ip: row.ip })
  } else {
    editingLeaseId.value = null
    Object.assign(leaseForm, defaultLeaseForm)
  }
  showLeaseDialog.value = true
}

const fetchAll = async () => {
  if (!selectedDevice.value) { pools.value = []; leases.value = []; return }
  try {
    const [p, l] = await Promise.all([
      getDHCPPools(selectedDevice.value),
      getStaticLeases(selectedDevice.value),
    ])
    pools.value = p.data
    leases.value = l.data
  } catch { ElMessage.error('获取 DHCP 数据失败') }
}

const submitPool = async () => {
  submitting.value = true
  try {
    if (editingPoolId.value) {
      await updateDHCPPool(editingPoolId.value, { ...poolForm, device_id: selectedDevice.value })
      ElMessage.success('已更新')
    } else {
      await createDHCPPool({ ...poolForm, device_id: selectedDevice.value })
      ElMessage.success('已添加')
    }
    showPoolDialog.value = false
    Object.assign(poolForm, defaultPoolForm)
    editingPoolId.value = null
    fetchAll()
  } catch (e: any) { ElMessage.error(apiErr(e, '保存地址池失败')) }
  finally { submitting.value = false }
}

const handleDeletePool = async (id: number) => {
  await ElMessageBox.confirm('确认删除此地址池？', '确认')
  try {
    await deleteDHCPPool(id); ElMessage.success('已删除'); fetchAll()
  } catch { ElMessage.error('删除失败') }
}

const submitLease = async () => {
  submitting.value = true
  try {
    if (editingLeaseId.value) {
      await updateStaticLease(editingLeaseId.value, { ...leaseForm, device_id: selectedDevice.value })
      ElMessage.success('已更新')
    } else {
      await createStaticLease({ ...leaseForm, device_id: selectedDevice.value })
      ElMessage.success('已添加')
    }
    showLeaseDialog.value = false
    Object.assign(leaseForm, defaultLeaseForm)
    editingLeaseId.value = null
    fetchAll()
  } catch (e: any) { ElMessage.error(apiErr(e, '保存绑定失败')) }
  finally { submitting.value = false }
}

const handleDeleteLease = async (id: number) => {
  await ElMessageBox.confirm('确认删除此绑定？', '确认')
  try {
    await deleteStaticLease(id); ElMessage.success('已删除'); fetchAll()
  } catch { ElMessage.error('删除失败') }
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
