<template>
  <div class="page-container">
    <div class="page-header">
      <h2>退款管理</h2>
      <p>管理所有退款申请，支持审核、批量审核、查看详情等操作</p>
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
        <div>
          <el-button type="success" @click="loadRefunds">
            <el-icon><Refresh /></el-icon>
            <span>刷新</span>
          </el-button>
          <el-button
            type="warning"
            @click="openBatchReview('approved')"
            :disabled="selectedRefunds.length === 0"
          >
            <el-icon><CircleCheck /></el-icon>
            <span>批量通过 ({{ selectedRefunds.length }})</span>
          </el-button>
          <el-button
            type="danger"
            @click="openBatchReview('rejected')"
            :disabled="selectedRefunds.length === 0"
          >
            <el-icon><CircleClose /></el-icon>
            <span>批量拒绝 ({{ selectedRefunds.length }})</span>
          </el-button>
        </div>
      </div>

      <el-table
        :data="refunds"
        style="width: 100%"
        v-loading="loading"
        stripe
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" align="center" />
        <el-table-column prop="refund_no" label="退款编号" width="200" />
        <el-table-column prop="order_no" label="关联订单" width="200" />
        <el-table-column label="行程名称" min-width="160">
          <template #default="{ row }">
            {{ row.order?.trip_name || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="退还人数" width="90" align="center">
          <template #default="{ row }">
            {{ row.order?.travelers || '-' }}
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
        <el-table-column label="操作" width="330" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="viewDetail(row)">
              <el-icon><View /></el-icon>
              详情
            </el-button>
            <el-button type="info" link size="small" @click="openReviewLogsDialog(row)">
              <el-icon><Clock /></el-icon>
              审核记录
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
        <el-descriptions-item label="退还人数">{{ currentRefund.order?.travelers || '-' }} 人</el-descriptions-item>
        <el-descriptions-item label="退款原因">{{ reasonText(currentRefund.reason) }}</el-descriptions-item>
        <el-descriptions-item label="余位归还" v-if="currentRefund.status === 'approved'">
          <el-tag type="success" effect="light">已归还 {{ currentRefund.order?.travelers || '-' }} 个余位</el-tag>
        </el-descriptions-item>
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

      <div v-if="currentRefund" style="margin-top: 16px;">
        <el-divider style="margin: 0 0 16px;" />
        <div style="display: flex; align-items: center; gap: 8px; margin-bottom: 16px;">
          <el-icon><Clock /></el-icon>
          <span style="font-weight: 600;">审核记录</span>
          <el-tag size="small" type="info" effect="plain">{{ currentRefund.review_logs?.length || 0 }} 条</el-tag>
        </div>
        <el-timeline v-if="currentRefund.review_logs?.length">
          <el-timeline-item
            v-for="log in currentRefund.review_logs"
            :key="log.id"
            :timestamp="formatDate(log.created_at)"
            placement="top"
            :type="actionType(log.action)"
          >
            <div style="display: flex; align-items: center; gap: 8px; margin-bottom: 6px;">
              <el-tag :type="actionType(log.action)" effect="light" size="small">{{ actionText(log.action) }}</el-tag>
              <span style="color: #909399; font-size: 13px;">操作人：{{ operatorName(log) }}</span>
            </div>
            <div style="color: #606266; font-size: 13px;">
              状态变更：
              <el-tag size="small" type="info" effect="plain">{{ statusText(log.from_status) || '—' }}</el-tag>
              <el-icon style="margin: 0 4px; vertical-align: middle;"><ArrowRight /></el-icon>
              <el-tag size="small" :type="statusType(log.to_status)" effect="light">{{ statusText(log.to_status) }}</el-tag>
            </div>
            <div v-if="log.remark" style="color: #909399; font-size: 13px; margin-top: 4px;">
              备注：{{ log.remark }}
            </div>
          </el-timeline-item>
        </el-timeline>
        <el-empty v-else description="暂无审核记录" :image-size="60" />
      </div>
    </el-dialog>

    <el-dialog v-model="reviewVisible" :title="reviewTitle" width="500px">
      <el-alert
        :title="reviewAlertTitle"
        :type="reviewType === 'approved' ? 'success' : 'warning'"
        :closable="false"
        style="margin-bottom: 20px;"
      >
        <p v-if="reviewType === 'approved'">• 审核通过后，将按退款金额退款至用户账户</p>
        <p v-if="reviewType === 'approved'">• 该订单的 {{ currentReviewRefund?.order?.travelers || '-' }} 个余位将自动归还至行程</p>
        <p v-if="reviewType === 'rejected'">• 拒绝后订单将恢复为已支付状态，余位不归还</p>
      </el-alert>
      <el-form :model="reviewForm" label-width="100px">
        <el-form-item label="退款编号">
          <span>{{ currentReviewRefund?.refund_no }}</span>
        </el-form-item>
        <el-form-item label="退款金额">
          <span style="color: #409EFF; font-weight: 600;">¥{{ currentReviewRefund?.refund_amount?.toFixed(2) }}</span>
        </el-form-item>
        <el-form-item v-if="reviewType === 'approved'" label="退还余位">
          <el-tag type="success" effect="light">+{{ currentReviewRefund?.order?.travelers || 0 }} 位</el-tag>
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

    <el-dialog v-model="batchReviewVisible" :title="batchReviewTitle" width="500px">
      <el-alert
        :title="batchReviewAlertTitle"
        :type="batchReviewType === 'approved' ? 'success' : 'warning'"
        :closable="false"
        style="margin-bottom: 20px;"
      >
        <p v-if="batchReviewType === 'approved'">• 将批量通过 {{ selectedRefunds.length }} 条退款申请</p>
        <p v-if="batchReviewType === 'approved'">• 所有关联订单的余位将自动归还至对应行程</p>
        <p v-if="batchReviewType === 'rejected'">• 将批量拒绝 {{ selectedRefunds.length }} 条退款申请</p>
        <p v-if="batchReviewType === 'rejected'">• 所有关联订单将恢复为已支付状态</p>
      </el-alert>
      <div style="margin-bottom: 16px;">
        <p><strong>待处理数量：</strong>{{ selectedRefunds.length }} 条</p>
        <p v-if="batchReviewType === 'approved'">
          <strong>预计归还余位：</strong>
          <el-tag type="success" effect="light">+{{ totalReturnSpots }} 位</el-tag>
        </p>
      </div>
      <el-form :model="reviewForm" label-width="100px">
        <el-form-item label="统一审核备注">
          <el-input
            v-model="reviewForm.remark"
            type="textarea"
            :rows="3"
            :placeholder="batchReviewType === 'approved' ? '请输入统一备注（选填）' : '请输入统一拒绝原因（选填）'"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="batchReviewVisible = false">取消</el-button>
        <el-button
          :type="batchReviewType === 'approved' ? 'success' : 'danger'"
          @click="submitBatchReview"
          :loading="submitting"
        >
          确认{{ batchReviewType === 'approved' ? '批量通过' : '批量拒绝' }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="reviewLogsVisible" title="审核记录" width="600px">
      <div v-loading="reviewLogsLoading">
        <el-descriptions :column="1" border v-if="currentLogsRefund" style="margin-bottom: 16px;">
          <el-descriptions-item label="退款编号">{{ currentLogsRefund.refund_no }}</el-descriptions-item>
          <el-descriptions-item label="关联订单">
            {{ currentLogsRefund.order_no }}
            <span style="margin-left: 8px; color: #909399;">{{ currentLogsRefund.order?.trip_name || '-' }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="当前状态">
            <el-tag :type="statusType(currentLogsRefund.status)" effect="light">{{ statusText(currentLogsRefund.status) }}</el-tag>
          </el-descriptions-item>
        </el-descriptions>

        <el-timeline v-if="reviewLogs.length">
          <el-timeline-item
            v-for="log in reviewLogs"
            :key="log.id"
            :timestamp="formatDate(log.created_at)"
            placement="top"
            :type="actionType(log.action)"
          >
            <div style="display: flex; align-items: center; gap: 8px; margin-bottom: 6px;">
              <el-tag :type="actionType(log.action)" effect="light" size="small">{{ actionText(log.action) }}</el-tag>
              <span style="color: #909399; font-size: 13px;">操作人：{{ operatorName(log) }}</span>
            </div>
            <div style="color: #606266; font-size: 13px;">
              状态变更：
              <el-tag size="small" type="info" effect="plain">{{ statusText(log.from_status) || '—' }}</el-tag>
              <el-icon style="margin: 0 4px; vertical-align: middle;"><ArrowRight /></el-icon>
              <el-tag size="small" :type="statusType(log.to_status)" effect="light">{{ statusText(log.to_status) }}</el-tag>
            </div>
            <div v-if="log.remark" style="color: #909399; font-size: 13px; margin-top: 4px;">
              备注：{{ log.remark }}
            </div>
          </el-timeline-item>
        </el-timeline>
        <el-empty v-else description="暂无审核记录" />
      </div>
      <template #footer>
        <el-button type="primary" @click="reviewLogsVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getRefundRequests, getRefundRequest, getRefundReviewLogs, reviewRefundRequest, batchReviewRefundRequests } from '@/api/refund'

const loading = ref(false)
const refunds = ref([])
const filterStatus = ref('')
const selectedRefunds = ref([])

const detailVisible = ref(false)
const currentRefund = ref(null)

const reviewVisible = ref(false)
const currentReviewRefund = ref(null)
const reviewType = ref('')
const submitting = ref(false)
const reviewForm = reactive({
  remark: ''
})

const batchReviewVisible = ref(false)
const batchReviewType = ref('')

const reviewLogsVisible = ref(false)
const reviewLogsLoading = ref(false)
const reviewLogs = ref([])
const currentLogsRefund = ref(null)

const reviewTitle = computed(() => {
  return reviewType.value === 'approved' ? '审核通过' : '审核拒绝'
})

const reviewAlertTitle = computed(() => {
  return reviewType.value === 'approved' ? '通过后将按退款金额退款至用户账户并归还余位' : '拒绝后订单将恢复为已支付状态'
})

const batchReviewTitle = computed(() => {
  return batchReviewType.value === 'approved' ? '批量审核通过' : '批量审核拒绝'
})

const batchReviewAlertTitle = computed(() => {
  return batchReviewType.value === 'approved' ? '批量通过后将自动退款并归还所有余位' : '批量拒绝后所有订单将恢复为已支付状态'
})

const totalReturnSpots = computed(() => {
  return selectedRefunds.value.reduce((sum, item) => sum + (item.order?.travelers || 0), 0)
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

function handleSelectionChange(selection) {
  selectedRefunds.value = selection.filter(item => item.status === 'pending')
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

function actionText(action) {
  const map = {
    submitted: '提交申请',
    approved: '审核通过',
    rejected: '审核拒绝'
  }
  return map[action] || action
}

function actionType(action) {
  const map = {
    submitted: 'primary',
    approved: 'success',
    rejected: 'danger'
  }
  return map[action] || 'info'
}

function operatorName(log) {
  if (log.operator?.username) {
    const role = log.operator.role === 'admin' ? '管理员' : '用户'
    return `${role}（${log.operator.username}）`
  }
  if (log.action === 'submitted') return '游客用户'
  return '管理员'
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

async function openReviewLogsDialog(row) {
  currentLogsRefund.value = row
  reviewLogs.value = []
  reviewLogsVisible.value = true
  reviewLogsLoading.value = true
  try {
    const res = await getRefundReviewLogs(row.id)
    reviewLogs.value = res.data || []
  } finally {
    reviewLogsLoading.value = false
  }
}

function openReviewDialog(row, type) {
  currentReviewRefund.value = row
  reviewType.value = type
  reviewForm.remark = ''
  reviewVisible.value = true
}

function openBatchReview(type) {
  if (selectedRefunds.value.length === 0) {
    ElMessage.warning('请先选择待审核的退款申请')
    return
  }
  batchReviewType.value = type
  reviewForm.remark = ''
  batchReviewVisible.value = true
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
      ElMessage.success(
        reviewType.value === 'approved'
          ? '审核通过，已完成退款并归还余位'
          : '已拒绝该申请'
      )
      reviewVisible.value = false
      loadRefunds()
    } finally {
      submitting.value = false
    }
  }).catch(() => {})
}

async function submitBatchReview() {
  const ids = selectedRefunds.value.map(item => item.id)
  ElMessageBox.confirm(
    `确认${batchReviewType.value === 'approved' ? '批量通过' : '批量拒绝'} ${ids.length} 条退款申请？`,
    '批量审核确认',
    {
      confirmButtonText: '确认',
      cancelButtonText: '取消',
      type: batchReviewType.value === 'approved' ? 'success' : 'warning'
    }
  ).then(async () => {
    submitting.value = true
    try {
      const res = await batchReviewRefundRequests({
        ids,
        status: batchReviewType.value,
        review_remark: reviewForm.remark
      })
      ElMessage.success(
        `批量审核完成！成功处理 ${res.data.success_count}/${res.data.total_count} 条`
      )
      batchReviewVisible.value = false
      selectedRefunds.value = []
      loadRefunds()
    } finally {
      submitting.value = false
    }
  }).catch(() => {})
}

onMounted(loadRefunds)
</script>
