<template>
  <div>
    <el-row justify="space-between" align="middle" style="margin-bottom: 16px">
      <el-col :span="12">
        <el-radio-group v-model="targetFilter" @change="fetchFirmwares">
          <el-radio-button value="">全部</el-radio-button>
          <el-radio-button value="x86-64">x86-64</el-radio-button>
          <el-radio-button value="nanopi-r4s">NanoPi R4S</el-radio-button>
          <el-radio-button value="nanopi-r5s">NanoPi R5S</el-radio-button>
        </el-radio-group>
      </el-col>
      <el-col :span="12" style="text-align: right">
        <el-button type="primary" @click="showUpload = true">上传固件</el-button>
        <el-button type="warning" @click="showBatch = true">批量升级</el-button>
      </el-col>
    </el-row>

    <el-table :data="firmwares" v-loading="loading" stripe>
      <el-table-column prop="version" label="版本" width="120" />
      <el-table-column prop="target" label="目标平台" width="130" />
      <el-table-column prop="filename" label="文件名" />
      <el-table-column label="大小" width="100">
        <template #default="{ row }">{{ (row.file_size / 1048576).toFixed(1) }} MB</template>
      </el-table-column>
      <el-table-column prop="sha256" label="SHA256" width="180">
        <template #default="{ row }">
          <span style="font-family: monospace; font-size: 11px">{{ row.sha256?.substring(0, 16) }}...</span>
        </template>
      </el-table-column>
      <el-table-column prop="is_stable" label="稳定版" width="80">
        <template #default="{ row }">
          <el-tag v-if="row.is_stable" type="success" size="small">稳定</el-tag>
          <el-tag v-else type="info" size="small">测试</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="上传时间" width="170" />
      <el-table-column label="操作" width="260" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="handleUpgradeOne(row)">推送升级</el-button>
          <el-button size="small" type="success" v-if="!row.is_stable" @click="handleMarkStable(row)">标记稳定</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- Upload Dialog -->
    <el-dialog v-model="showUpload" title="上传固件" width="500">
      <el-form :model="uploadForm" label-width="100px">
        <el-form-item label="固件文件">
          <el-upload
            ref="uploadRef"
            :auto-upload="false"
            :limit="1"
            :on-change="handleFileChange"
            accept=".img,.img.gz,.bin,.tar,.tar.gz"
          >
            <el-button type="primary">选择文件</el-button>
          </el-upload>
        </el-form-item>
        <el-form-item label="版本号"><el-input v-model="uploadForm.version" placeholder="例：23.05.5-nexusgate-1" /></el-form-item>
        <el-form-item label="目标平台">
          <el-select v-model="uploadForm.target" style="width: 100%">
            <el-option label="x86-64" value="x86-64" />
            <el-option label="NanoPi R4S" value="nanopi-r4s" />
            <el-option label="NanoPi R5S" value="nanopi-r5s" />
          </el-select>
        </el-form-item>
        <el-form-item label="变更日志"><el-input v-model="uploadForm.changelog" type="textarea" :rows="3" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showUpload = false">取消</el-button>
        <el-button type="primary" :loading="uploading" @click="handleUpload">上传</el-button>
      </template>
    </el-dialog>

    <!-- Push upgrade to one device -->
    <el-dialog v-model="showPushOne" title="推送固件升级" width="450">
      <el-form label-width="100px">
        <el-form-item label="固件">{{ pushTarget?.version }} ({{ pushTarget?.target }})</el-form-item>
        <el-form-item label="目标设备">
          <el-select v-model="pushDeviceId" placeholder="选择设备" style="width: 100%">
            <el-option v-for="d in devices" :key="d.id" :label="`${d.name} (${d.ip_address})`" :value="d.id" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showPushOne = false">取消</el-button>
        <el-button type="warning" @click="confirmPushOne">确认推送</el-button>
      </template>
    </el-dialog>

    <!-- Batch upgrade -->
    <el-dialog v-model="showBatch" title="批量固件升级" width="500">
      <el-form :model="batchForm" label-width="100px">
        <el-form-item label="固件版本">
          <el-select v-model="batchForm.firmware_id" placeholder="选择固件" style="width: 100%">
            <el-option v-for="f in firmwares" :key="f.id" :label="`${f.version} (${f.target})`" :value="f.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="按分组"><el-input v-model="batchForm.group" placeholder="可选，例：上海分部" /></el-form-item>
        <el-form-item label="按型号"><el-input v-model="batchForm.model" placeholder="可选，例：NanoPi R4S" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showBatch = false">取消</el-button>
        <el-button type="warning" @click="confirmBatch">确认批量升级</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getFirmwares, uploadFirmware, deleteFirmware, markFirmwareStable,
  pushFirmwareUpgrade, batchFirmwareUpgrade, getDevices,
} from '../api'

