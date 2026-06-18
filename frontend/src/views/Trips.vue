<template>
  <div class="page-container">
    <div class="page-header">
      <h2>行程管理</h2>
      <p>查看所有可预订的旅游行程信息</p>
    </div>

    <div class="card-content">
      <div v-loading="loading" style="min-height: 400px;">
        <el-row :gutter="20" v-if="trips.length">
          <el-col :span="8" v-for="trip in trips" :key="trip.id" style="margin-bottom: 20px;">
            <el-card shadow="hover" class="trip-card">
              <div class="trip-image">
                <el-icon :size="48" color="#409EFF"><Picture /></el-icon>
              </div>
              <div class="trip-body">
                <h3 class="trip-name">{{ trip.name }}</h3>
                <div class="trip-desc">{{ trip.description }}</div>
                <div class="trip-meta">
                  <div class="meta-item">
                    <el-icon><Location /></el-icon>
                    <span>{{ trip.destination }}</span>
                  </div>
                  <div class="meta-item">
                    <el-icon><Calendar /></el-icon>
                    <span>{{ formatDate(trip.start_date) }} 至 {{ formatDate(trip.end_date) }}</span>
                  </div>
                </div>
                <div class="trip-footer">
                  <div class="price">
                    <span class="price-label">价格</span>
                    <span class="price-value">¥{{ trip.price }}</span>
                    <span class="price-unit">/人</span>
                  </div>
                  <div class="spots">
                    <el-tag :type="trip.left_spots > 5 ? 'success' : 'warning'" effect="light">
                      剩余 {{ trip.left_spots }} / {{ trip.total_spots }} 位
                    </el-tag>
                  </div>
                </div>
                <el-divider style="margin: 12px 0;" />
                <div style="display: flex; gap: 8px;">
                  <el-button type="primary" @click="handleBook(trip)" :disabled="trip.left_spots <= 0">
                    <el-icon><ShoppingCart /></el-icon>
                    <span>立即预订</span>
                  </el-button>
                  <el-button @click="handleDetail(trip)">查看详情</el-button>
                </div>
              </div>
            </el-card>
          </el-col>
        </el-row>
        <el-empty v-else description="暂无行程数据，请先初始化测试数据" />
      </div>
    </div>

    <el-dialog v-model="bookDialogVisible" title="预订行程" width="500px">
      <el-form :model="bookForm" label-width="100px" ref="bookFormRef" :rules="bookRules">
        <el-form-item label="行程名称">
          <span style="color: #303133;">{{ currentTrip?.name }}</span>
        </el-form-item>
        <el-form-item label="行程价格">
          <span style="color: #f56c6c; font-weight: 600;">¥{{ currentTrip?.price }} /人</span>
        </el-form-item>
        <el-form-item label="出行人数" prop="travelers">
          <el-input-number v-model="bookForm.travelers" :min="1" :max="currentTrip?.left_spots || 1" />
        </el-form-item>
        <el-form-item label="联系人" prop="contactName">
          <el-input v-model="bookForm.contactName" placeholder="请输入联系人姓名" />
        </el-form-item>
        <el-form-item label="联系电话" prop="contactPhone">
          <el-input v-model="bookForm.contactPhone" placeholder="请输入联系电话" />
        </el-form-item>
        <el-form-item label="订单金额">
          <span style="color: #f56c6c; font-size: 18px; font-weight: 700;">
            ¥{{ (bookForm.travelers * (currentTrip?.price || 0)).toFixed(2) }}
          </span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="bookDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitBook" :loading="submitting">提交订单</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getTrips } from '@/api/common'
import { createOrder, payOrder } from '@/api/order'

const loading = ref(false)
const trips = ref([])
const bookDialogVisible = ref(false)
const currentTrip = ref(null)
const submitting = ref(false)
const bookFormRef = ref(null)

const bookForm = reactive({
  travelers: 1,
  contactName: '',
  contactPhone: ''
})

