<template>
  <div v-loading="loading">
    <el-page-header @back="$router.push('/devices')" :content="device?.name || ''" />

    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="12">
        <el-card>
          <template #header>设备信息</template>
          <el-descriptions :column="1" border>
            <el-descriptions-item label="名称">{{ device?.name }}</el-descriptions-item>
            <el-descriptions-item label="MAC">{{ device?.mac }}</el-descriptions-item>
            <el-descriptions-item label="IP 地址">{{ device?.ip_address }}</el-descriptions-item>
            <el-descriptions-item label="型号">{{ device?.model }}</el-descriptions-item>
            <el-descriptions-item label="固件">{{ device?.firmware }}</el-descriptions-item>
            <el-descriptions-item label="分组">{{ device?.group || '-' }}</el-descriptions-item>
            <el-descriptions-item label="状态">
              <el-tag :type="device?.status === 'online' ? 'success' : 'danger'" size="small">
                {{ device?.status === 'online' ? '在线' : '离线' }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="运行时间">{{ formatUptime(device?.uptime_secs) }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>

      <el-col :span="12">
        <el-card>
          <template #header>实时状态</template>
          <el-descriptions :column="1" border>
            <el-descriptions-item label="CPU 使用率">
              <el-progress :percentage="device?.cpu_usage || 0" :stroke-width="16" striped />
            </el-descriptions-item>
            <el-descriptions-item label="内存使用率">
              <el-progress :percentage="device?.mem_usage || 0" :stroke-width="16" striped />
            </el-descriptions-item>
          </el-descriptions>
        </el-card>

        <el-card style="margin-top: 20px">
          <template #header>操作</template>
          <el-space>
            <el-button type="warning" @click="handleReboot">重启设备</el-button>
            <el-button type="primary" @click="showPushConfig = true">下发配置</el-button>
          </el-space>
        </el-card>
      </el-col>
    </el-row>

    <!-- Config push dialog -->
    <el-dialog v-model="showPushConfig" title="下发配置" width="600">
      <el-form label-width="80px">
        <el-form-item label="配置模板">
          <el-select v-model="configForm.template_id" placeholder="选择模板（可选）" clearable style="width: 100%">
            <el-option v-for="t in templates" :key="t.id" :label="t.name" :value="t.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="UCI 配置">
          <el-input v-model="configForm.content" type="textarea" :rows="10" placeholder="直接输入 UCI 配置内容" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showPushConfig = false">取消</el-button>
        <el-button type="primary" @click="handlePushConfig">下发</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getDevice, rebootDevice, pushConfig, getTemplates } from '../api'

const route = useRoute()
const router = useRouter()
const deviceId = Number(route.params.id)

const device = ref<any>(null)
const templates = ref<any[]>([])
const loading = ref(false)
const showPushConfig = ref(false)
const configForm = reactive({ template_id: null as number | null, content: '' })

const formatUptime = (secs?: number) => {
  if (!secs) return '-'
  const d = Math.floor(secs / 86400)
  const h = Math.floor((secs % 86400) / 3600)
  const m = Math.floor((secs % 3600) / 60)
  return `${d}天 ${h}小时 ${m}分钟`
}

const handleReboot = async () => {
  await ElMessageBox.confirm('确认重启该设备？', '重启确认', { type: 'warning' })
  await rebootDevice(deviceId)
  ElMessage.success('重启指令已发送')
}

const handlePushConfig = async () => {
  await pushConfig(deviceId, {
    template_id: configForm.template_id ?? undefined,
    content: configForm.content || undefined,
  })
  ElMessage.success('配置已下发')
  showPushConfig.value = false
}

onMounted(async () => {
  loading.value = true
  try {
    const [devRes, tplRes] = await Promise.all([getDevice(deviceId), getTemplates()])
    device.value = devRes.data
    templates.value = tplRes.data
  } catch {
    ElMessage.error('获取设备信息失败')
    router.push('/devices')
  } finally {
    loading.value = false
  }
})
</script>
