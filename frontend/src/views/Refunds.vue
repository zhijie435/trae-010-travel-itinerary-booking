<template>
  <div class="page-container">
    <div class="page-header">
      <h2>退款管理</h2>
      <p>管理所有退款申请，支持审核、查看详情等操作</p>
    </div>

    <div class="card-content">
      <div class="toolbar">
        <div class="search-filters">
          <el-select v-model="filterStatus" placeholder="审核状态" clearable style="width: 160px;">
            <el-option label="待审核" value="pending" />
            <el-option label="已通过" value="approved" />
            <el-option label="已拒绝" value="rejected" />
          </el-select>
          <el-button type="primary" @click="loadRefunds">
            <el-icon><Search /></el-icon>
            <span>查询</span>
          </el-button>
        </div>
        <el-button type="success" @click="loadRefunds">
          <el-icon><Refresh /></el-icon>
          <span>刷新</span>
        </el-button>
      </div>

      <el-table :data="refunds" style="width: 100%" v-loading="loading" stripe>
        <el-table-column prop="refund_no" label="退款编号" width="200" />
        <el-table-column prop="order_no" label="关联订单" width="200" />
        <el-table-column label="行程名称" min-width="160">
          <template #default="{ row }">
            {{ row.order?.trip_name || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="退款金额" width="120">
          <template #default="{ row }">
            <span style="color: #409EFF; font-weight: 600;">¥{{ row.refund_amount?.toFixed(2) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="退款原因" width="120">
          <template #default="{ row }">
            {{ reasonText(row.reason) }}
          </template>
        </el-table-column>
        <el-table-column label="审核状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)" effect="light">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="申请时间" width="170">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="260" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="viewDetail(row)">
              <el-icon><View /></el-icon>
              详情
            </el-button>
            <el-button
              v-if="row.status === 'pending'"
              type="success" link size="small"
              @click="openReviewDialog(row, 'approved')"
            >
              <el-icon><CircleCheck /></el-icon>
              通过
            </el-button>
            <el-button
              v-if="row.status === 'pending'"
              type="danger" link size="small"
              @click="openReviewDialog(row, 'rejected')"
            >
              <el-icon><CircleClose /></el-icon>
              拒绝
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="!loading && refunds.length === 0" description="暂无退款申请" />
    </div>

    <el-dialog v-model="detailVisible" title="退款申请详情" width="600px">
      <el-descriptions :column="2" border v-if="currentRefund">
        <el-descriptions-item label="退款编号">{{ currentRefund.refund_no }}</el-descriptions-item>
        <el-descriptions-item label="审核状态">
          <el-tag :type="statusType(currentRefund.status)" effect="light">{{ statusText(currentRefund.status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="关联订单" :span="2">
          {{ currentRefund.order_no }}
          <span style="margin-left: 12px; color: #909399;">
            {{ currentRefund.order?.trip_name || '-' }}
          </span>
        </el-descriptions-item>
        <el-descriptions-item label="退款金额">
          <span style="color: #409EFF; font-weight: 600;">¥{{ currentRefund.refund_amount?.toFixed(2) }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="退款原因">{{ reasonText(currentRefund.reason) }}</el-descriptions-item>
        <el-descriptions-item label="详细说明" :span="2">
          <span style="white-space: pre-wrap;">{{ currentRefund.description || '-' }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="申请时间" :span="2">{{ formatDate(currentRefund.created_at) }}</el-descriptions-item>
        <template v-if="currentRefund.status !== 'pending'">
          <el-descriptions-item label="审核时间">{{ currentRefund.review_time ? formatDate(currentRefund.review_time) : '-' }}</el-descriptions-item>
          <el-descriptions-item label="审核人">{{ currentRefund.reviewer_id ? '管理员' : '-' }}</el-descriptions-item>
          <el-descriptions-item label="审核备注" :span="2">
            {{ currentRefund.review_remark || '-' }}
          </el-descriptions-item>
        </template>
      </el-descriptions>
    </el-dialog>

    <el-dialog v-model="reviewVisible" :title="reviewTitle" width="500px">
      <el-alert
        :title="reviewType === 'approved' ? '通过后将按退款金额退款至用户账户' : '拒绝后订单将恢复为已支付状态'"
        :type="reviewType === 'approved' ? 'success' : 'warning'"
        :closable="false"
        style="margin-bottom: 20px;"
      />
      <el-form :model="reviewForm" label-width="100px">
        <el-form-item label="退款编号">
          <span>{{ currentReviewRefund?.refund_no }}</span>
        </el-form-item>
        <el-form-item label="退款金额">
          <span style="color: #409EFF; font-weight: 600;">¥{{ currentReviewRefund?.refund_amount?.toFixed(2) }}</span>
        </el-form-item>
        <el-form-item label="审核备注">
          <el-input
            v-model="reviewForm.remark"
            type="textarea"
            :rows="3"
            :placeholder="reviewType === 'approved' ? '请输入审核通过备注（选填）' : '请输入拒绝原因（建议填写）'"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="reviewVisible = false">取消</el-button>
        <el-button
          :type="reviewType === 'approved' ? 'success' : 'danger'"
          @click="submitReview"
          :loading="submitting"
        >
          确认{{ reviewType === 'approved' ? '通过' : '拒绝' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getRefundRequests, getRefundRequest, reviewRefundRequest } from '@/api/refund'

const loading = ref(false)
const refunds = ref([])
const filterStatus = ref('')

const detailVisible = ref(false)
const currentRefund = ref(null)

const reviewVisible = ref(false)
const currentReviewRefund = ref(null)
const reviewType = ref('')
const submitting = ref(false)
const reviewForm = reactive({
  remark: ''
})

const reviewTitle = computed(() => {
  return reviewType.value === 'approved' ? '审核通过' : '审核拒绝'
})

async function loadRefunds() {
  loading.value = true
  try {
    const params = {}
    if (filterStatus.value) params.status = filterStatus.value
    const res = await getRefundRequests(params)
    refunds.value = res.data || []
  } finally {
    loading.value = false
  }
}

function statusText(status) {
  const map = {
    pending: '待审核',
    approved: '已通过',
    rejected: '已拒绝'
  }
  return map[status] || status
}

function statusType(status) {
  const map = {
    pending: 'warning',
    approved: 'success',
    rejected: 'danger'
  }
  return map[status] || 'info'
}

function reasonText(reason) {
  const map = {
    change_of_plan: '行程变更',
    health_issue: '身体原因',
    financial: '资金问题',
    service_issue: '服务不满意',
    other: '其他原因'
  }
  return map[reason] || reason || '-'
}

function formatDate(date) {
  if (!date) return '-'
  return new Date(date).toLocaleString('zh-CN', { hour12: false }).replace(/\//g, '-')
}

async function viewDetail(row) {
  try {
    const res = await getRefundRequest(row.id)
    currentRefund.value = res.data
  } catch (e) {
    currentRefund.value = row
  }
  detailVisible.value = true
}

function openReviewDialog(row, type) {
  currentReviewRefund.value = row
  reviewType.value = type
  reviewForm.remark = ''
  reviewVisible.value = true
}

async function submitReview() {
  ElMessageBox.confirm(
    `确认${reviewType.value === 'approved' ? '通过' : '拒绝'}该退款申请？`,
    '审核确认',
    {
      confirmButtonText: '确认',
      cancelButtonText: '取消',
      type: reviewType.value === 'approved' ? 'success' : 'warning'
    }
  ).then(async () => {
    submitting.value = true
    try {
      await reviewRefundRequest(currentReviewRefund.value.id, {
        status: reviewType.value,
        review_remark: reviewForm.remark
      })
      ElMessage.success(reviewType.value === 'approved' ? '审核通过，已完成退款' : '已拒绝该申请')
      reviewVisible.value = false
      loadRefunds()
    } finally {
      submitting.value = false
    }
  }).catch(() => {})
}

onMounted(loadRefunds)
</script>
