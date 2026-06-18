# 旅游行程预订系统 - 后端结构

## 领域模型结构

### 1. 线路领域 (Route)
**文件**: [models/route.go](file:///Users/wuzhijie/Documents/xiaohongshu/biaozhu/trae-web-projects/010-旅游行程预订系统/backend/models/route.go)
- **Route**: 线路主表，包含线路基本信息、行程日期、价格
- **Itinerary**: 行程安排表，按天存储每日行程详情

### 2. 库存领域 (Inventory)
**文件**: [models/inventory.go](file:///Users/wuzhijie/Documents/xiaohongshu/biaozhu/trae-web-projects/010-旅游行程预订系统/backend/models/inventory.go)
- **Inventory**: 库存表，按日期维度管理线路余位
  - RouteID: 关联线路
  - Date: 出发日期
  - TotalSpots: 总名额
  - LeftSpots: 剩余名额
- **InventoryAdjustLog**: 库存调整日志，记录所有库存变动

### 3. 订单领域 (Order)
**文件**: [models/order.go](file:///Users/wuzhijie/Documents/xiaohongshu/biaozhu/trae-web-projects/010-旅游行程预订系统/backend/models/order.go)
- **Order**: 订单表
  - RouteID: 关联线路
  - InventoryID: 关联库存（按日期）
  - 包含订单金额、出行人数、联系人信息、支付状态
- **RefundRequest**: 退款申请表
- **RefundReviewLog**: 退款审核日志

### 4. 用户领域 (User)
**文件**: [models/user.go](file:///Users/wuzhijie/Documents/xiaohongshu/biaozhu/trae-web-projects/010-旅游行程预订系统/backend/models/user.go)
- **User**: 用户表

---

## 服务层结构

### RouteService ([services/route_service.go](file:///Users/wuzhijie/Documents/xiaohongshu/biaozhu/trae-web-projects/010-旅游行程预订系统/backend/services/route_service.go))
- `List()` - 线路列表查询
- `GetByID()` - 线路详情
- `Create()` - 创建线路（自动创建对应库存）
- `Update()` - 更新线路（同步更新总名额）
- `Delete()` - 删除线路
- `ListItineraries()` - 行程安排列表
- `CreateItinerary()` - 新增行程安排
- `UpdateItinerary()` - 更新行程安排
- `DeleteItinerary()` - 删除行程安排

### InventoryService ([services/inventory_service.go](file:///Users/wuzhijie/Documents/xiaohongshu/biaozhu/trae-web-projects/010-旅游行程预订系统/backend/services/inventory_service.go))
- `GetByID()` - 库存详情
- `GetByRouteAndDate()` - 按线路和日期查库存
- `ListByRoute()` - 线路库存列表
- `Create()` / `CreateWithTx()` - 创建库存
- `Adjust()` / `AdjustWithTx()` - 调整库存
- `AdjustWithOrder()` - 订单相关库存调整
- `AdjustWithRefund()` - 退款相关库存调整
- `ListAdjustLogs()` - 库存调整日志
- `UpdateTotalSpots()` - 更新总名额

### OrderService ([services/order_service.go](file:///Users/wuzhijie/Documents/xiaohongshu/biaozhu/trae-web-projects/010-旅游行程预订系统/backend/services/order_service.go))
- `List()` - 订单列表
- `GetByID()` - 订单详情
- `Create()` - 创建订单（锁定库存）
- `Pay()` - 订单支付
- `Cancel()` - 取消订单（释放库存）
- `CreateRefund()` - 申请退款
- `ReviewRefund()` - 审核退款
- `BatchReviewRefunds()` - 批量审核退款
- `ListRefunds()` - 退款列表
- `GetRefundByID()` - 退款详情
- `GetRefundReviewLogs()` - 退款审核日志

---

## 控制器结构

- [route_controller.go](file:///Users/wuzhijie/Documents/xiaohongshu/biaozhu/trae-web-projects/010-旅游行程预订系统/backend/controllers/route_controller.go) - 线路相关API
- [inventory_controller.go](file:///Users/wuzhijie/Documents/xiaohongshu/biaozhu/trae-web-projects/010-旅游行程预订系统/backend/controllers/inventory_controller.go) - 库存相关API
- [order_controller.go](file:///Users/wuzhijie/Documents/xiaohongshu/biaozhu/trae-web-projects/010-旅游行程预订系统/backend/controllers/order_controller.go) - 订单/退款相关API
- [user_controller.go](file:///Users/wuzhijie/Documents/xiaohongshu/biaozhu/trae-web-projects/010-旅游行程预订系统/backend/controllers/user_controller.go) - 用户相关API
- [seed_controller.go](file:///Users/wuzhijie/Documents/xiaohongshu/biaozhu/trae-web-projects/010-旅游行程预订系统/backend/controllers/seed_controller.go) - 测试数据初始化

---

## 工具包

- [pkg/utils/helpers.go](file:///Users/wuzhijie/Documents/xiaohongshu/biaozhu/trae-web-projects/010-旅游行程预订系统/backend/pkg/utils/helpers.go)
  - `GenerateOrderNo()` - 生成订单号
  - `GenerateRefundNo()` - 生成退款号
  - `TimeNow()` - 当前时间
  - `ParseDate()` - 日期解析

---

## 核心业务流程

### 线路预订流程
1. **创建线路** → 自动创建对应日期的库存记录
2. **用户下单** → 校验库存 → 锁定库存（扣减LeftSpots）→ 创建订单（状态pending）
3. **订单支付** → 更新订单状态为paid
4. **订单取消** → 释放库存（归还LeftSpots）→ 订单状态cancelled
5. **退款申请** → 订单状态refunding → 创建退款申请
6. **退款审核通过** → 释放库存 → 订单状态refunded
7. **退款审核拒绝** → 订单状态恢复paid

---

## API 路由

### 线路管理
- `GET    /api/routes` - 线路列表
- `GET    /api/routes/:id` - 线路详情
- `POST   /api/routes` - 创建线路
- `PUT    /api/routes/:id` - 更新线路
- `DELETE /api/routes/:id` - 删除线路
- `GET    /api/routes/:id/itineraries` - 行程安排列表
- `POST   /api/routes/:id/itineraries` - 新增行程安排
- `PUT    /api/routes/:id/itineraries/:itinerary_id` - 更新行程安排
- `DELETE /api/routes/:id/itineraries/:itinerary_id` - 删除行程安排

### 库存管理
- `GET    /api/inventories/route/:id` - 线路库存列表
- `GET    /api/inventories/:id` - 库存详情
- `POST   /api/inventories/:id/adjust` - 手动调整库存
- `GET    /api/inventories/route/:id/logs` - 库存调整日志

### 订单管理
- `GET    /api/orders` - 订单列表
- `GET    /api/orders/:id` - 订单详情
- `POST   /api/orders` - 创建订单
- `POST   /api/orders/:id/pay` - 订单支付

### 退款管理
- `GET    /api/refunds` - 退款列表
- `GET    /api/refunds/:id` - 退款详情
- `POST   /api/refunds` - 申请退款
- `POST   /api/refunds/:id/review` - 审核退款
- `POST   /api/refunds/batch-review` - 批量审核
- `GET    /api/refunds/:id/review-logs` - 审核日志

### 兼容路由 (旧 trips 接口)
- `GET    /api/trips` - 线路列表
- `GET    /api/trips/:id` - 线路详情
- `POST   /api/trips` - 创建线路
- `PUT    /api/trips/:id` - 更新线路
- `DELETE /api/trips/:id` - 删除线路
- `POST   /api/trips/:id/adjust-spots` - 调整库存
- `GET    /api/trips/:id/itineraries` - 行程安排列表
- `POST   /api/trips/:id/itineraries` - 新增行程安排
- `PUT    /api/trips/:id/itineraries/:itinerary_id` - 更新行程安排
- `DELETE /api/trips/:id/itineraries/:itinerary_id` - 删除行程安排
- `GET    /api/trips/:id/spot-logs` - 库存调整日志
