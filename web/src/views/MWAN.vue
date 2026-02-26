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
          <el-button type="success" :disabled="!selectedDevice" :loading="applying" @click="handleApply">应用到设备</el-button>
        </el-col>
      </el-row>
    </el-card>

    <el-tabs v-model="activeTab">
      <!-- WAN Interfaces -->
      <el-tab-pane label="WAN 接口" name="wan">
        <el-row justify="end" style="margin-bottom: 12px">
          <el-button type="primary" @click="openWanDialog()" :disabled="!selectedDevice">添加 WAN 接口</el-button>
        </el-row>
        <el-table :data="wans" stripe border size="small">
          <template #empty><el-empty description="暂无 WAN 接口" :image-size="60" /></template>
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
          <el-table-column label="操作" width="120">
            <template #default="{ row }">
              <el-button size="small" type="primary" link @click="openWanDialog(row)">编辑</el-button>
              <el-button size="small" type="danger" link @click="handleDeleteWan(row.id)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- Policies -->
      <el-tab-pane label="负载策略" name="policy">
        <el-row justify="end" style="margin-bottom: 12px">
          <el-button type="primary" @click="openPolicyDialog()" :disabled="!selectedDevice">添加策略</el-button>
        </el-row>
        <el-table :data="policies" stripe border size="small">
          <template #empty><el-empty description="暂无负载策略" :image-size="60" /></template>
          <el-table-column prop="id" label="ID" width="60" />
          <el-table-column prop="name" label="策略名称" width="150" />
          <el-table-column prop="members" label="成员配置" />
          <el-table-column prop="last_resort" label="兜底策略" width="120" />
          <el-table-column label="操作" width="120">
            <template #default="{ row }">
              <el-button size="small" type="primary" link @click="openPolicyDialog(row)">编辑</el-button>
              <el-button size="small" type="danger" link @click="handleDeletePolicy(row.id)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- Rules -->
      <el-tab-pane label="路由规则" name="rule">
        <el-row justify="end" style="margin-bottom: 12px">
          <el-button type="primary" @click="openRuleDialog()" :disabled="!selectedDevice">添加规则</el-button>
        </el-row>
        <el-table :data="rules" stripe border size="small">
          <template #empty><el-empty description="暂无路由规则" :image-size="60" /></template>
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
          <el-table-column label="操作" width="120">
            <template #default="{ row }">
              <el-button size="small" type="primary" link @click="openRuleDialog(row)">编辑</el-button>
              <el-button size="small" type="danger" link @click="handleDeleteRule(row.id)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>
    </el-tabs>

    <!-- WAN dialog -->
    <el-dialog v-model="showWanDialog" :title="editingWanId ? '编辑 WAN 接口' : '添加 WAN 接口'" width="500">
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
        <el-button type="primary" @click="submitWan" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>

    <!-- Policy dialog -->
    <el-dialog v-model="showPolicyDialog" :title="editingPolicyId ? '编辑负载策略' : '添加负载策略'" width="500">
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
        <el-button type="primary" @click="submitPolicy" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>

    <!-- Rule dialog -->
    <el-dialog v-model="showRuleDialog" :title="editingRuleId ? '编辑路由规则' : '添加路由规则'" width="500">
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
        <el-button type="primary" @click="submitRule" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getDevices, getWANInterfaces, createWANInterface, updateWANInterface, deleteWANInterface,
  getMWANPolicies, createMWANPolicy, updateMWANPolicy, deleteMWANPolicy,
  getMWANRules, createMWANRule, updateMWANRule, deleteMWANRule, applyMWAN, apiErr,
} from '../api'

const devices = ref<any[]>([])
const selectedDevice = ref<number | null>(null)
const activeTab = ref('wan')
const applying = ref(false)

const wans = ref<any[]>([])
const policies = ref<any[]>([])
const rules = ref<any[]>([])

const showWanDialog = ref(false)
const showPolicyDialog = ref(false)
const showRuleDialog = ref(false)
const editingWanId = ref<number | null>(null)
const editingPolicyId = ref<number | null>(null)
const editingRuleId = ref<number | null>(null)

const defaultWan = {
  name: '', interface: '', enabled: true, weight: 1,
  track_ips: '8.8.8.8,114.114.114.114', reliability: 2, interval: 5, down: 3, up: 3,
}
const defaultPolicy = { name: '', members: '', last_resort: 'default' }
const defaultRule = {
  name: '', src_ip: '', dest_ip: '', proto: 'all',
  src_port: '', dest_port: '', policy: '', enabled: true, position: 0,
}

const submitting = ref(false)
const wanForm = reactive({ ...defaultWan })
const policyForm = reactive({ ...defaultPolicy })
const ruleForm = reactive({ ...defaultRule })

