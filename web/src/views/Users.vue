<template>
  <div>
    <el-row justify="end" style="margin-bottom: 16px">
      <el-button type="primary" @click="showCreate = true">新建用户</el-button>
    </el-row>

    <el-table :data="users" v-loading="loading" stripe>
      <template #empty><el-empty description="暂无用户" :image-size="60" /></template>
      <el-table-column prop="username" label="用户名" />
      <el-table-column prop="role" label="角色" width="120">
        <template #default="{ row }">
          <el-tag :type="row.role === 'admin' ? 'danger' : row.role === 'operator' ? '' : 'info'" size="small">
            {{ row.role }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="email" label="邮箱" />
      <el-table-column prop="created_at" label="创建时间" width="180" />
      <el-table-column label="操作" width="180">
        <template #default="{ row }">
          <el-button size="small" @click="openEdit(row)">编辑</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="showCreate" title="新建用户" width="450">
      <el-form :model="form" label-width="80px">
        <el-form-item label="用户名"><el-input v-model="form.username" /></el-form-item>
        <el-form-item label="密码"><el-input v-model="form.password" type="password" show-password /></el-form-item>
        <el-form-item label="角色">
          <el-select v-model="form.role">
            <el-option label="管理员" value="admin" />
            <el-option label="运维" value="operator" />
            <el-option label="只读" value="viewer" />
          </el-select>
        </el-form-item>
        <el-form-item label="邮箱"><el-input v-model="form.email" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreate = false">取消</el-button>
        <el-button type="primary" @click="handleCreate">创建</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showEdit" title="编辑用户" width="450">
      <el-form :model="editForm" label-width="80px">
        <el-form-item label="用户名">
          <el-input :model-value="editForm.username" disabled />
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="editForm.role">
            <el-option label="管理员" value="admin" />
            <el-option label="运维" value="operator" />
            <el-option label="只读" value="viewer" />
          </el-select>
        </el-form-item>
        <el-form-item label="邮箱"><el-input v-model="editForm.email" /></el-form-item>
        <el-form-item label="新密码">
          <el-input v-model="editForm.password" type="password" show-password placeholder="留空不修改" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showEdit = false">取消</el-button>
        <el-button type="primary" @click="handleEdit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getUsers, createUser, updateUser, deleteUser } from '../api'

const users = ref<any[]>([])
const loading = ref(false)
const showCreate = ref(false)
const showEdit = ref(false)
const form = reactive({ username: '', password: '', role: 'operator', email: '' })
const editForm = reactive({ id: 0, username: '', role: '', email: '', password: '' })

const fetchUsers = async () => {
  loading.value = true
  try {
    const { data } = await getUsers()
    users.value = data
  } finally {
    loading.value = false
  }
}

const handleCreate = async () => {
  await createUser({ ...form })
  ElMessage.success('用户已创建')
  showCreate.value = false
  fetchUsers()
}

const openEdit = (user: any) => {
  editForm.id = user.id
  editForm.username = user.username
  editForm.role = user.role
  editForm.email = user.email || ''
  editForm.password = ''
  showEdit.value = true
}

const handleEdit = async () => {
  const payload: Record<string, string> = {
    role: editForm.role,
    email: editForm.email,
  }
  if (editForm.password) {
    payload.password = editForm.password
  }
  await updateUser(editForm.id, payload)
  ElMessage.success('用户已更新')
  showEdit.value = false
  fetchUsers()
}

const handleDelete = async (user: any) => {
  await ElMessageBox.confirm(`删除用户 "${user.username}"？`, '确认', { type: 'warning' })
  await deleteUser(user.id)
  ElMessage.success('已删除')
  fetchUsers()
}

onMounted(fetchUsers)
</script>
