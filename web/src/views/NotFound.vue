<template>
  <div style="display: flex; flex-direction: column; align-items: center; justify-content: center; height: 60vh">
    <h1 style="font-size: 72px; color: #909399; margin: 0">404</h1>
    <p style="font-size: 18px; color: #606266; margin: 16px 0">页面未找到</p>
    <el-button type="primary" @click="$router.push('/dashboard')">返回仪表板</el-button>
  </div>
</template>
