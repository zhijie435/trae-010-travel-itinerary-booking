<template>
  <el-container class="layout-container">
    <el-aside width="220px" class="sidebar">
      <div class="logo">
        <el-icon :size="24" color="#409EFF"><Tourism /></el-icon>
        <span>旅游售后系统</span>
      </div>
      <el-menu
        :default-active="activeMenu"
        router
        background-color="#001529"
        text-color="#b7bec9"
        active-text-color="#ffffff"
      >
        <el-menu-item index="/dashboard">
          <el-icon><Odometer /></el-icon>
          <span>工作台</span>
        </el-menu-item>
        <el-menu-item index="/trips">
          <el-icon><MapLocation /></el-icon>
          <span>行程管理</span>
        </el-menu-item>
        <el-menu-item index="/orders">
          <el-icon><Tickets /></el-icon>
          <span>订单管理</span>
        </el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header class="header">
        <div class="breadcrumb">
          <el-breadcrumb separator="/">
            <el-breadcrumb-item :to="{ path: '/dashboard' }">首页</el-breadcrumb-item>
            <el-breadcrumb-item>{{ $route.meta.title }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        <div class="user-info">
          <el-button type="primary" size="small" @click="handleSeed" v-if="showSeed">
            <el-icon><MagicStick /></el-icon>
            <span>初始化测试数据</span>
          </el-button>
          <el-dropdown>
            <span class="user-name">
              <el-avatar :size="32" class="avatar">{{ userStore.currentUser.username.charAt(0).toUpperCase() }}</el-avatar>
              <span class="name-text">{{ userStore.currentUser.username }}</span>
              <el-icon class="el-icon--right"><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item>个人中心</el-dropdown-item>
                <el-dropdown-item divided>退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>
      <el-main class="main-content">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { seedData } from '@/api/common'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const activeMenu = computed(() => route.path)
const showSeed = ref(false)

onMounted(() => {
  if (localStorage.getItem('seeded') !== 'true') {
    showSeed.value = true
  }
})

function handleSeed() {
  ElMessageBox.confirm(
    '将初始化用户、行程等测试数据，是否继续？',
    '提示',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'info'
    }
  ).then(async () => {
    await seedData()
    localStorage.setItem('seeded', 'true')
    showSeed.value = false
    ElMessage.success('测试数据初始化成功')
    router.go(0)
  }).catch(() => {})
}
</script>

<style scoped>
.layout-container {
  height: 100vh;
}
.sidebar {
  background: #001529;
  overflow-x: hidden;
}
.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  color: #fff;
  font-size: 16px;
  font-weight: 600;
  border-bottom: 1px solid rgba(255,255,255,0.1);
}
:deep(.el-menu) {
  border-right: none;
}
.header {
  background: #fff;
  border-bottom: 1px solid #e6e6e6;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
}
.user-info {
  display: flex;
  align-items: center;
  gap: 16px;
}
.user-name {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  user-select: none;
}
.avatar {
  background: #409EFF;
}
.name-text {
  font-size: 14px;
  color: #303133;
}
.main-content {
  padding: 0;
  background: #f5f7fa;
}
</style>