const bookRules = {
  travelers: [{ required: true, message: '请选择出行人数', trigger: 'change' }],
  contactName: [{ required: true, message: '请输入联系人姓名', trigger: 'blur' }],
  contactPhone: [{ required: true, message: '请输入联系电话', trigger: 'blur' }]
}

async function loadTrips() {
  loading.value = true
  try {
    const res = await getTrips()
    trips.value = res.data || []
  } finally {
    loading.value = false
  }
}

function handleBook(trip) {
  currentTrip.value = trip
  bookForm.travelers = 1
  bookForm.contactName = ''
  bookForm.contactPhone = ''
  bookDialogVisible.value = true
}

async function submitBook() {
  if (!bookFormRef.value) return
  await bookFormRef.value.validate(async (valid) => {
    if (!valid) return
    submitting.value = true
    try {
      const orderData = {
        user_id: 2,
        trip_id: currentTrip.value.id,
        trip_name: currentTrip.value.name,
        trip_price: currentTrip.value.price,
        travelers: bookForm.travelers,
        total_amount: bookForm.travelers * currentTrip.value.price,
        contact_name: bookForm.contactName,
        contact_phone: bookForm.contactPhone
      }
      const res = await createOrder(orderData)
      bookDialogVisible.value = false

      ElMessageBox.confirm(
        `订单创建成功！订单号：${res.data.order_no}\n是否立即支付？`,
        '支付确认',
        {
          confirmButtonText: '立即支付',
          cancelButtonText: '稍后支付',
          type: 'success'
        }
      ).then(async () => {
        await payOrder(res.data.id)
        ElMessage.success('支付成功！')
        loadTrips()
      }).catch(() => {
        ElMessage.info('您可以在订单管理中完成支付')
      })
    } finally {
      submitting.value = false
    }
  })
}

function handleDetail(trip) {
  ElMessageBox.alert(
    `
      <div style="line-height: 1.8;">
        <p><strong>目的地：</strong>${trip.destination}</p>
        <p><strong>出发日期：</strong>${formatDate(trip.start_date)}</p>
        <p><strong>返程日期：</strong>${formatDate(trip.end_date)}</p>
        <p><strong>价格：</strong>¥${trip.price}/人</p>
        <p><strong>余位：</strong>${trip.left_spots}/${trip.total_spots} 位</p>
        <p style="margin-top: 12px;"><strong>行程介绍：</strong></p>
        <p>${trip.description}</p>
      </div>
    `,
    trip.name,
    {
      dangerouslyUseHTMLString: true,
      confirmButtonText: '关闭'
    }
  )
}

function formatDate(date) {
  if (!date) return '-'
  return new Date(date).toLocaleDateString('zh-CN').replace(/\//g, '-')
}

onMounted(loadTrips)
</script>

<style scoped>
.trip-card {
  border-radius: 12px;
  overflow: hidden;
  transition: transform 0.3s;
}
.trip-card:hover {
  transform: translateY(-4px);
}
.trip-image {
  height: 140px;
  background: linear-gradient(135deg, #e0f2fe 0%, #dbeafe 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 12px 12px 0 0;
  margin: -1px;
}
.trip-body {
  padding-top: 8px;
}
.trip-name {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 6px;
}
.trip-desc {
  color: #909399;
  font-size: 13px;
  margin-bottom: 12px;
  line-height: 1.5;
  height: 39px;
  overflow: hidden;
}
.trip-meta {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-bottom: 12px;
}
.meta-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: #606266;
}
.trip-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.price {
  display: flex;
  align-items: baseline;
  gap: 4px;
}
.price-label {
  font-size: 12px;
  color: #909399;
}
.price-value {
  font-size: 22px;
  font-weight: 700;
  color: #f56c6c;
}
.price-unit {
  font-size: 12px;
  color: #909399;
}
</style>
