<template>
  <div>
    <el-row justify="space-between" align="middle" style="margin-bottom: 16px">
      <el-col :span="12">
        <span style="margin-right: 12px">选择设备：</span>
        <el-select v-model="deviceId" placeholder="选择设备" @change="fetchInterfaces" style="width: 260px">
          <el-option v-for="d in devices" :key="d.id" :label="`${d.name} (${d.ip_address})`" :value="d.id" />
        </el-select>
      </el-col>
      <el-col :span="12" style="text-align: right">
        <el-button type="success" :disabled="!deviceId" :loading="applying" @click="handleApply">应用到设备</el-button>
      </el-col>
    </el-row>

    <el-row :gutter="20">
      <!-- Interfaces -->
      <el-col :span="10">
        <el-card>
          <template #header>
            <el-row justify="space-between" align="middle">
              <span>WireGuard 接口</span>
              <el-button type="primary" size="small" :disabled="!deviceId" @click="showIfaceDialog = true">新建</el-button>
            </el-row>
          </template>
          <el-table :data="interfaces" v-loading="loading" stripe highlight-current-row @current-change="selectInterface">
            <el-table-column prop="name" label="接口" width="80" />
            <el-table-column prop="address" label="地址" />
            <el-table-column prop="listen_port" label="端口" width="80" />
            <el-table-column prop="enabled" label="状态" width="70">
              <template #default="{ row }">
                <el-tag :type="row.enabled ? 'success' : 'info'" size="small">{{ row.enabled ? 'UP' : 'DOWN' }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="" width="60">
              <template #default="{ row }">
                <el-button size="small" type="danger" link @click.stop="handleDeleteIface(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>

      <!-- Peers -->
      <el-col :span="14">
        <el-card>
          <template #header>
            <el-row justify="space-between" align="middle">
              <span>Peers {{ selectedIface ? `(${selectedIface.name})` : '' }}</span>
              <el-button type="primary" size="small" :disabled="!selectedIface" @click="showPeerDialog = true">添加 Peer</el-button>
            </el-row>
          </template>
          <el-table :data="peers" stripe v-if="selectedIface">
            <el-table-column prop="description" label="描述" />
            <el-table-column prop="public_key" label="公钥" width="180">
              <template #default="{ row }">
                <span style="font-family: monospace; font-size: 12px">{{ row.public_key?.substring(0, 20) }}...</span>
              </template>
            </el-table-column>
            <el-table-column prop="allowed_ips" label="允许 IP" />
            <el-table-column prop="endpoint" label="Endpoint" width="160" />
            <el-table-column label="" width="60">
              <template #default="{ row }">
                <el-button size="small" type="danger" link @click="handleDeletePeer(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
          <el-empty v-else description="请先选择左侧接口" />
        </el-card>
      </el-col>
    </el-row>

    <!-- Create Interface -->
    <el-dialog v-model="showIfaceDialog" title="新建 WireGuard 接口" width="480">
      <el-form :model="ifaceForm" label-width="100px">
        <el-form-item label="接口名"><el-input v-model="ifaceForm.name" placeholder="例：wg0" /></el-form-item>
        <el-form-item label="地址"><el-input v-model="ifaceForm.address" placeholder="例：10.99.0.1/24" /></el-form-item>
        <el-form-item label="监听端口"><el-input-number v-model="ifaceForm.listen_port" :min="1" :max="65535" /></el-form-item>
        <el-form-item label="私钥"><el-input v-model="ifaceForm.private_key" placeholder="WireGuard 私钥" show-password /></el-form-item>
        <el-form-item label="公钥"><el-input v-model="ifaceForm.public_key" placeholder="对应公钥（展示用）" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showIfaceDialog = false">取消</el-button>
        <el-button type="primary" @click="handleCreateIface" :loading="submitting">创建</el-button>
      </template>
    </el-dialog>

    <!-- Create Peer -->
    <el-dialog v-model="showPeerDialog" title="添加 WireGuard Peer" width="500">
      <el-form :model="peerForm" label-width="100px">
        <el-form-item label="描述"><el-input v-model="peerForm.description" placeholder="例：上海分支" /></el-form-item>
        <el-form-item label="公钥"><el-input v-model="peerForm.public_key" placeholder="Peer 的 WireGuard 公钥" /></el-form-item>
        <el-form-item label="允许 IP"><el-input v-model="peerForm.allowed_ips" placeholder="逗号分隔，例：10.99.0.2/32, 10.1.0.0/24" /></el-form-item>
        <el-form-item label="Endpoint"><el-input v-model="peerForm.endpoint" placeholder="例：203.0.113.1:51820" /></el-form-item>
        <el-form-item label="Keepalive"><el-input-number v-model="peerForm.keepalive" :min="0" :max="300" /> 秒</el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showPeerDialog = false">取消</el-button>
        <el-button type="primary" @click="handleCreatePeer" :loading="submitting">添加</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getDevices, getVPNInterfaces, createVPNInterface, deleteVPNInterface,
  getVPNPeers, createVPNPeer, deleteVPNPeer, applyVPN, apiErr,
} from '../api'

