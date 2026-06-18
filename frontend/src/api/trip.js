import request from '../utils/request'

export function getTrips() {
  return request({
    url: '/trips',
    method: 'get'
  })
}

export function getTrip(id) {
  return request({
    url: `/trips/${id}`,
    method: 'get'
  })
}

export function createTrip(data) {
  return request({
    url: '/trips',
    method: 'post',
    data
  })
}

export function updateTrip(id, data) {
  return request({
    url: `/trips/${id}`,
    method: 'put',
    data
  })
}

export function deleteTrip(id) {
  return request({
    url: `/trips/${id}`,
    method: 'delete'
  })
}

export function adjustTripSpots(id, data) {
  return request({
    url: `/trips/${id}/adjust-spots`,
    method: 'post',
    data
  })
}

export function getTripItineraries(id) {
  return request({
    url: `/trips/${id}/itineraries`,
    method: 'get'
  })
}

export function createItinerary(tripId, data) {
  return request({
    url: `/trips/${tripId}/itineraries`,
    method: 'post',
    data
  })
}

export function updateItinerary(tripId, itineraryId, data) {
  return request({
    url: `/trips/${tripId}/itineraries/${itineraryId}`,
    method: 'put',
    data
  })
}

export function deleteItinerary(tripId, itineraryId) {
  return request({
    url: `/trips/${tripId}/itineraries/${itineraryId}`,
    method: 'delete'
  })
}

export function getSpotLogs(id) {
  return request({
    url: `/trips/${id}/spot-logs`,
    method: 'get'
  })
}