const firmwares = ref<any[]>([])
const devices = ref<any[]>([])
const loading = ref(false)
const uploading = ref(false)
const targetFilter = ref('')
const showUpload = ref(false)
const showPushOne = ref(false)
const showBatch = ref(false)
const pushTarget = ref<any>(null)
const pushDeviceId = ref<number | null>(null)

const uploadForm = reactive({ version: '', target: 'x86-64', changelog: '' })
let selectedFile: File | null = null

const batchForm = reactive({ firmware_id: null as number | null, group: '', model: '' })

const fetchFirmwares = async () => {
  loading.value = true
  try {
    const { data } = await getFirmwares(targetFilter.value || undefined)
    firmwares.value = data
  } finally {
    loading.value = false
  }
}

const handleFileChange = (file: any) => { selectedFile = file.raw }

const handleUpload = async () => {
  if (!selectedFile) { ElMessage.warning('请选择固件文件'); return }
  uploading.value = true
  try {
    const fd = new FormData()
    fd.append('file', selectedFile)
    fd.append('version', uploadForm.version)
    fd.append('target', uploadForm.target)
    fd.append('changelog', uploadForm.changelog)
    await uploadFirmware(fd)
    ElMessage.success('固件已上传')
    showUpload.value = false
    selectedFile = null
    fetchFirmwares()
  } catch { ElMessage.error('上传失败') }
  finally { uploading.value = false }
}

const handleDelete = async (fw: any) => {
  await ElMessageBox.confirm(`删除固件 ${fw.version}？`, '确认', { type: 'warning' })
  await deleteFirmware(fw.id)
  ElMessage.success('已删除')
  fetchFirmwares()
}

const handleMarkStable = async (fw: any) => {
  await markFirmwareStable(fw.id)
  ElMessage.success('已标记为稳定版')
  fetchFirmwares()
}

const handleUpgradeOne = (fw: any) => {
  pushTarget.value = fw
  pushDeviceId.value = null
  showPushOne.value = true
}

const confirmPushOne = async () => {
  if (!pushDeviceId.value || !pushTarget.value) return
  await pushFirmwareUpgrade({ device_id: pushDeviceId.value, firmware_id: pushTarget.value.id })
  ElMessage.success('升级指令已发送')
  showPushOne.value = false
}

const confirmBatch = async () => {
  if (!batchForm.firmware_id) { ElMessage.warning('请选择固件'); return }
  await ElMessageBox.confirm('确认批量推送固件升级？此操作将重启匹配的所有在线设备', '批量升级', { type: 'warning' })
  const { data } = await batchFirmwareUpgrade({
    firmware_id: batchForm.firmware_id,
    group: batchForm.group || undefined,
    model: batchForm.model || undefined,
  })
  ElMessage.success(data.message)
  showBatch.value = false
}

onMounted(async () => {
  await Promise.all([fetchFirmwares(), getDevices().then(r => { devices.value = r.data })])
})
</script>
