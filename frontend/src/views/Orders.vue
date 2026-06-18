<template>
  <div class="page-container">
    <div class="page-header">
      <h2>订单管理</h2>
      <p>管理所有预订订单，支持查看详情、支付和申请退款等操作</p>
    </div>

    <div class="card-content">
      <div class="toolbar">
        <div class="search-filters">
          <el-date-picker
            v-model="filterDateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="YYYY-MM-DD"
            style="width: 280px;"
          />
          <el-select v-model="filterStatus" placeholder="订单状态" clearable style="width: 160px;">
            <el-option label="待支付" value="pending" />
            <el-option label="已支付" value="paid" />
            <el-option label="退款中" value="refunding" />
            <el-option label="已退款" value="refunded" />
            <el-option label="已完成" value="completed" />
            <el-option label="已取消" value="cancelled" />
          </el-select>
          <el-button type="primary" @click="loadOrders">
            <el-icon><Search /></el-icon>
            <span>查询</span>
          </el-button>
        </div>
        <el-button type="success" @click="loadOrders">
          <el-icon><Refresh /></el-icon>
          <span>刷新</span>
        </el-button>
      </div>

      <el-table :data="orders" style="width: 100%" v-loading="loading" stripe>
        <el-table-column prop="order_no" label="订单编号" width="200" />
        <el-table-column prop="trip_name" label="行程名称" />
        <el-table-column prop="travelers" label="人数" width="80" align="center" />
        <el-table-column label="订单金额" width="120">
          <template #default="{ row }">
            <span style="color: #f56c6c; font-weight: 600;">¥{{ row.total_amount?.toFixed(2) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="contact_name" label="联系人" width="120" />
        <el-table-column prop="contact_phone" label="联系电话" width="140" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)" effect="light">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="170">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="viewDetail(row)">
              <el-icon><View /></el-icon>
              查看详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="!loading && orders.length === 0" description="暂无订单数据" />
    </div>

    <el-dialog v-model="detailVisible" title="订单详情" width="720px" class="order-detail-dialog">
      <el-tabs v-model="detailTab" v-if="currentOrder">
        <el-tab-pane label="基本信息" name="basic">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="订单编号">{{ currentOrder.order_no }}</el-descriptions-item>
            <el-descriptions-item label="订单状态">
              <el-tag :type="statusType(currentOrder.status)" effect="light">{{ statusText(currentOrder.status) }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="行程名称" :span="2">{{ currentOrder.trip_name }}</el-descriptions-item>
            <el-descriptions-item label="出发日期">{{ currentOrder.trip ? formatDate(currentOrder.trip.start_date) : '-' }}</el-descriptions-item>
            <el-descriptions-item label="返程日期">{{ currentOrder.trip ? formatDate(currentOrder.trip.end_date) : '-' }}</el-descriptions-item>
            <el-descriptions-item label="出行人数">{{ currentOrder.travelers }} 人</el-descriptions-item>
            <el-descriptions-item label="单价">¥{{ currentOrder.trip_price?.toFixed(2) }}</el-descriptions-item>
            <el-descriptions-item label="订单金额">
              <span style="color: #f56c6c; font-weight: 600;">¥{{ currentOrder.total_amount?.toFixed(2) }}</span>
            </el-descriptions-item>
            <el-descriptions-item label="联系人">{{ currentOrder.contact_name }}</el-descriptions-item>
            <el-descriptions-item label="联系电话">{{ currentOrder.contact_phone }}</el-descriptions-item>
            <el-descriptions-item label="创建时间" :span="2">{{ formatDate(currentOrder.created_at) }}</el-descriptions-item>
            <el-descriptions-item label="支付时间" :span="2">{{ currentOrder.pay_time ? formatDate(currentOrder.pay_time) : '未支付' }}</el-descriptions-item>
          </el-descriptions>
        </el-tab-pane>
        <el-tab-pane label="每日行程安排" name="itinerary">
          <div v-if="itineraries && itineraries.length" style="max-height: 480px; overflow-y: auto; padding-right: 8px;">
            <el-timeline>
              <el-timeline-item
                v-for="item in itineraries"
                :key="item.id"
                :timestamp="`第 ${item.day_number} 天`"
                placement="top"
                :type="getDayType(item.day_number)"
              >
                <el-card shadow="hover" class="itinerary-card">
                  <template #header>
                    <div style="display: flex; justify-content: space-between; align-items: center;">
                      <span style="font-weight: 600; font-size: 15px;">{{ item.title }}</span>
                    </div>
                  </template>
                  <el-descriptions :column="2" size="small" border>
                    <el-descriptions-item label="早餐">{{ item.breakfast || '-' }}</el-descriptions-item>
                    <el-descriptions-item label="午餐">{{ item.lunch || '-' }}</el-descriptions-item>
                    <el-descriptions-item label="晚餐">{{ item.dinner || '-' }}</el-descriptions-item>
                    <el-descriptions-item label="住宿">{{ item.accommodation || '-' }}</el-descriptions-item>
                    <el-descriptions-item label="交通">{{ item.transportation || '-' }}</el-descriptions-item>
                    <el-descriptions-item label="活动安排" :span="2">
                      <span style="white-space: pre-wrap;">{{ item.activities || '-' }}</span>
                    </el-descriptions-item>
                    <el-descriptions-item label="备注" :span="2">
                      <span style="white-space: pre-wrap; color: #e6a23c;">{{ item.notes || '-' }}</span>
                    </el-descriptions-item>
                  </el-descriptions>
                </el-card>
              </el-timeline-item>
            </el-timeline>
          </div>
          <el-empty v-else description="暂无每日行程安排" />
        </el-tab-pane>
      </el-tabs>
      <template #footer>
        <div class="detail-footer-actions">
          <el-button
            v-if="currentOrder && currentOrder.status === 'pending'"
            type="success"
            @click="handlePay(currentOrder); detailVisible = false;"
          >
            <el-icon><Money /></el-icon>
            <span>立即支付</span>
          </el-button>
          <el-button
            v-if="currentOrder && (currentOrder.status === 'paid' || currentOrder.status === 'confirmed')"
            type="warning"
            @click="openRefundDialog(currentOrder); detailVisible = false;"
          >
            <el-icon><Refund /></el-icon>
            <span>申请售后退款</span>
          </el-button>
          <el-button @click="detailVisible = false">关闭</el-button>
        </div>
      </template>
    </el-dialog>

    <el-dialog v-model="refundDialogVisible" title="申请退款" width="520px">
      <el-alert
        title="退款说明"
        type="info"
        :closable="false"
        style="margin-bottom: 20px;"
      >
        <p>• 退款申请提交后，需要客服审核，审核通过后款项将原路退回</p>
        <p>• 出发前7天以上申请可全额退款，7天内将收取一定手续费</p>
      </el-alert>
      <el-form :model="refundForm" label-width="100px" ref="refundFormRef" :rules="refundRules">
        <el-form-item label="订单编号">
          <span>{{ currentRefundOrder?.order_no }}</span>
        </el-form-item>
        <el-form-item label="订单金额">
          <span style="color: #f56c6c; font-weight: 600;">¥{{ currentRefundOrder?.total_amount?.toFixed(2) }}</span>
        </el-form-item>
        <el-form-item label="退款原因" prop="reason">
          <el-select v-model="refundForm.reason" placeholder="请选择退款原因" style="width: 100%;">
            <el-option label="行程变更" value="change_of_plan" />
            <el-option label="身体原因" value="health_issue" />
            <el-option label="资金问题" value="financial" />
            <el-option label="服务不满意" value="service_issue" />
            <el-option label="其他原因" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="详细说明" prop="description">
          <el-input
            v-model="refundForm.description"
            type="textarea"
            :rows="4"
            placeholder="请详细说明退款原因，便于我们为您更好地处理"
          />
        </el-form-item>
        <el-form-item label="退款金额" prop="refundAmount">
          <el-input-number
            v-model="refundForm.refundAmount"
            :min="0"
            :max="currentRefundOrder?.total_amount || 0"
            :precision="2"
            style="width: 200px;"
          />
          <span style="margin-left: 8px; color: #909399; font-size: 12px;">
            最多可退 ¥{{ currentRefundOrder?.total_amount?.toFixed(2) }}
          </span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="refundDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitRefund" :loading="submitting">提交申请</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getOrders, getOrder, payOrder } from '@/api/order'
import { createRefundRequest } from '@/api/refund'

const loading = ref(false)
const orders = ref([])
const filterStatus = ref('')
const filterDateRange = ref([])

const detailVisible = ref(false)
const currentOrder = ref(null)
const detailTab = ref('basic')
const itineraries = ref([])

const refundDialogVisible = ref(false)
const currentRefundOrder = ref(null)
const submitting = ref(false)
const refundFormRef = ref(null)

const refundForm = reactive({
  reason: '',
  description: '',
  refundAmount: 0
})

const refundRules = {
  reason: [{ required: true, message: '请选择退款原因', trigger: 'change' }],
  refundAmount: [{ required: true, message: '请填写退款金额', trigger: 'blur' }]
}

async function loadOrders() {
  loading.value = true
  try {
    const params = {}
    if (filterStatus.value) params.status = filterStatus.value
    if (filterDateRange.value && filterDateRange.value.length === 2) {
      params.start_date = filterDateRange.value[0]
      params.end_date = filterDateRange.value[1]
    }
    const res = await getOrders(params)
    orders.value = res.data || []
  } finally {
    loading.value = false
  }
}

function statusText(status) {
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

function statusType(status) {
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

function formatDate(date) {
  if (!date) return '-'
  return new Date(date).toLocaleString('zh-CN', { hour12: false }).replace(/\//g, '-')
}

async function viewDetail(row) {
  detailTab.value = 'basic'
  itineraries.value = []
  try {
    const res = await getOrder(row.id)
    currentOrder.value = res.data
    if (res.data.trip && res.data.trip.itineraries) {
      itineraries.value = [...res.data.trip.itineraries].sort((a, b) => a.day_number - b.day_number)
    }
    detailVisible.value = true
  } catch (e) {
    currentOrder.value = row
    detailVisible.value = true
  }
}

function getDayType(day) {
  const types = ['primary', 'success', 'warning', 'danger', 'info']
  return types[(day - 1) % types.length]
}

async function handlePay(row) {
  ElMessageBox.confirm(
    `确认支付订单 ${row.order_no}，金额 ¥${row.total_amount?.toFixed(2)}？`,
    '支付确认',
    {
      confirmButtonText: '确认支付',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(async () => {
    await payOrder(row.id)
    ElMessage.success('支付成功！')
    loadOrders()
  }).catch(() => {})
}

function openRefundDialog(row) {
  currentRefundOrder.value = row
  refundForm.reason = ''
  refundForm.description = ''
  refundForm.refundAmount = row.total_amount
  refundDialogVisible.value = true
}

async function submitRefund() {
  if (!refundFormRef.value) return
  await refundFormRef.value.validate(async (valid) => {
    if (!valid) return
    submitting.value = true
    try {
      await createRefundRequest({
        order_id: currentRefundOrder.value.id,
        reason: refundForm.reason,
        description: refundForm.description,
        refund_amount: refundForm.refundAmount
      })
      ElMessage.success('退款申请已提交，请等待审核')
      refundDialogVisible.value = false
      loadOrders()
    } finally {
      submitting.value = false
    }
  })
}

onMounted(loadOrders)
</script>

<style scoped>
.order-detail-dialog :deep(.el-dialog__body) {
  padding-top: 8px;
}
.itinerary-card {
  margin-bottom: 4px;
}
.itinerary-card :deep(.el-descriptions__label) {
  width: 70px;
}
.detail-footer-actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}
</style>
