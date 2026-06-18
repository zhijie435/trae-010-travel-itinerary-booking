import request from '../utils/request'

export function getRefundRequests(params) {
  return request({
    url: '/refunds',
    method: 'get',
    params
  })
}

export function getRefundRequest(id) {
  return request({
    url: `/refunds/${id}`,
    method: 'get'
  })
}

export function createRefundRequest(data) {
  return request({
    url: '/refunds',
    method: 'post',
    data
  })
}

export function reviewRefundRequest(id, data) {
  return request({
    url: `/refunds/${id}/review`,
    method: 'post',
    data
  })
}
