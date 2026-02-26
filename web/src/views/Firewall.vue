<template>
  <div>
    <el-row justify="space-between" align="middle" style="margin-bottom: 16px">
      <el-col :span="12">
        <span style="margin-right: 12px">选择设备：</span>
        <el-select v-model="deviceId" placeholder="选择设备" @change="fetchAll" style="width: 260px">
          <el-option v-for="d in devices" :key="d.id" :label="`${d.name} (${d.ip_address})`" :value="d.id" />
        </el-select>
      </el-col>
      <el-col :span="12" style="text-align: right">
        <el-button type="success" :disabled="!deviceId" @click="handleApply">应用到设备</el-button>
      </el-col>
    </el-row>

    <el-tabs>
      <!-- Zones -->
      <el-tab-pane label="区域 (Zones)">
        <el-row justify="end" style="margin-bottom: 12px">
          <el-button type="primary" size="small" :disabled="!deviceId" @click="showZoneDialog = true">新建区域</el-button>
        </el-row>
        <el-table :data="zones" stripe border>
          <el-table-column prop="name" label="区域名称" width="120" />
          <el-table-column prop="networks" label="关联网络" />
          <el-table-column prop="input" label="入站" width="100">
            <template #default="{ row }">
              <el-tag :type="row.input === 'ACCEPT' ? 'success' : row.input === 'DROP' ? 'danger' : 'warning'" size="small">{{ row.input }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="output" label="出站" width="100">
            <template #default="{ row }">
              <el-tag :type="row.output === 'ACCEPT' ? 'success' : 'warning'" size="small">{{ row.output }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="forward" label="转发" width="100">
            <template #default="{ row }">
              <el-tag :type="row.forward === 'ACCEPT' ? 'success' : 'warning'" size="small">{{ row.forward }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="masq" label="NAT" width="80">
            <template #default="{ row }">{{ row.masq ? '是' : '否' }}</template>
          </el-table-column>
          <el-table-column label="操作" width="140">
            <template #default="{ row }">
              <el-button size="small" type="danger" @click="handleDeleteZone(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- Rules -->
      <el-tab-pane label="规则 (Rules)">
        <el-row justify="end" style="margin-bottom: 12px">
          <el-button type="primary" size="small" :disabled="!deviceId" @click="showRuleDialog = true">新建规则</el-button>
        </el-row>
        <el-table :data="rules" stripe border>
          <el-table-column prop="position" label="#" width="50" />
          <el-table-column prop="name" label="规则名称" width="160" />
          <el-table-column prop="src" label="源区域" width="100" />
          <el-table-column prop="dest" label="目标区域" width="100" />
          <el-table-column prop="proto" label="协议" width="80" />
          <el-table-column prop="src_ip" label="源 IP" />
          <el-table-column prop="dest_port" label="目标端口" width="100" />
          <el-table-column prop="target" label="动作" width="100">
            <template #default="{ row }">
              <el-tag :type="row.target === 'ACCEPT' ? 'success' : row.target === 'DROP' ? 'danger' : 'warning'" size="small">{{ row.target }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="enabled" label="启用" width="80">
            <template #default="{ row }">
              <el-tag :type="row.enabled ? 'success' : 'info'" size="small">{{ row.enabled ? '是' : '否' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="140">
            <template #default="{ row }">
              <el-button size="small" type="danger" @click="handleDeleteRule(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>
    </el-tabs>

    <!-- Create Zone Dialog -->
    <el-dialog v-model="showZoneDialog" title="新建防火墙区域" width="500">
      <el-form :model="zoneForm" label-width="100px">
        <el-form-item label="区域名称"><el-input v-model="zoneForm.name" placeholder="例：lan, wan, dmz, guest" /></el-form-item>
        <el-form-item label="关联网络"><el-input v-model="zoneForm.networks" placeholder="逗号分隔，例：lan,guest" /></el-form-item>
        <el-form-item label="入站策略">
          <el-select v-model="zoneForm.input"><el-option label="ACCEPT" value="ACCEPT" /><el-option label="REJECT" value="REJECT" /><el-option label="DROP" value="DROP" /></el-select>
        </el-form-item>
        <el-form-item label="出站策略">
          <el-select v-model="zoneForm.output"><el-option label="ACCEPT" value="ACCEPT" /><el-option label="REJECT" value="REJECT" /><el-option label="DROP" value="DROP" /></el-select>
        </el-form-item>
        <el-form-item label="转发策略">
          <el-select v-model="zoneForm.forward"><el-option label="ACCEPT" value="ACCEPT" /><el-option label="REJECT" value="REJECT" /><el-option label="DROP" value="DROP" /></el-select>
        </el-form-item>
        <el-form-item label="NAT 伪装"><el-switch v-model="zoneForm.masq" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showZoneDialog = false">取消</el-button>
        <el-button type="primary" @click="handleCreateZone">创建</el-button>
      </template>
    </el-dialog>

    <!-- Create Rule Dialog -->
    <el-dialog v-model="showRuleDialog" title="新建防火墙规则" width="550">
      <el-form :model="ruleForm" label-width="100px">
        <el-form-item label="规则名称"><el-input v-model="ruleForm.name" placeholder="例：allow-ssh" /></el-form-item>
        <el-form-item label="源区域"><el-input v-model="ruleForm.src" placeholder="例：wan" /></el-form-item>
        <el-form-item label="目标区域"><el-input v-model="ruleForm.dest" placeholder="例：lan" /></el-form-item>
        <el-form-item label="协议">
          <el-select v-model="ruleForm.proto"><el-option label="TCP" value="tcp" /><el-option label="UDP" value="udp" /><el-option label="TCP+UDP" value="tcp udp" /><el-option label="ICMP" value="icmp" /><el-option label="任意" value="any" /></el-select>
        </el-form-item>
        <el-form-item label="源 IP"><el-input v-model="ruleForm.src_ip" placeholder="可选，例：192.168.1.0/24" /></el-form-item>
        <el-form-item label="目标 IP"><el-input v-model="ruleForm.dest_ip" placeholder="可选" /></el-form-item>
        <el-form-item label="目标端口"><el-input v-model="ruleForm.dest_port" placeholder="例：22 或 80,443 或 8000-9000" /></el-form-item>
        <el-form-item label="动作">
          <el-select v-model="ruleForm.target"><el-option label="ACCEPT" value="ACCEPT" /><el-option label="REJECT" value="REJECT" /><el-option label="DROP" value="DROP" /></el-select>
        </el-form-item>
        <el-form-item label="优先级"><el-input-number v-model="ruleForm.position" :min="0" :max="999" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showRuleDialog = false">取消</el-button>
        <el-button type="primary" @click="handleCreateRule">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getDevices, getFirewallZones, createFirewallZone, deleteFirewallZone,
  getFirewallRules, createFirewallRule, deleteFirewallRule, applyFirewall,
} from '../api'

const devices = ref<any[]>([])
const deviceId = ref<number | null>(null)
const zones = ref<any[]>([])
const rules = ref<any[]>([])
const showZoneDialog = ref(false)
const showRuleDialog = ref(false)

const zoneForm = reactive({ name: '', networks: '', input: 'REJECT', output: 'ACCEPT', forward: 'REJECT', masq: false })
const ruleForm = reactive({ name: '', src: '', dest: '', proto: 'tcp', src_ip: '', dest_ip: '', dest_port: '', target: 'ACCEPT', position: 0 })

const fetchAll = async () => {
  if (!deviceId.value) return
  try {
    const [z, r] = await Promise.all([getFirewallZones(deviceId.value), getFirewallRules(deviceId.value)])
    zones.value = z.data
    rules.value = r.data
  } catch { ElMessage.error('获取防火墙数据失败') }
}

const handleCreateZone = async () => {
  try {
    await createFirewallZone({ ...zoneForm, device_id: deviceId.value })
    ElMessage.success('区域已创建')
    showZoneDialog.value = false
    fetchAll()
  } catch { ElMessage.error('创建区域失败') }
}

const handleDeleteZone = async (zone: any) => {
  await ElMessageBox.confirm(`删除区域 "${zone.name}"？`, '确认', { type: 'warning' })
  try {
    await deleteFirewallZone(zone.id)
    ElMessage.success('已删除')
    fetchAll()
  } catch { ElMessage.error('删除区域失败') }
}

const handleCreateRule = async () => {
  try {
    await createFirewallRule({ ...ruleForm, device_id: deviceId.value, enabled: true })
    ElMessage.success('规则已创建')
    showRuleDialog.value = false
    fetchAll()
  } catch { ElMessage.error('创建规则失败') }
}

const handleDeleteRule = async (rule: any) => {
  await ElMessageBox.confirm(`删除规则 "${rule.name}"？`, '确认', { type: 'warning' })
  try {
    await deleteFirewallRule(rule.id)
    ElMessage.success('已删除')
    fetchAll()
  } catch { ElMessage.error('删除规则失败') }
}

const handleApply = async () => {
  if (!deviceId.value) return
  await ElMessageBox.confirm('确认将当前防火墙配置应用到设备？', '应用确认', { type: 'warning' })
  try {
    await applyFirewall(deviceId.value)
    ElMessage.success('防火墙配置已下发')
  } catch { ElMessage.error('防火墙配置下发失败') }
}

onMounted(async () => {
  try {
    const { data } = await getDevices()
    devices.value = data
  } catch { ElMessage.error('获取设备列表失败') }
})
</script>
