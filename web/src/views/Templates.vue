<template>
  <div>
    <el-row justify="space-between" style="margin-bottom: 16px">
      <el-radio-group v-model="category" @change="fetchTemplates">
        <el-radio-button value="">全部</el-radio-button>
        <el-radio-button value="network">网络</el-radio-button>
        <el-radio-button value="firewall">防火墙</el-radio-button>
        <el-radio-button value="vpn">VPN</el-radio-button>
        <el-radio-button value="qos">QoS</el-radio-button>
      </el-radio-group>
      <el-button type="primary" @click="showCreate = true">新建模板</el-button>
    </el-row>

    <el-table :data="templates" v-loading="loading" stripe>
      <el-table-column prop="name" label="模板名称" />
      <el-table-column prop="category" label="分类" width="120" />
      <el-table-column prop="description" label="描述" />
      <el-table-column prop="version" label="版本" width="80" />
      <el-table-column prop="updated_at" label="更新时间" width="180" />
      <el-table-column label="操作" width="180" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="editTemplate(row)">编辑</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="showCreate" :title="editingId ? '编辑模板' : '新建模板'" width="700">
      <el-form :model="form" label-width="80px">
        <el-form-item label="名称">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="分类">
          <el-select v-model="form.category">
            <el-option label="网络" value="network" />
            <el-option label="防火墙" value="firewall" />
            <el-option label="VPN" value="vpn" />
            <el-option label="QoS" value="qos" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" />
        </el-form-item>
        <el-form-item label="配置内容">
          <el-input v-model="form.content" type="textarea" :rows="12" placeholder="UCI 配置内容" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreate = false">取消</el-button>
        <el-button type="primary" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getTemplates, createTemplate, updateTemplate, deleteTemplate } from '../api'

const templates = ref<any[]>([])
const loading = ref(false)
const showCreate = ref(false)
const category = ref('')
const editingId = ref<number | null>(null)
const form = reactive({ name: '', category: '', description: '', content: '' })

const fetchTemplates = async () => {
  loading.value = true
  try {
    const { data } = await getTemplates(category.value || undefined)
    templates.value = data
  } finally {
    loading.value = false
  }
}

const editTemplate = (tpl: any) => {
  editingId.value = tpl.id
  Object.assign(form, { name: tpl.name, category: tpl.category, description: tpl.description, content: tpl.content })
  showCreate.value = true
}

const handleSave = async () => {
  if (editingId.value) {
    await updateTemplate(editingId.value, { ...form })
  } else {
    await createTemplate({ ...form })
  }
  ElMessage.success('已保存')
  showCreate.value = false
  editingId.value = null
  fetchTemplates()
}

const handleDelete = async (tpl: any) => {
  await ElMessageBox.confirm(`删除模板 "${tpl.name}"？`, '确认', { type: 'warning' })
  await deleteTemplate(tpl.id)
  ElMessage.success('已删除')
  fetchTemplates()
}

onMounted(fetchTemplates)
</script>
