<template>
  <div class="page-container">
    <div class="page-header">
      <h2>工作台</h2>
      <p>实时了解订单和退款数据概览</p>
    </div>

    <el-row :gutter="20" style="margin-bottom: 24px;">
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);">
              <el-icon :size="28" color="#fff"><Tickets /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-label">订单总数</div>
              <div class="stat-value">{{ stats.totalOrders }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);">
              <el-icon :size="28" color="#fff"><Money /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-label">订单金额</div>
              <div class="stat-value">¥{{ stats.totalAmount.toLocaleString() }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);">
              <el-icon :size="28" color="#fff"><Refund /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-label">退款申请</div>
              <div class="stat-value">{{ stats.totalRefunds }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background: linear-gradient(135deg, #fa709a 0%, #fee140 100%);">
              <el-icon :size="28" color="#fff"><Clock /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-label">待审核退款</div>
              <div class="stat-value">{{ stats.pendingRefunds }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20">
      <el-col :span="14">
        <el-card>
          <template #header>
            <div class="card-header">
              <span><strong>最新订单</strong></span>
              <el-button type="primary" link @click="$router.push('/orders')">查看全部</el-button>
            </div>
          </template>
          <el-table :data="recentOrders" style="width: 100%" v-loading="loading.orders">
            <el-table-column prop="order_no" label="订单编号" width="180" />
            <el-table-column prop="trip_name" label="行程名称" />
            <el-table-column label="金额" width="120">
              <template #default="{ row }">
                <span style="color: #f56c6c; font-weight: 600;">¥{{ row.total_amount?.toFixed(2) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="orderStatusType(row.status)" effect="light">{{ orderStatusText(row.status) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="创建时间" width="170">
              <template #default="{ row }">
                {{ formatDate(row.created_at) }}
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
      <el-col :span="10">
        <el-card>
          <template #header>
            <div class="card-header">
              <span><strong>退款申请列表</strong></span>
              <el-button type="primary" link @click="$router.push('/refunds')">查看全部</el-button>
            </div>
          </template>
          <el-table :data="recentRefunds" style="width: 100%" v-loading="loading.refunds">
            <el-table-column prop="refund_no" label="退款编号" width="180" />
            <el-table-column prop="order_no" label="关联订单" width="180" />
            <el-table-column label="金额" width="110">
              <template #default="{ row }">
                <span style="color: #409EFF; font-weight: 600;">¥{{ row.refund_amount?.toFixed(2) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="refundStatusType(row.status)" effect="light">{{ refundStatusText(row.status) }}</el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { getOrders } from '@/api/order'
import { getRefundRequests } from '@/api/refund'

const loading = reactive({
  orders: false,
  refunds: false
})

const stats = reactive({
  totalOrders: 0,
  totalAmount: 0,
  totalRefunds: 0,
  pendingRefunds: 0
})

const recentOrders = ref([])
const recentRefunds = ref([])

async function loadStats() {
  loading.orders = true
  loading.refunds = true
  try {
    const [ordersRes, refundsRes] = await Promise.all([
      getOrders(),
      getRefundRequests()
    ])

    const orders = ordersRes.data || []
    const refunds = refundsRes.data || []

    stats.totalOrders = orders.length
    stats.totalAmount = orders.reduce((sum, o) => sum + (o.total_amount || 0), 0)
    stats.totalRefunds = refunds.length
    stats.pendingRefunds = refunds.filter(r => r.status === 'pending').length

    recentOrders.value = orders.slice(0, 6)
    recentRefunds.value = refunds.slice(0, 6)
  } finally {
    loading.orders = false
    loading.refunds = false
  }
}

function orderStatusText(status) {
  const map = {
    pending: '待支付',
    paid: '已支付',
    confirmed: '已确认',
    refunding: '退款中',
    refunded: '已退款',
    completed: '已完成',
    cancelled: '已取消'
  }
  return map[status] || status
}

function orderStatusType(status) {
  const map = {
    pending: 'warning',
    paid: 'success',
    confirmed: 'success',
    refunding: 'warning',
    refunded: 'info',
    completed: 'success',
    cancelled: 'danger'
  }
  return map[status] || 'info'
}

function refundStatusText(status) {
  const map = {
    pending: '待审核',
    approved: '已通过',
    rejected: '已拒绝'
  }
  return map[status] || status
}

function refundStatusType(status) {
  const map = {
    pending: 'warning',
    approved: 'success',
    rejected: 'danger'
  }
  return map[status] || 'info'
}

function formatDate(date) {
  if (!date) return '-'
  return new Date(date).toLocaleString('zh-CN', { hour12: false }).replace(/\//g, '-')
}

onMounted(loadStats)
</script>

<style scoped>
.stat-card {
  border: none;
  border-radius: 12px;
}
.stat-content {
  display: flex;
  align-items: center;
  gap: 16px;
}
.stat-icon {
  width: 64px;
  height: 64px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.stat-info {
  flex: 1;
}
.stat-label {
  font-size: 13px;
  color: #909399;
  margin-bottom: 6px;
}
.stat-value {
  font-size: 24px;
  font-weight: 700;
  color: #303133;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