const devices = ref<any[]>([])
const deviceId = ref<number | null>(null)
const interfaces = ref<any[]>([])
const peers = ref<any[]>([])
const selectedIface = ref<any>(null)
const loading = ref(false)
const submitting = ref(false)
const applying = ref(false)
const showIfaceDialog = ref(false)
const showPeerDialog = ref(false)

const ifaceForm = reactive({ name: 'wg0', address: '10.99.0.1/24', listen_port: 51820, private_key: '', public_key: '' })
const peerForm = reactive({ description: '', public_key: '', allowed_ips: '', endpoint: '', keepalive: 25 })

const fetchInterfaces = async () => {
  if (!deviceId.value) return
  selectedIface.value = null
  peers.value = []
  loading.value = true
  try {
    const { data } = await getVPNInterfaces(deviceId.value)
    interfaces.value = data
  } catch (e: any) { ElMessage.error(apiErr(e, '获取 VPN 接口失败')) }
  finally { loading.value = false }
}

const selectInterface = async (iface: any) => {
  selectedIface.value = iface
  if (iface) {
    try {
      const { data } = await getVPNPeers(iface.id)
      peers.value = data
    } catch (e: any) { ElMessage.error(apiErr(e, '获取 Peer 列表失败')) }
  }
}

const handleCreateIface = async () => {
  submitting.value = true
  try {
    await createVPNInterface({ ...ifaceForm, device_id: deviceId.value, enabled: true })
    ElMessage.success('接口已创建')
    showIfaceDialog.value = false
    Object.assign(ifaceForm, { name: 'wg0', address: '10.99.0.1/24', listen_port: 51820, private_key: '', public_key: '' })
    fetchInterfaces()
  } catch (e: any) { ElMessage.error(apiErr(e, '创建接口失败')) }
  finally { submitting.value = false }
}

const handleDeleteIface = async (iface: any) => {
  await ElMessageBox.confirm(`删除接口 "${iface.name}" 及其所有 Peers？`, '确认', { type: 'warning' })
  try {
    await deleteVPNInterface(iface.id)
    ElMessage.success('已删除')
    fetchInterfaces()
  } catch (e: any) { ElMessage.error(apiErr(e, '删除失败')) }
}

const handleCreatePeer = async () => {
  submitting.value = true
  try {
    await createVPNPeer({ ...peerForm, interface_id: selectedIface.value.id, enabled: true })
    ElMessage.success('Peer 已添加')
    showPeerDialog.value = false
    Object.assign(peerForm, { description: '', public_key: '', allowed_ips: '', endpoint: '', keepalive: 25 })
    selectInterface(selectedIface.value)
  } catch (e: any) { ElMessage.error(apiErr(e, '添加 Peer 失败')) }
  finally { submitting.value = false }
}

const handleDeletePeer = async (peer: any) => {
  await ElMessageBox.confirm('删除此 Peer？', '确认', { type: 'warning' })
  try {
    await deleteVPNPeer(peer.id)
    ElMessage.success('已删除')
    selectInterface(selectedIface.value)
  } catch (e: any) { ElMessage.error(apiErr(e, '删除失败')) }
}

const handleApply = async () => {
  if (!deviceId.value) return
  await ElMessageBox.confirm('确认将 VPN 配置应用到设备？', '应用确认', { type: 'warning' })
  applying.value = true
  try {
    await applyVPN(deviceId.value)
    ElMessage.success('VPN 配置已下发')
  } catch (e: any) { ElMessage.error(apiErr(e, '配置下发失败')) }
  finally { applying.value = false }
}

onMounted(async () => {
  try {
    const { data } = await getDevices()
    devices.value = data
  } catch (e: any) { ElMessage.error(apiErr(e, '获取设备列表失败')) }
})
</script>
