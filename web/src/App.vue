<template>
  <el-config-provider :locale="zhCn">
    <router-view v-if="!needAuth" />
    <Layout v-else>
      <router-view />
    </Layout>
  </el-config-provider>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import Layout from './layout/Layout.vue'

const route = useRoute()
const needAuth = computed(() => route.meta?.requiresAuth !== false)
</script>
