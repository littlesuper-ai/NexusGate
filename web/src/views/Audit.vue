<template>
  <div>
    <el-table :data="logs" v-loading="loading" stripe>
      <el-table-column prop="created_at" label="时间" width="180" />
      <el-table-column prop="username" label="用户" width="120" />
      <el-table-column prop="action" label="操作" width="150" />
      <el-table-column prop="resource" label="资源" width="200" />
      <el-table-column prop="detail" label="详情" />
      <el-table-column prop="ip" label="IP" width="150" />
    </el-table>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getAuditLogs } from '../api'

const logs = ref<any[]>([])
const loading = ref(false)

onMounted(async () => {
  loading.value = true
  try {
    const { data } = await getAuditLogs()
    logs.value = data
  } finally {
    loading.value = false
  }
})
</script>
