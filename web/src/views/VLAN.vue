<template>
  <div>
    <!-- Device selector -->
    <el-card style="margin-bottom: 16px">
      <el-row align="middle" :gutter="16">
        <el-col :span="6">
          <el-select v-model="selectedDevice" placeholder="选择设备" filterable clearable style="width: 100%" @change="fetchVLANs">
            <el-option v-for="d in devices" :key="d.id" :label="`${d.name} (${d.ip_address})`" :value="d.id" />
          </el-select>
        </el-col>
        <el-col :span="4">
          <el-button type="primary" @click="showDialog = true" :disabled="!selectedDevice">添加 VLAN</el-button>
        </el-col>
        <el-col :span="4">
          <el-button type="success" @click="handleApply" :disabled="!selectedDevice" :loading="applying">应用到设备</el-button>
        </el-col>
      </el-row>
    </el-card>

    <el-table :data="vlans" stripe border size="small">
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column prop="vid" label="VLAN ID" width="100" />
      <el-table-column prop="name" label="名称" width="120" />
      <el-table-column prop="interface" label="接口" width="140">
        <template #default="{ row }">
          <span style="font-family: monospace">{{ row.interface || '-' }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="ip_addr" label="IP 地址" width="140" />
      <el-table-column prop="netmask" label="子网掩码" width="150" />
      <el-table-column prop="isolated" label="隔离" width="70">
        <template #default="{ row }">
          <el-tag :type="row.isolated ? 'danger' : 'success'" size="small">{{ row.isolated ? '是' : '否' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="140">
        <template #default="{ row }">
          <el-button size="small" link @click="openEdit(row)">编辑</el-button>
          <el-button size="small" type="danger" link @click="handleDelete(row.id)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-empty v-if="vlans.length === 0 && selectedDevice" description="暂无 VLAN 配置" />

    <!-- Create/Edit dialog -->
    <el-dialog v-model="showDialog" :title="editing ? '编辑 VLAN' : '添加 VLAN'" width="480">
      <el-form :model="form" label-width="90px">
        <el-form-item label="VLAN ID"><el-input-number v-model="form.vid" :min="1" :max="4094" /></el-form-item>
        <el-form-item label="名称"><el-input v-model="form.name" placeholder="office / server / guest" /></el-form-item>
        <el-form-item label="接口"><el-input v-model="form.interface" placeholder="br-lan.10" /></el-form-item>
        <el-form-item label="IP 地址"><el-input v-model="form.ip_addr" placeholder="10.0.10.1" /></el-form-item>
        <el-form-item label="子网掩码"><el-input v-model="form.netmask" placeholder="255.255.255.0" /></el-form-item>
        <el-form-item label="隔离"><el-switch v-model="form.isolated" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" @click="submitForm">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getDevices, getVLANs, createVLAN, updateVLAN, deleteVLAN, applyVLAN } from '../api'

const devices = ref<any[]>([])
const selectedDevice = ref<number | null>(null)
const vlans = ref<any[]>([])
const applying = ref(false)
const showDialog = ref(false)
const editing = ref(false)
const editId = ref(0)

const form = reactive({
  vid: 10, name: '', interface: '', ip_addr: '', netmask: '255.255.255.0', isolated: false,
})

const resetForm = () => {
  Object.assign(form, { vid: 10, name: '', interface: '', ip_addr: '', netmask: '255.255.255.0', isolated: false })
  editing.value = false; editId.value = 0
}

const fetchVLANs = async () => {
  if (!selectedDevice.value) { vlans.value = []; return }
  try {
    const { data } = await getVLANs(selectedDevice.value)
    vlans.value = data
  } catch { ElMessage.error('获取 VLAN 数据失败') }
}

const openEdit = (row: any) => {
  editing.value = true; editId.value = row.id
  Object.assign(form, {
    vid: row.vid, name: row.name, interface: row.interface,
    ip_addr: row.ip_addr, netmask: row.netmask, isolated: row.isolated,
  })
  showDialog.value = true
}

const submitForm = async () => {
  try {
    if (editing.value) {
      await updateVLAN(editId.value, { ...form, device_id: selectedDevice.value })
      ElMessage.success('已更新')
    } else {
      await createVLAN({ ...form, device_id: selectedDevice.value })
      ElMessage.success('已添加')
    }
    showDialog.value = false; resetForm(); fetchVLANs()
  } catch { ElMessage.error('保存 VLAN 失败') }
}

const handleDelete = async (id: number) => {
  await ElMessageBox.confirm('确认删除此 VLAN？', '确认')
  try {
    await deleteVLAN(id); ElMessage.success('已删除'); fetchVLANs()
  } catch { ElMessage.error('删除 VLAN 失败') }
}

const handleApply = async () => {
  if (!selectedDevice.value) return
  await ElMessageBox.confirm('确认将 VLAN 配置应用到该设备？', '应用配置')
  applying.value = true
  try {
    await applyVLAN(selectedDevice.value)
    ElMessage.success('VLAN 配置已推送')
  } catch {
    ElMessage.error('VLAN 配置下发失败')
  } finally {
    applying.value = false
  }
}

onMounted(async () => {
  try {
    const { data } = await getDevices()
    devices.value = data
  } catch { ElMessage.error('获取设备列表失败') }
})
</script>
