import { createRouter, createWebHashHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    component: () => import('@/layouts/DefaultLayout.vue'),
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/Dashboard.vue'),
        meta: { title: '工作台' }
      },
      {
        path: 'orders',
        name: 'Orders',
        component: () => import('@/views/Orders.vue'),
        meta: { title: '订单管理' }
      },
      {
        path: 'refunds',
        name: 'Refunds',
        component: () => import('@/views/Refunds.vue'),
        meta: { title: '退款管理' }
      },
      {
        path: 'trips',
        name: 'Trips',
        component: () => import('@/views/Trips.vue'),
        meta: { title: '行程管理' }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  document.title = to.meta.title ? `${to.meta.title} - 旅游行程售后系统` : '旅游行程售后系统'
  next()
})

export default router
