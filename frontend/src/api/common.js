import request from '../utils/request'

export function getTrips() {
  return request({
    url: '/trips',
    method: 'get'
  })
}

export function getUsers() {
  return request({
    url: '/users',
    method: 'get'
  })
}

export function seedData() {
  return request({
    url: '/seed',
    method: 'get'
  })
}
