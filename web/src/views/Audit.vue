<template>
  <div>
    <!-- Filters -->
    <el-card style="margin-bottom: 16px">
      <el-row :gutter="16" align="middle">
        <el-col :span="4">
          <el-select v-model="filter.action" placeholder="操作类型" clearable style="width: 100%" @change="fetchLogs">
            <el-option label="登录" value="login" />
            <el-option label="创建" value="create" />
            <el-option label="更新" value="update" />
            <el-option label="删除" value="delete" />
            <el-option label="重启" value="reboot" />
            <el-option label="应用" value="apply" />
            <el-option label="上传" value="upload" />
            <el-option label="升级" value="upgrade" />
          </el-select>
        </el-col>
        <el-col :span="4">
          <el-select v-model="filter.resource" placeholder="资源类型" clearable style="width: 100%" @change="fetchLogs">
            <el-option label="设备" value="device" />
            <el-option label="配置" value="config" />
            <el-option label="模板" value="template" />
            <el-option label="固件" value="firmware" />
            <el-option label="防火墙" value="firewall" />
            <el-option label="VPN" value="vpn" />
            <el-option label="MWAN" value="mwan" />
            <el-option label="DHCP" value="dhcp" />
            <el-option label="VLAN" value="vlan" />
          </el-select>
        </el-col>
        <el-col :span="4">
          <el-input v-model="filter.username" placeholder="用户名" clearable @clear="fetchLogs" @keyup.enter="fetchLogs" />
        </el-col>
        <el-col :span="8">
          <el-date-picker
            v-model="dateRange"
            type="datetimerange"
            range-separator="至"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            style="width: 100%"
            @change="fetchLogs"
          />
        </el-col>
        <el-col :span="2">
          <el-button @click="resetFilters">重置</el-button>
        </el-col>
      </el-row>
    </el-card>

    <el-table :data="logs" v-loading="loading" stripe>
      <el-table-column prop="created_at" label="时间" width="180">
        <template #default="{ row }">{{ new Date(row.created_at).toLocaleString() }}</template>
      </el-table-column>
      <el-table-column prop="username" label="用户" width="120" />
      <el-table-column prop="action" label="操作" width="100">
        <template #default="{ row }">
          <el-tag :type="actionTagType(row.action)" size="small">{{ row.action }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="resource" label="资源" width="120" />
      <el-table-column prop="detail" label="详情" />
      <el-table-column prop="ip" label="IP" width="150" />
    </el-table>

    <el-row justify="end" style="margin-top: 16px">
      <el-pagination
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :page-sizes="[20, 50, 100]"
        :total="total"
        layout="total, sizes, prev, pager, next"
        @size-change="fetchLogs"
        @current-change="fetchLogs"
      />
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getAuditLogs, apiErr } from '../api'

const logs = ref<any[]>([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(50)
const total = ref(0)
const dateRange = ref<[Date, Date] | null>(null)
const filter = reactive({ action: '', resource: '', username: '' })

const actionTagType = (action: string) => {
  if (action === 'login') return 'primary'
  if (action === 'create' || action === 'upload') return 'success'
  if (action === 'delete' || action === 'reboot') return 'danger'
  return 'warning'
}

const fetchLogs = async () => {
  loading.value = true
  try {
    const params: Record<string, string | number> = { page: page.value, page_size: pageSize.value }
    if (filter.action) params.action = filter.action
    if (filter.resource) params.resource = filter.resource
    if (filter.username) params.username = filter.username
    if (dateRange.value) {
      params.from = dateRange.value[0].toISOString()
      params.to = dateRange.value[1].toISOString()
    }
    const { data } = await getAuditLogs(params)
    if (data.data) {
      logs.value = data.data
      total.value = data.total
    } else {
      logs.value = data
      total.value = data.length
    }
  } catch (e: any) {
    ElMessage.error(apiErr(e, '获取审计日志失败'))
  } finally {
    loading.value = false
  }
}

const resetFilters = () => {
  filter.action = ''
  filter.resource = ''
  filter.username = ''
  dateRange.value = null
  page.value = 1
  fetchLogs()
}

onMounted(fetchLogs)
</script>
