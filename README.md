# 旅游行程预订售后退款系统

基于 **Vue 3 + Gin** 构建的旅游行程预订售后退款管理框架。

## 技术栈

### 后端
- **Gin** - Go 语言 HTTP Web 框架
- **GORM** - ORM 库
- **SQLite** - 轻量级数据库（无需单独安装）
- **CORS** - 跨域支持

### 前端
- **Vue 3** - 渐进式 JavaScript 框架
- **Vite** - 下一代前端构建工具
- **Element Plus** - Vue 3 组件库
- **Vue Router** - 路由管理
- **Pinia** - 状态管理
- **Axios** - HTTP 客户端

## 项目结构

```
.
├── backend/                 # 后端 (Gin)
│   ├── config/             # 配置（数据库等）
│   ├── controllers/        # 控制器
│   ├── models/             # 数据模型
│   ├── routes/             # 路由定义
│   ├── main.go             # 入口文件
│   └── go.mod              # Go 模块
├── frontend/               # 前端 (Vue 3)
│   ├── src/
│   │   ├── api/            # API 接口
│   │   ├── layouts/        # 布局组件
│   │   ├── router/         # 路由
│   │   ├── stores/         # 状态管理
│   │   ├── utils/          # 工具函数
│   │   ├── views/          # 页面组件
│   │   ├── App.vue         # 根组件
│   │   └── main.js         # 入口文件
│   ├── index.html          # HTML 模板
│   ├── vite.config.js      # Vite 配置
│   └── package.json        # NPM 包配置
└── start.sh                # 一键启动脚本
```

## 快速开始

### 方式一：一键启动（推荐）

```bash
chmod +x start.sh
./start.sh
```

### 方式二：分别启动

#### 1. 启动后端

```bash
cd backend
go mod tidy          # 安装依赖
go run main.go       # 启动服务 (端口 8080)
```

#### 2. 启动前端（新终端）

```bash
cd frontend
npm install          # 安装依赖
npm run dev          # 启动服务 (端口 5173)
```

## 访问地址

- 前端: http://localhost:5173
- 后端 API: http://localhost:8080

## 功能模块

### 1. 工作台 (Dashboard)
- 订单总数、订单金额、退款申请数、待审核退款数统计
- 最新订单列表
- 退款申请列表

### 2. 行程管理 (Trips)
- 行程卡片列表展示
- 行程详情查看
- 行程预订（创建订单）

### 3. 订单管理 (Orders)
- 订单列表（支持按状态筛选）
- 订单详情查看
- 订单支付
- 申请退款

### 4. 退款管理 (Refunds)
- 退款申请列表（支持按状态筛选）
- 退款详情查看
- 审核退款（通过 / 拒绝）

## 业务流程

### 订单生命周期
```
待支付(pending) → 已支付(paid) → 退款中(refunding) → 已退款(refunded)
                                      ↓
                                   已拒绝 → 恢复为已支付(paid)
```

### 退款流程
1. 用户在已支付订单上点击「申请退款」
2. 填写退款原因、详细说明、退款金额
3. 提交申请后订单状态变为「退款中」
4. 管理员在退款管理中审核：
   - **通过**：订单状态变为「已退款」
   - **拒绝**：订单状态恢复为「已支付」

## API 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/seed` | 初始化测试数据 |
| GET | `/api/trips` | 获取行程列表 |
| GET | `/api/users` | 获取用户列表 |
| GET | `/api/orders` | 获取订单列表（参数：user_id, status） |
| GET | `/api/orders/:id` | 获取订单详情 |
| POST | `/api/orders` | 创建订单 |
| POST | `/api/orders/:id/pay` | 支付订单 |
| GET | `/api/refunds` | 获取退款列表（参数：user_id, status） |
| GET | `/api/refunds/:id` | 获取退款详情 |
| POST | `/api/refunds` | 创建退款申请 |
| POST | `/api/refunds/:id/review` | 审核退款 |

## 状态常量

### 订单状态
| 值 | 说明 |
|----|------|
| `pending` | 待支付 |
| `paid` | 已支付 |
| `confirmed` | 已确认 |
| `refunding` | 退款中 |
| `refunded` | 已退款 |
| `completed` | 已完成 |
| `cancelled` | 已取消 |

### 退款状态
| 值 | 说明 |
|----|------|
| `pending` | 待审核 |
| `approved` | 已通过 |
| `rejected` | 已拒绝 |

### 退款原因
| 值 | 说明 |
|----|------|
| `change_of_plan` | 行程变更 |
| `health_issue` | 身体原因 |
| `financial` | 资金问题 |
| `service_issue` | 服务不满意 |
| `other` | 其他原因 |

## 扩展建议

1. **用户认证**：添加 JWT 登录鉴权
2. **文件上传**：退款凭证上传功能
3. **邮件通知**：退款审核结果邮件通知
4. **退款规则**：根据距离出发时间自动计算退款比例
5. **日志系统**：操作日志、审计日志
6. **数据统计**：退款率、退款原因分析图表
