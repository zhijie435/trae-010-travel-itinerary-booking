import axios from 'axios'
import { ElMessage } from 'element-plus'

const service = axios.create({
  baseURL: '/api',
  timeout: 10000
})

service.interceptors.response.use(
  response => {
    return response.data
  },
  error => {
    console.error('Request error:', error)
    ElMessage.error(error.response?.data?.error || '请求失败，请稍后重试')
    return Promise.reject(error)
  }
)

export default service
