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
          <el-button type="success" :disabled="!selectedDevice" @click="handleApply">应用到设备</el-button>
        </el-col>
      </el-row>
    </el-card>

    <el-tabs v-model="activeTab">
      <!-- WAN Interfaces -->
      <el-tab-pane label="WAN 接口" name="wan">
        <el-row justify="end" style="margin-bottom: 12px">
          <el-button type="primary" @click="showWanDialog = true" :disabled="!selectedDevice">添加 WAN 接口</el-button>
        </el-row>
        <el-table :data="wans" stripe border size="small">
          <el-table-column prop="id" label="ID" width="60" />
          <el-table-column prop="name" label="名称" width="100" />
          <el-table-column prop="interface" label="接口" width="120" />
          <el-table-column prop="enabled" label="启用" width="70">
            <template #default="{ row }">
              <el-tag :type="row.enabled ? 'success' : 'info'" size="small">{{ row.enabled ? '是' : '否' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="weight" label="权重" width="70" />
          <el-table-column prop="track_ips" label="探测 IP" />
          <el-table-column prop="reliability" label="可靠性" width="80" />
          <el-table-column prop="interval" label="间隔(s)" width="80" />
          <el-table-column prop="down" label="Down" width="60" />
          <el-table-column prop="up" label="Up" width="60" />
          <el-table-column label="操作" width="80">
            <template #default="{ row }">
              <el-button size="small" type="danger" link @click="handleDeleteWan(row.id)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- Policies -->
      <el-tab-pane label="负载策略" name="policy">
        <el-row justify="end" style="margin-bottom: 12px">
          <el-button type="primary" @click="showPolicyDialog = true" :disabled="!selectedDevice">添加策略</el-button>
        </el-row>
        <el-table :data="policies" stripe border size="small">
          <el-table-column prop="id" label="ID" width="60" />
          <el-table-column prop="name" label="策略名称" width="150" />
          <el-table-column prop="members" label="成员配置" />
          <el-table-column prop="last_resort" label="兜底策略" width="120" />
          <el-table-column label="操作" width="80">
            <template #default="{ row }">
              <el-button size="small" type="danger" link @click="handleDeletePolicy(row.id)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- Rules -->
      <el-tab-pane label="路由规则" name="rule">
        <el-row justify="end" style="margin-bottom: 12px">
          <el-button type="primary" @click="showRuleDialog = true" :disabled="!selectedDevice">添加规则</el-button>
        </el-row>
        <el-table :data="rules" stripe border size="small">
          <el-table-column prop="id" label="ID" width="60" />
          <el-table-column prop="name" label="规则名" width="120" />
          <el-table-column prop="src_ip" label="源 IP" width="130" />
          <el-table-column prop="dest_ip" label="目标 IP" width="130" />
          <el-table-column prop="proto" label="协议" width="70" />
          <el-table-column prop="dest_port" label="端口" width="80" />
          <el-table-column prop="policy" label="使用策略" width="120" />
          <el-table-column prop="enabled" label="启用" width="70">
            <template #default="{ row }">
              <el-tag :type="row.enabled ? 'success' : 'info'" size="small">{{ row.enabled ? '是' : '否' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="position" label="优先级" width="80" />
          <el-table-column label="操作" width="80">
            <template #default="{ row }">
              <el-button size="small" type="danger" link @click="handleDeleteRule(row.id)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>
    </el-tabs>

    <!-- Add WAN dialog -->
    <el-dialog v-model="showWanDialog" title="添加 WAN 接口" width="500">
      <el-form :model="wanForm" label-width="90px">
        <el-form-item label="名称"><el-input v-model="wanForm.name" placeholder="wan1" /></el-form-item>
        <el-form-item label="接口"><el-input v-model="wanForm.interface" placeholder="eth1 / pppoe-wan" /></el-form-item>
        <el-form-item label="启用"><el-switch v-model="wanForm.enabled" /></el-form-item>
        <el-form-item label="权重"><el-input-number v-model="wanForm.weight" :min="1" :max="100" /></el-form-item>
        <el-form-item label="探测 IP"><el-input v-model="wanForm.track_ips" placeholder="8.8.8.8,114.114.114.114" /></el-form-item>
        <el-form-item label="可靠性"><el-input-number v-model="wanForm.reliability" :min="1" :max="10" /></el-form-item>
        <el-form-item label="间隔(s)"><el-input-number v-model="wanForm.interval" :min="1" :max="60" /></el-form-item>
        <el-form-item label="Down 阈值"><el-input-number v-model="wanForm.down" :min="1" :max="20" /></el-form-item>
        <el-form-item label="Up 阈值"><el-input-number v-model="wanForm.up" :min="1" :max="20" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showWanDialog = false">取消</el-button>
        <el-button type="primary" @click="submitWan">确定</el-button>
      </template>
    </el-dialog>

    <!-- Add Policy dialog -->
    <el-dialog v-model="showPolicyDialog" title="添加负载策略" width="500">
      <el-form :model="policyForm" label-width="90px">
        <el-form-item label="策略名称"><el-input v-model="policyForm.name" placeholder="balanced" /></el-form-item>
        <el-form-item label="成员">
          <el-input v-model="policyForm.members" type="textarea" :rows="3" placeholder='[{"iface":"wan1","metric":1,"weight":1}]' />
        </el-form-item>
        <el-form-item label="兜底策略">
          <el-select v-model="policyForm.last_resort" style="width: 100%">
            <el-option label="default" value="default" />
            <el-option label="unreachable" value="unreachable" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showPolicyDialog = false">取消</el-button>
        <el-button type="primary" @click="submitPolicy">确定</el-button>
      </template>
    </el-dialog>

    <!-- Add Rule dialog -->
    <el-dialog v-model="showRuleDialog" title="添加路由规则" width="500">
      <el-form :model="ruleForm" label-width="90px">
        <el-form-item label="规则名称"><el-input v-model="ruleForm.name" placeholder="video_traffic" /></el-form-item>
        <el-form-item label="源 IP"><el-input v-model="ruleForm.src_ip" placeholder="0.0.0.0/0" /></el-form-item>
        <el-form-item label="目标 IP"><el-input v-model="ruleForm.dest_ip" placeholder="0.0.0.0/0" /></el-form-item>
        <el-form-item label="协议">
          <el-select v-model="ruleForm.proto" style="width: 100%">
            <el-option label="all" value="all" />
            <el-option label="tcp" value="tcp" />
            <el-option label="udp" value="udp" />
            <el-option label="icmp" value="icmp" />
          </el-select>
        </el-form-item>
        <el-form-item label="源端口"><el-input v-model="ruleForm.src_port" /></el-form-item>
        <el-form-item label="目标端口"><el-input v-model="ruleForm.dest_port" placeholder="80,443" /></el-form-item>
        <el-form-item label="使用策略">
          <el-select v-model="ruleForm.policy" style="width: 100%">
            <el-option v-for="p in policies" :key="p.id" :label="p.name" :value="p.name" />
          </el-select>
        </el-form-item>
        <el-form-item label="启用"><el-switch v-model="ruleForm.enabled" /></el-form-item>
        <el-form-item label="优先级"><el-input-number v-model="ruleForm.position" :min="0" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showRuleDialog = false">取消</el-button>
        <el-button type="primary" @click="submitRule">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getDevices, getWANInterfaces, createWANInterface, deleteWANInterface,
  getMWANPolicies, createMWANPolicy, deleteMWANPolicy,
  getMWANRules, createMWANRule, deleteMWANRule, applyMWAN,
} from '../api'

const devices = ref<any[]>([])
const selectedDevice = ref<number | null>(null)
const activeTab = ref('wan')

const wans = ref<any[]>([])
const policies = ref<any[]>([])
const rules = ref<any[]>([])

const showWanDialog = ref(false)
const showPolicyDialog = ref(false)
const showRuleDialog = ref(false)

const wanForm = reactive({
  name: '', interface: '', enabled: true, weight: 1,
  track_ips: '8.8.8.8,114.114.114.114', reliability: 2, interval: 5, down: 3, up: 3,
})
const policyForm = reactive({ name: '', members: '', last_resort: 'default' })
const ruleForm = reactive({
  name: '', src_ip: '', dest_ip: '', proto: 'all',
  src_port: '', dest_port: '', policy: '', enabled: true, position: 0,
})

const fetchAll = async () => {
  if (!selectedDevice.value) { wans.value = []; policies.value = []; rules.value = []; return }
  const [w, p, r] = await Promise.all([
    getWANInterfaces(selectedDevice.value),
    getMWANPolicies(selectedDevice.value),
    getMWANRules(selectedDevice.value),
  ])
  wans.value = w.data
  policies.value = p.data
  rules.value = r.data
}

const submitWan = async () => {
  await createWANInterface({ ...wanForm, device_id: selectedDevice.value })
  ElMessage.success('已添加'); showWanDialog.value = false
  Object.assign(wanForm, { name: '', interface: '', enabled: true, weight: 1, track_ips: '8.8.8.8,114.114.114.114', reliability: 2, interval: 5, down: 3, up: 3 })
  fetchAll()
}
const handleDeleteWan = async (id: number) => {
  await ElMessageBox.confirm('确认删除此 WAN 接口？', '确认')
  await deleteWANInterface(id); ElMessage.success('已删除'); fetchAll()
}
const submitPolicy = async () => {
  await createMWANPolicy({ ...policyForm, device_id: selectedDevice.value })
  ElMessage.success('已添加'); showPolicyDialog.value = false
  Object.assign(policyForm, { name: '', members: '', last_resort: 'default' })
  fetchAll()
}
const handleDeletePolicy = async (id: number) => {
  await ElMessageBox.confirm('确认删除此策略？', '确认')
  await deleteMWANPolicy(id); ElMessage.success('已删除'); fetchAll()
}
const submitRule = async () => {
  await createMWANRule({ ...ruleForm, device_id: selectedDevice.value })
  ElMessage.success('已添加'); showRuleDialog.value = false
  Object.assign(ruleForm, { name: '', src_ip: '', dest_ip: '', proto: 'all', src_port: '', dest_port: '', policy: '', enabled: true, position: 0 })
  fetchAll()
}
const handleDeleteRule = async (id: number) => {
  await ElMessageBox.confirm('确认删除此规则？', '确认')
  await deleteMWANRule(id); ElMessage.success('已删除'); fetchAll()
}

const handleApply = async () => {
  if (!selectedDevice.value) return
  await ElMessageBox.confirm('确认应用 mwan3 配置到设备？', '确认应用', { type: 'warning' })
  const { data } = await applyMWAN(selectedDevice.value)
  ElMessage.success(data.message || '配置已推送')
}

onMounted(async () => {
  const { data } = await getDevices()
  devices.value = data
})
</script>
