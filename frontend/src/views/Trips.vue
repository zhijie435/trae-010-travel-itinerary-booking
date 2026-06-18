<template>
  <div class="page-container">
    <div class="page-header">
      <h2>行程管理</h2>
      <p>管理所有旅游行程信息，支持新增、编辑、调整余位、设置每日行程安排等操作</p>
    </div>

    <div class="card-content">
      <div class="toolbar">
        <div class="search-filters">
          <el-input v-model="searchKeyword" placeholder="搜索行程名称/目的地" clearable style="width: 240px;" />
          <el-button type="primary" @click="loadTrips">
            <el-icon><Search /></el-icon>
            <span>查询</span>
          </el-button>
        </div>
        <div>
          <el-button type="success" @click="loadTrips">
            <el-icon><Refresh /></el-icon>
            <span>刷新</span>
          </el-button>
          <el-button type="primary" @click="openCreateDialog">
            <el-icon><Plus /></el-icon>
            <span>新增行程</span>
          </el-button>
        </div>
      </div>

      <div v-loading="loading" style="min-height: 400px;">
        <el-row :gutter="20" v-if="filteredTrips.length">
          <el-col :span="8" v-for="trip in filteredTrips" :key="trip.id" style="margin-bottom: 20px;">
            <el-card shadow="hover" class="trip-card">
              <div class="trip-image">
                <el-icon :size="48" color="#409EFF"><Picture /></el-icon>
                <div class="trip-actions">
                  <el-dropdown trigger="click" @command="(cmd) => handleAction(cmd, trip)">
                    <el-button type="primary" circle size="small" :icon="MoreFilled" />
                    <template #dropdown>
                      <el-dropdown-menu>
                        <el-dropdown-item command="edit">
                          <el-icon><Edit /></el-icon> 编辑行程
                        </el-dropdown-item>
                        <el-dropdown-item command="adjust">
                          <el-icon><User /></el-icon> 调整余位
                        </el-dropdown-item>
                        <el-dropdown-item command="itinerary">
                          <el-icon><Calendar /></el-icon> 每日行程安排
                        </el-dropdown-item>
                        <el-dropdown-item command="logs">
                          <el-icon><Document /></el-icon> 余位调整日志
                        </el-dropdown-item>
                        <el-dropdown-item command="detail">
                          <el-icon><View /></el-icon> 查看详情
                        </el-dropdown-item>
                        <el-dropdown-item divided command="delete" class="danger-text">
                          <el-icon><Delete /></el-icon> 删除行程
                        </el-dropdown-item>
                      </el-dropdown-menu>
                    </template>
                  </el-dropdown>
                </div>
              </div>
              <div class="trip-body">
                <div style="display: flex; justify-content: space-between; align-items: start;">
                  <h3 class="trip-name">{{ trip.name }}</h3>
                  <el-tag :type="trip.status === 'active' ? 'success' : 'info'" effect="light" size="small">
                    {{ trip.status === 'active' ? '进行中' : '已下架' }}
                  </el-tag>
                </div>
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
                  <el-button type="primary" size="small" @click="openItineraryDialog(trip)">
                    <el-icon><Calendar /></el-icon>
                    <span>行程安排</span>
                  </el-button>
                  <el-button size="small" @click="handleAdjustSpots(trip)">
                    <el-icon><User /></el-icon>
                    <span>调整余位</span>
                  </el-button>
                  <el-button size="small" @click="handleBook(trip)" :disabled="trip.left_spots <= 0 || trip.status !== 'active'">
                    <el-icon><ShoppingCart /></el-icon>
                    <span>预订</span>
                  </el-button>
                </div>
              </div>
            </el-card>
          </el-col>
        </el-row>
        <el-empty v-else description="暂无行程数据，请先初始化测试数据或新增行程" />
      </div>
    </div>

    <el-dialog v-model="tripDialogVisible" :title="isEdit ? '编辑行程' : '新增行程'" width="620px" :close-on-click-modal="false">
      <el-form :model="tripForm" label-width="100px" ref="tripFormRef" :rules="tripRules">
        <el-form-item label="行程名称" prop="name">
          <el-input v-model="tripForm.name" placeholder="请输入行程名称" />
        </el-form-item>
        <el-form-item label="目的地" prop="destination">
          <el-input v-model="tripForm.destination" placeholder="请输入目的地" />
        </el-form-item>
        <el-form-item label="行程描述" prop="description">
          <el-input v-model="tripForm.description" type="textarea" :rows="3" placeholder="请输入行程描述" />
        </el-form-item>
        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="开始日期" prop="start_date">
              <el-date-picker v-model="tripForm.start_date" type="date" placeholder="选择开始日期" style="width: 100%;" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="结束日期" prop="end_date">
              <el-date-picker v-model="tripForm.end_date" type="date" placeholder="选择结束日期" style="width: 100%;" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="价格" prop="price">
              <el-input-number v-model="tripForm.price" :min="0" :precision="2" style="width: 100%;" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="总名额" prop="total_spots">
              <el-input-number v-model="tripForm.total_spots" :min="1" style="width: 100%;" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="状态">
          <el-radio-group v-model="tripForm.status">
            <el-radio value="active">上架中</el-radio>
            <el-radio value="inactive">已下架</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="tripDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitTripForm" :loading="submitting">
          {{ isEdit ? '保存修改' : '创建行程' }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="adjustDialogVisible" title="调整余位" width="480px" :close-on-click-modal="false">
      <div style="margin-bottom: 16px;">
        <p style="margin-bottom: 8px;">
          <strong>行程：</strong>{{ currentTrip?.name }}
        </p>
        <p>
          <strong>当前余位：</strong>
          <span style="color: #f56c6c; font-weight: 600;">{{ currentTrip?.left_spots }} / {{ currentTrip?.total_spots }}</span>
        </p>
      </div>
      <el-form :model="adjustForm" label-width="100px" ref="adjustFormRef" :rules="adjustRules">
        <el-form-item label="调整数量" prop="adjust_amount">
          <el-input-number
            v-model="adjustForm.adjust_amount"
            :min="-currentTrip?.left_spots || 0"
            :max="(currentTrip?.total_spots || 0) - (currentTrip?.left_spots || 0)"
            style="width: 100%;"
          />
          <div style="color: #909399; font-size: 12px; margin-top: 4px;">
            正数增加余位，负数减少余位。可调整范围：-{{ currentTrip?.left_spots || 0 }} 到 {{ (currentTrip?.total_spots || 0) - (currentTrip?.left_spots || 0) }}
          </div>
        </el-form-item>
        <el-form-item label="调整后余位">
          <span style="color: #409EFF; font-weight: 600; font-size: 18px;">
            {{ (currentTrip?.left_spots || 0) + (adjustForm.adjust_amount || 0) }}
          </span>
        </el-form-item>
        <el-form-item label="调整原因" prop="reason">
          <el-input v-model="adjustForm.reason" type="textarea" :rows="3" placeholder="请输入调整原因" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="adjustDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitAdjustSpots" :loading="submitting">确认调整</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="itineraryDialogVisible" :title="`每日行程安排 - ${currentTrip?.name}`" width="900px" class="itinerary-dialog">
      <div style="margin-bottom: 16px;">
        <el-button type="primary" size="small" @click="openItineraryForm(null)">
          <el-icon><Plus /></el-icon>
          <span>新增行程安排</span>
        </el-button>
        <el-button size="small" @click="loadItineraries">
          <el-icon><Refresh /></el-icon>
          <span>刷新</span>
        </el-button>
      </div>

      <div v-loading="itineraryLoading" style="max-height: 500px; overflow-y: auto;">
        <el-timeline v-if="itineraries.length">
          <el-timeline-item
            v-for="item in itineraries"
            :key="item.id"
            :timestamp="`第 ${item.day_number} 天`"
            placement="top"
            :type="getDayType(item.day_number)"
          >
            <el-card shadow="hover">
              <template #header>
                <div style="display: flex; justify-content: space-between; align-items: center;">
                  <div>
                    <span style="font-weight: 600; font-size: 16px;">{{ item.title }}</span>
                  </div>
                  <div>
                    <el-button type="primary" link size="small" @click="openItineraryForm(item)">
                      <el-icon><Edit /></el-icon> 编辑
                    </el-button>
                    <el-button type="danger" link size="small" @click="deleteItinerary(item)">
                      <el-icon><Delete /></el-icon> 删除
                    </el-button>
                  </div>
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
                  <span style="white-space: pre-wrap;">{{ item.notes || '-' }}</span>
                </el-descriptions-item>
              </el-descriptions>
            </el-card>
          </el-timeline-item>
        </el-timeline>
        <el-empty v-else description="暂无每日行程安排，请点击上方按钮新增" />
      </div>
    </el-dialog>

    <el-dialog v-model="itineraryFormVisible" :title="isItineraryEdit ? '编辑行程安排' : '新增行程安排'" width="560px" :close-on-click-modal="false">
      <el-form :model="itineraryForm" label-width="100px" ref="itineraryFormRef" :rules="itineraryRules">
        <el-form-item label="第几天" prop="day_number">
          <el-input-number v-model="itineraryForm.day_number" :min="1" />
        </el-form-item>
        <el-form-item label="行程标题" prop="title">
          <el-input v-model="itineraryForm.title" placeholder="请输入行程标题，如：洱海环游" />
        </el-form-item>
        <el-row :gutter="12">
          <el-col :span="8">
            <el-form-item label="早餐">
              <el-input v-model="itineraryForm.breakfast" placeholder="如：酒店早餐" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="午餐">
              <el-input v-model="itineraryForm.lunch" placeholder="如：特色餐厅" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="晚餐">
              <el-input v-model="itineraryForm.dinner" placeholder="如：自理" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="住宿">
          <el-input v-model="itineraryForm.accommodation" placeholder="如：大理古城特色客栈" />
        </el-form-item>
        <el-form-item label="交通">
          <el-input v-model="itineraryForm.transportation" placeholder="如：商务车" />
        </el-form-item>
        <el-form-item label="活动安排">
          <el-input v-model="itineraryForm.activities" type="textarea" :rows="3" placeholder="请详细描述当天的活动安排" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="itineraryForm.notes" type="textarea" :rows="2" placeholder="温馨提示或注意事项" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="itineraryFormVisible = false">取消</el-button>
        <el-button type="primary" @click="submitItinerary" :loading="submitting">
          {{ isItineraryEdit ? '保存修改' : '新增' }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="logsDialogVisible" title="余位调整日志" width="720px">
      <div v-loading="logsLoading">
        <el-table :data="spotLogs" stripe style="width: 100%;" v-if="spotLogs.length">
          <el-table-column label="调整类型" width="160">
            <template #default="{ row }">
              <el-tag :type="adjustLogType(row.adjust_type)" effect="light" size="small">
                {{ adjustLogText(row.adjust_type) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="调整数量" width="100" align="center">
            <template #default="{ row }">
              <span :style="{ color: row.adjust_amount > 0 ? '#67c23a' : '#f56c6c' }">
                {{ row.adjust_amount > 0 ? '+' : '' }}{{ row.adjust_amount }}
              </span>
            </template>
          </el-table-column>
          <el-table-column label="变更前" width="80" align="center">
            <template #default="{ row }">{{ row.old_spots }}</template>
          </el-table-column>
          <el-table-column label="变更后" width="80" align="center">
            <template #default="{ row }">{{ row.new_spots }}</template>
          </el-table-column>
          <el-table-column label="调整原因" min-width="180">
            <template #default="{ row }">{{ row.reason || '-' }}</template>
          </el-table-column>
          <el-table-column label="关联订单" width="100">
            <template #default="{ row }">{{ row.order_id || '-' }}</template>
          </el-table-column>
          <el-table-column label="调整时间" width="160">
            <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
          </el-table-column>
        </el-table>
        <el-empty v-else description="暂无调整日志" />
      </div>
    </el-dialog>

    <el-dialog v-model="bookDialogVisible" title="预订行程" width="520px">
      <el-form :model="bookForm" label-width="100px" ref="bookFormRef" :rules="bookRules">
        <el-form-item label="行程名称">
          <span style="color: #303133;">{{ currentTrip?.name }}</span>
        </el-form-item>
        <el-form-item label="行程价格">
          <span style="color: #f56c6c; font-weight: 600;">¥{{ currentTrip?.price }} /人</span>
        </el-form-item>
        <el-form-item label="预订方式">
          <el-radio-group v-model="bookForm.bookMode">
            <el-radio value="guest">
              <el-icon><User /></el-icon>
              <span>游客下单（无需登录）</span>
            </el-radio>
            <el-radio value="user">
              <el-icon><Avatar /></el-icon>
              <span>会员下单</span>
            </el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="bookForm.bookMode === 'user'" label="选择用户" prop="userId">
          <el-select v-model="bookForm.userId" placeholder="请选择会员用户" style="width: 100%;" filterable>
            <el-option
              v-for="user in userList"
              :key="user.id"
              :label="`${user.username}（${user.phone || user.email}）`"
              :value="user.id"
            />
          </el-select>
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
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { MoreFilled, Search, Refresh, Plus, Picture, Location, Calendar, User, Edit, Document, View, Delete, ShoppingCart, Avatar } from '@element-plus/icons-vue'
import {
  getTrips,
  getTrip,
  createTrip,
  updateTrip,
  deleteTrip,
  adjustTripSpots,
  getTripItineraries,
  createItinerary,
  updateItinerary,
  deleteItinerary as apiDeleteItinerary,
  getSpotLogs
} from '@/api/trip'
import { createOrder, payOrder } from '@/api/order'
import request from '@/utils/request'

function getUsers() {
  return request({
    url: '/users',
    method: 'get'
  })
}

const loading = ref(false)
const trips = ref([])
const searchKeyword = ref('')

const tripDialogVisible = ref(false)
const isEdit = ref(false)
const currentTrip = ref(null)
const submitting = ref(false)
const tripFormRef = ref(null)

const tripForm = reactive({
  name: '',
  description: '',
  destination: '',
  start_date: '',
  end_date: '',
  price: 0,
  total_spots: 10,
  status: 'active'
})

const tripRules = {
  name: [{ required: true, message: '请输入行程名称', trigger: 'blur' }],
  destination: [{ required: true, message: '请输入目的地', trigger: 'blur' }],
  description: [{ required: true, message: '请输入行程描述', trigger: 'blur' }],
  start_date: [{ required: true, message: '请选择开始日期', trigger: 'change' }],
  end_date: [{ required: true, message: '请选择结束日期', trigger: 'change' }],
  price: [{ required: true, message: '请输入价格', trigger: 'change' }],
  total_spots: [{ required: true, message: '请输入总名额', trigger: 'change' }]
}

const adjustDialogVisible = ref(false)
const adjustFormRef = ref(null)
const adjustForm = reactive({
  adjust_amount: 0,
  reason: ''
})

const adjustRules = {
  adjust_amount: [{ required: true, message: '请输入调整数量', trigger: 'change' }],
  reason: [{ required: true, message: '请输入调整原因', trigger: 'blur' }]
}

const itineraryDialogVisible = ref(false)
const itineraryLoading = ref(false)
const itineraries = ref([])

const itineraryFormVisible = ref(false)
const isItineraryEdit = ref(false)
const currentItinerary = ref(null)
const itineraryFormRef = ref(null)
const itineraryForm = reactive({
  day_number: 1,
  title: '',
  breakfast: '',
  lunch: '',
  dinner: '',
  accommodation: '',
  transportation: '',
  activities: '',
  notes: ''
})

const itineraryRules = {
  day_number: [{ required: true, message: '请输入天数', trigger: 'change' }],
  title: [{ required: true, message: '请输入行程标题', trigger: 'blur' }]
}

const logsDialogVisible = ref(false)
const logsLoading = ref(false)
const spotLogs = ref([])

const bookDialogVisible = ref(false)
const bookFormRef = ref(null)
const userList = ref([])
const bookForm = reactive({
  bookMode: 'guest',
  userId: null,
  travelers: 1,
  contactName: '',
  contactPhone: ''
})

const bookRules = {
  travelers: [{ required: true, message: '请选择出行人数', trigger: 'change' }],
  contactName: [{ required: true, message: '请输入联系人姓名', trigger: 'blur' }],
  contactPhone: [{ required: true, message: '请输入联系电话', trigger: 'blur' }],
  userId: [{ required: true, message: '请选择会员用户', trigger: 'change' }]
}

const filteredTrips = computed(() => {
  if (!searchKeyword.value) return trips.value
  const keyword = searchKeyword.value.toLowerCase()
  return trips.value.filter(trip =>
    trip.name.toLowerCase().includes(keyword) ||
    trip.destination.toLowerCase().includes(keyword)
  )
})

async function loadTrips() {
  loading.value = true
  try {
    const res = await getTrips()
    trips.value = res.data || []
  } finally {
    loading.value = false
  }
}

function openCreateDialog() {
  isEdit.value = false
  currentTrip.value = null
  Object.assign(tripForm, {
    name: '',
    description: '',
    destination: '',
    start_date: '',
    end_date: '',
    price: 0,
    total_spots: 10,
    status: 'active'
  })
  tripDialogVisible.value = true
}

function openEditDialog(trip) {
  isEdit.value = true
  currentTrip.value = trip
  Object.assign(tripForm, {
    name: trip.name,
    description: trip.description,
    destination: trip.destination,
    start_date: new Date(trip.start_date),
    end_date: new Date(trip.end_date),
    price: trip.price,
    total_spots: trip.total_spots,
    status: trip.status
  })
  tripDialogVisible.value = true
}

async function submitTripForm() {
  if (!tripFormRef.value) return
  await tripFormRef.value.validate(async (valid) => {
    if (!valid) return
    submitting.value = true
    try {
      const data = {
        name: tripForm.name,
        description: tripForm.description,
        destination: tripForm.destination,
        start_date: formatDateForAPI(tripForm.start_date),
        end_date: formatDateForAPI(tripForm.end_date),
        price: tripForm.price,
        total_spots: tripForm.total_spots,
        status: tripForm.status
      }

      if (isEdit.value) {
        await updateTrip(currentTrip.value.id, data)
        ElMessage.success('行程更新成功')
      } else {
        await createTrip(data)
        ElMessage.success('行程创建成功')
      }
      tripDialogVisible.value = false
      loadTrips()
    } finally {
      submitting.value = false
    }
  })
}

function handleAdjustSpots(trip) {
  currentTrip.value = trip
  adjustForm.adjust_amount = 0
  adjustForm.reason = ''
  adjustDialogVisible.value = true
}

async function submitAdjustSpots() {
  if (!adjustFormRef.value) return
  await adjustFormRef.value.validate(async (valid) => {
    if (!valid) return
    submitting.value = true
    try {
      await adjustTripSpots(currentTrip.value.id, {
        adjust_amount: adjustForm.adjust_amount,
        reason: adjustForm.reason
      })
      ElMessage.success('余位调整成功')
      adjustDialogVisible.value = false
      loadTrips()
    } finally {
      submitting.value = false
    }
  })
}

function openItineraryDialog(trip) {
  currentTrip.value = trip
  itineraries.value = []
  itineraryDialogVisible.value = true
  loadItineraries()
}

async function loadItineraries() {
  if (!currentTrip.value) return
  itineraryLoading.value = true
  try {
    const res = await getTripItineraries(currentTrip.value.id)
    itineraries.value = res.data || []
  } finally {
    itineraryLoading.value = false
  }
}

function openItineraryForm(item) {
  isItineraryEdit.value = !!item
  currentItinerary.value = item
  if (item) {
    Object.assign(itineraryForm, {
      day_number: item.day_number,
      title: item.title,
      breakfast: item.breakfast || '',
      lunch: item.lunch || '',
      dinner: item.dinner || '',
      accommodation: item.accommodation || '',
      transportation: item.transportation || '',
      activities: item.activities || '',
      notes: item.notes || ''
    })
  } else {
    Object.assign(itineraryForm, {
      day_number: itineraries.value.length + 1,
      title: '',
      breakfast: '',
      lunch: '',
      dinner: '',
      accommodation: '',
      transportation: '',
      activities: '',
      notes: ''
    })
  }
  itineraryFormVisible.value = true
}

async function submitItinerary() {
  if (!itineraryFormRef.value) return
  await itineraryFormRef.value.validate(async (valid) => {
    if (!valid) return
    submitting.value = true
    try {
      const data = {
        day_number: itineraryForm.day_number,
        title: itineraryForm.title,
        breakfast: itineraryForm.breakfast,
        lunch: itineraryForm.lunch,
        dinner: itineraryForm.dinner,
        accommodation: itineraryForm.accommodation,
        transportation: itineraryForm.transportation,
        activities: itineraryForm.activities,
        notes: itineraryForm.notes
      }

      if (isItineraryEdit.value) {
        await updateItinerary(currentTrip.value.id, currentItinerary.value.id, data)
        ElMessage.success('行程安排更新成功')
      } else {
        await createItinerary(currentTrip.value.id, data)
        ElMessage.success('行程安排新增成功')
      }
      itineraryFormVisible.value = false
      loadItineraries()
    } finally {
      submitting.value = false
    }
  })
}

async function deleteItinerary(item) {
  await ElMessageBox.confirm('确认删除该行程安排？', '删除确认', {
    confirmButtonText: '确认删除',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    await apiDeleteItinerary(currentTrip.value.id, item.id)
    ElMessage.success('删除成功')
    loadItineraries()
  }).catch(() => {})
}

async function openLogsDialog(trip) {
  currentTrip.value = trip
  spotLogs.value = []
  logsDialogVisible.value = true
  logsLoading.value = true
  try {
    const res = await getSpotLogs(trip.id)
    spotLogs.value = res.data || []
  } finally {
    logsLoading.value = false
  }
}

async function handleDeleteTrip(trip) {
  await ElMessageBox.confirm(
    `确认删除行程「${trip.name}」？删除后无法恢复。`,
    '删除确认',
    {
      confirmButtonText: '确认删除',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(async () => {
    await deleteTrip(trip.id)
    ElMessage.success('删除成功')
    loadTrips()
  }).catch(() => {})
}

function handleAction(cmd, trip) {
  switch (cmd) {
    case 'edit':
      openEditDialog(trip)
      break
    case 'adjust':
      handleAdjustSpots(trip)
      break
    case 'itinerary':
      openItineraryDialog(trip)
      break
    case 'logs':
      openLogsDialog(trip)
      break
    case 'detail':
      handleDetail(trip)
      break
    case 'delete':
      handleDeleteTrip(trip)
      break
  }
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

async function loadUserList() {
  try {
    const res = await getUsers()
    userList.value = res.data || []
  } catch (e) {
    userList.value = []
  }
}

function handleBook(trip) {
  currentTrip.value = trip
  bookForm.bookMode = 'guest'
  bookForm.userId = null
  bookForm.travelers = 1
  bookForm.contactName = ''
  bookForm.contactPhone = ''
  loadUserList()
  bookDialogVisible.value = true
}

async function submitBook() {
  if (!bookFormRef.value) return
  await bookFormRef.value.validate(async (valid) => {
    if (!valid) return
    if (bookForm.bookMode === 'user' && !bookForm.userId) {
      ElMessage.warning('请选择会员用户')
      return
    }
    submitting.value = true
    try {
      const orderData = {
        trip_id: currentTrip.value.id,
        trip_name: currentTrip.value.name,
        trip_price: currentTrip.value.price,
        travelers: bookForm.travelers,
        total_amount: bookForm.travelers * currentTrip.value.price,
        contact_name: bookForm.contactName,
        contact_phone: bookForm.contactPhone
      }
      if (bookForm.bookMode === 'user' && bookForm.userId) {
        orderData.user_id = bookForm.userId
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

function formatDate(date) {
  if (!date) return '-'
  return new Date(date).toLocaleDateString('zh-CN').replace(/\//g, '-')
}

function formatDateForAPI(date) {
  if (!date) return ''
  const d = new Date(date)
  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

function getDayType(day) {
  const types = ['primary', 'success', 'warning', 'danger', 'info']
  return types[(day - 1) % types.length]
}

function adjustLogText(type) {
  const map = {
    init: '初始化',
    increase_total: '增加总名额',
    decrease_total: '减少总名额',
    manual_increase: '手动增加',
    manual_decrease: '手动减少',
    refund_return: '退款归还',
    refund_return_batch: '批量退款归还',
    order_consume: '订单扣减'
  }
  return map[type] || type
}

function adjustLogType(type) {
  if (type.includes('increase') || type.includes('init') || type.includes('return')) return 'success'
  if (type.includes('decrease') || type.includes('consume')) return 'warning'
  return 'info'
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
  position: relative;
}
.trip-actions {
  position: absolute;
  top: 8px;
  right: 8px;
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
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
.search-filters {
  display: flex;
  gap: 12px;
}
.danger-text {
  color: #f56c6c !important;
}
.itinerary-dialog :deep(.el-dialog__body) {
  padding-top: 8px;
}
</style>