const openWanDialog = (row?: any) => {
  if (row) {
    editingWanId.value = row.id
    Object.assign(wanForm, { name: row.name, interface: row.interface, enabled: row.enabled, weight: row.weight, track_ips: row.track_ips || '', reliability: row.reliability, interval: row.interval, down: row.down, up: row.up })
  } else {
    editingWanId.value = null
    Object.assign(wanForm, defaultWan)
  }
  showWanDialog.value = true
}
const openPolicyDialog = (row?: any) => {
  if (row) {
    editingPolicyId.value = row.id
    Object.assign(policyForm, { name: row.name, members: row.members || '', last_resort: row.last_resort })
  } else {
    editingPolicyId.value = null
    Object.assign(policyForm, defaultPolicy)
  }
  showPolicyDialog.value = true
}
const openRuleDialog = (row?: any) => {
  if (row) {
    editingRuleId.value = row.id
    Object.assign(ruleForm, { name: row.name, src_ip: row.src_ip || '', dest_ip: row.dest_ip || '', proto: row.proto || 'all', src_port: row.src_port || '', dest_port: row.dest_port || '', policy: row.policy, enabled: row.enabled, position: row.position })
  } else {
    editingRuleId.value = null
    Object.assign(ruleForm, defaultRule)
  }
  showRuleDialog.value = true
}

const fetchAll = async () => {
  if (!selectedDevice.value) { wans.value = []; policies.value = []; rules.value = []; return }
  try {
    const [w, p, r] = await Promise.all([
      getWANInterfaces(selectedDevice.value),
      getMWANPolicies(selectedDevice.value),
      getMWANRules(selectedDevice.value),
    ])
    wans.value = w.data
    policies.value = p.data
    rules.value = r.data
  } catch { ElMessage.error('获取 MWAN 数据失败') }
}

const submitWan = async () => {
  submitting.value = true
  try {
    if (editingWanId.value) {
      await updateWANInterface(editingWanId.value, { ...wanForm, device_id: selectedDevice.value })
      ElMessage.success('已更新')
    } else {
      await createWANInterface({ ...wanForm, device_id: selectedDevice.value })
      ElMessage.success('已添加')
    }
    showWanDialog.value = false; editingWanId.value = null
    Object.assign(wanForm, defaultWan); fetchAll()
  } catch (e: any) { ElMessage.error(apiErr(e, '保存 WAN 接口失败')) }
  finally { submitting.value = false }
}
const handleDeleteWan = async (id: number) => {
  await ElMessageBox.confirm('确认删除此 WAN 接口？', '确认')
  try {
    await deleteWANInterface(id); ElMessage.success('已删除'); fetchAll()
  } catch { ElMessage.error('删除 WAN 接口失败') }
}
const submitPolicy = async () => {
  submitting.value = true
  try {
    if (editingPolicyId.value) {
      await updateMWANPolicy(editingPolicyId.value, { ...policyForm, device_id: selectedDevice.value })
      ElMessage.success('已更新')
    } else {
      await createMWANPolicy({ ...policyForm, device_id: selectedDevice.value })
      ElMessage.success('已添加')
    }
    showPolicyDialog.value = false; editingPolicyId.value = null
    Object.assign(policyForm, defaultPolicy); fetchAll()
  } catch (e: any) { ElMessage.error(apiErr(e, '保存策略失败')) }
  finally { submitting.value = false }
}
const handleDeletePolicy = async (id: number) => {
  await ElMessageBox.confirm('确认删除此策略？', '确认')
  try {
    await deleteMWANPolicy(id); ElMessage.success('已删除'); fetchAll()
  } catch { ElMessage.error('删除策略失败') }
}
const submitRule = async () => {
  submitting.value = true
  try {
    if (editingRuleId.value) {
      await updateMWANRule(editingRuleId.value, { ...ruleForm, device_id: selectedDevice.value })
      ElMessage.success('已更新')
    } else {
      await createMWANRule({ ...ruleForm, device_id: selectedDevice.value })
      ElMessage.success('已添加')
    }
    showRuleDialog.value = false; editingRuleId.value = null
    Object.assign(ruleForm, defaultRule); fetchAll()
  } catch (e: any) { ElMessage.error(apiErr(e, '保存规则失败')) }
  finally { submitting.value = false }
}
const handleDeleteRule = async (id: number) => {
  await ElMessageBox.confirm('确认删除此规则？', '确认')
  try {
    await deleteMWANRule(id); ElMessage.success('已删除'); fetchAll()
  } catch { ElMessage.error('删除规则失败') }
}

const handleApply = async () => {
  if (!selectedDevice.value) return
  await ElMessageBox.confirm('确认应用 mwan3 配置到设备？', '确认应用', { type: 'warning' })
  applying.value = true
  try {
    const { data } = await applyMWAN(selectedDevice.value)
    ElMessage.success(data.message || '配置已推送')
  } catch (e: any) { ElMessage.error(apiErr(e, 'MWAN 配置下发失败')) }
  finally { applying.value = false }
}

onMounted(async () => {
  try {
    const { data } = await getDevices()
    devices.value = data
  } catch { ElMessage.error('获取设备列表失败') }
})
</script>
