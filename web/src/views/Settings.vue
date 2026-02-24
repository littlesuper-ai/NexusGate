<template>
  <div v-loading="loading">
    <el-tabs v-model="activeTab" @tab-change="loadCategory">
      <!-- General -->
      <el-tab-pane label="基本设置" name="general">
        <el-form label-width="160px" style="max-width: 600px">
          <el-form-item label="系统名称">
            <el-input v-model="form.system_name" placeholder="NexusGate" />
          </el-form-item>
          <el-form-item label="设备离线阈值(秒)">
            <el-input-number v-model.number="form.offline_threshold" :min="30" :max="3600" :step="30" />
          </el-form-item>
          <el-form-item label="指标保留天数">
            <el-input-number v-model.number="form.metrics_retention_days" :min="1" :max="365" />
          </el-form-item>
          <el-form-item label="默认分页大小">
            <el-input-number v-model.number="form.page_size" :min="10" :max="200" :step="10" />
          </el-form-item>
        </el-form>
      </el-tab-pane>

      <!-- MQTT -->
      <el-tab-pane label="MQTT 配置" name="mqtt">
        <el-form label-width="160px" style="max-width: 600px">
          <el-form-item label="Broker 地址">
            <el-input v-model="form.mqtt_broker" placeholder="tcp://localhost:1883" />
          </el-form-item>
          <el-form-item label="Client ID">
            <el-input v-model="form.mqtt_client_id" placeholder="nexusgate-server" />
          </el-form-item>
          <el-form-item label="Topic 前缀">
            <el-input v-model="form.mqtt_topic_prefix" placeholder="nexusgate" />
          </el-form-item>
          <el-form-item label="心跳间隔(秒)">
            <el-input-number v-model.number="form.mqtt_keepalive" :min="10" :max="300" />
          </el-form-item>
        </el-form>
      </el-tab-pane>

      <!-- Alert -->
      <el-tab-pane label="告警阈值" name="alert">
        <el-form label-width="160px" style="max-width: 600px">
          <el-form-item label="CPU 告警阈值(%)">
            <el-input-number v-model.number="form.alert_cpu_threshold" :min="50" :max="100" />
          </el-form-item>
          <el-form-item label="内存告警阈值(%)">
            <el-input-number v-model.number="form.alert_mem_threshold" :min="50" :max="100" />
          </el-form-item>
          <el-form-item label="连接数告警阈值">
            <el-input-number v-model.number="form.alert_conntrack_threshold" :min="1000" :max="500000" :step="1000" />
          </el-form-item>
          <el-form-item label="告警通知方式">
            <el-select v-model="form.alert_notify_method" style="width: 100%">
              <el-option label="仅记录日志" value="log" />
              <el-option label="Webhook" value="webhook" />
              <el-option label="邮件" value="email" />
            </el-select>
          </el-form-item>
          <el-form-item label="Webhook URL" v-if="form.alert_notify_method === 'webhook'">
            <el-input v-model="form.alert_webhook_url" placeholder="https://hooks.example.com/..." />
          </el-form-item>
        </el-form>
      </el-tab-pane>

      <!-- Firmware -->
      <el-tab-pane label="固件设置" name="firmware">
        <el-form label-width="160px" style="max-width: 600px">
          <el-form-item label="固件存储路径">
            <el-input v-model="form.firmware_store_path" placeholder="./firmware_store" />
          </el-form-item>
          <el-form-item label="最大固件大小(MB)">
            <el-input-number v-model.number="form.firmware_max_size_mb" :min="1" :max="500" />
          </el-form-item>
          <el-form-item label="自动升级">
            <el-switch v-model="form.firmware_auto_upgrade" />
          </el-form-item>
        </el-form>
      </el-tab-pane>
    </el-tabs>

    <el-divider />
    <el-row justify="end">
      <el-button type="primary" @click="saveSettings" :loading="saving">保存设置</el-button>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getSettings, batchUpsertSettings } from '../api'

const loading = ref(false)
const saving = ref(false)
const activeTab = ref('general')

const form = reactive<Record<string, any>>({
  // General
  system_name: 'NexusGate',
  offline_threshold: 120,
  metrics_retention_days: 30,
  page_size: 50,
  // MQTT
  mqtt_broker: 'tcp://localhost:1883',
  mqtt_client_id: 'nexusgate-server',
  mqtt_topic_prefix: 'nexusgate',
  mqtt_keepalive: 60,
  // Alert
  alert_cpu_threshold: 85,
  alert_mem_threshold: 85,
  alert_conntrack_threshold: 50000,
  alert_notify_method: 'log',
  alert_webhook_url: '',
  // Firmware
  firmware_store_path: './firmware_store',
  firmware_max_size_mb: 100,
  firmware_auto_upgrade: false,
})

const categoryMap: Record<string, string[]> = {
  general: ['system_name', 'offline_threshold', 'metrics_retention_days', 'page_size'],
  mqtt: ['mqtt_broker', 'mqtt_client_id', 'mqtt_topic_prefix', 'mqtt_keepalive'],
  alert: ['alert_cpu_threshold', 'alert_mem_threshold', 'alert_conntrack_threshold', 'alert_notify_method', 'alert_webhook_url'],
  firmware: ['firmware_store_path', 'firmware_max_size_mb', 'firmware_auto_upgrade'],
}

const loadCategory = async () => {
  // Data is already loaded on mount
}

const loadAll = async () => {
  loading.value = true
  try {
    const { data } = await getSettings()
    for (const item of data) {
      if (item.key in form) {
        // Convert types
        if (typeof form[item.key] === 'number') {
          form[item.key] = Number(item.value)
        } else if (typeof form[item.key] === 'boolean') {
          form[item.key] = item.value === 'true'
        } else {
          form[item.key] = item.value
        }
      }
    }
  } catch { /* use defaults */ }
  loading.value = false
}

const saveSettings = async () => {
  saving.value = true
  try {
    const items: { key: string; value: string; category: string }[] = []
    for (const [cat, keys] of Object.entries(categoryMap)) {
      for (const key of keys) {
        items.push({ key, value: String(form[key]), category: cat })
      }
    }
    await batchUpsertSettings(items)
    ElMessage.success('设置已保存')
  } catch {
    ElMessage.error('保存失败')
  }
  saving.value = false
}

onMounted(loadAll)
</script>
