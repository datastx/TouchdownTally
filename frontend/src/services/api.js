import axios from 'axios'

// Create axios instance with base configuration
const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8080/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor to add auth token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('authToken')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor to handle errors
api.interceptors.response.use(
  (response) => {
    return response
  },
  (error) => {
    if (error.response?.status === 401) {
      // Handle unauthorized - redirect to login
      localStorage.removeItem('authToken')
      localStorage.removeItem('user')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

// Auth API
export const authAPI = {
  login: (credentials) => api.post('/auth/login', credentials),
  register: (userData) => api.post('/auth/register', userData),
  refreshToken: () => api.post('/auth/refresh'),
  logout: () => api.post('/auth/logout'),
}

// Teams API
export const teamsAPI = {
  getAll: () => api.get('/teams'),
  getById: (id) => api.get(`/teams/${id}`),
}

// Games API
export const gamesAPI = {
  getAll: (params) => api.get('/games', { params }),
  getById: (id) => api.get(`/games/${id}`),
  getByWeek: (week, season) => api.get(`/games/week/${week}`, { params: { season } }),
  getCurrent: () => api.get('/games/current'),
}

// Picks API
export const picksAPI = {
  getMyPicks: (params) => api.get('/picks', { params }),
  createPick: (pickData) => api.post('/picks', pickData),
  updatePick: (pickId, pickData) => api.put(`/picks/${pickId}`, pickData),
  deletePick: (pickId) => api.delete(`/picks/${pickId}`),
  getPicksByUser: (userId) => api.get(`/picks/user/${userId}`),
  getPicksByGame: (gameId) => api.get(`/picks/game/${gameId}`),
}

// Pools API
export const poolsAPI = {
  getAll: () => api.get('/pools'),
  getById: (id) => api.get(`/pools/${id}`),
  create: (poolData) => api.post('/pools', poolData),
  join: (poolId, joinData) => api.post(`/pools/${poolId}/join`, joinData),
  leave: (poolId) => api.post(`/pools/${poolId}/leave`),
  getMembers: (poolId) => api.get(`/pools/${poolId}/members`),
  updateSettings: (poolId, settings) => api.put(`/pools/${poolId}/settings`, settings),
}

// Standings API
export const standingsAPI = {
  getPoolStandings: (poolId, params) => api.get(`/standings/pool/${poolId}`, { params }),
  getUserStats: (userId, poolId) => api.get(`/standings/user/${userId}`, { params: { pool_id: poolId } }),
  getWeeklyStandings: (poolId, week) => api.get(`/standings/pool/${poolId}/week/${week}`),
}

// Chat API
export const chatAPI = {
  getMessages: (poolId, params) => api.get(`/chat/pool/${poolId}`, { params }),
  sendMessage: (messageData) => api.post('/chat/send', messageData),
  getHistory: (poolId, params) => api.get(`/chat/pool/${poolId}/history`, { params }),
}

// WebSocket connection for real-time features
export class WebSocketService {
  constructor() {
    this.socket = null
    this.listeners = new Map()
    this.reconnectAttempts = 0
    this.maxReconnectAttempts = 5
    this.reconnectDelay = 1000
  }

  connect(token) {
    const wsUrl = import.meta.env.VITE_WS_URL || 'ws://localhost:8080/ws'
    const url = `${wsUrl}?token=${token}`
    
    this.socket = new WebSocket(url)
    
    this.socket.onopen = () => {
      console.log('WebSocket connected')
      this.reconnectAttempts = 0
    }
    
    this.socket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        this.handleMessage(data)
      } catch (error) {
        console.error('Error parsing WebSocket message:', error)
      }
    }
    
    this.socket.onclose = () => {
      console.log('WebSocket disconnected')
      this.attemptReconnect(token)
    }
    
    this.socket.onerror = (error) => {
      console.error('WebSocket error:', error)
    }
  }

  disconnect() {
    if (this.socket) {
      this.socket.close()
      this.socket = null
    }
    this.listeners.clear()
  }

  attemptReconnect(token) {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++
      setTimeout(() => {
        console.log(`Attempting to reconnect (${this.reconnectAttempts}/${this.maxReconnectAttempts})`)
        this.connect(token)
      }, this.reconnectDelay * this.reconnectAttempts)
    }
  }

  send(message) {
    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      this.socket.send(JSON.stringify(message))
    }
  }

  subscribe(type, callback) {
    if (!this.listeners.has(type)) {
      this.listeners.set(type, [])
    }
    this.listeners.get(type).push(callback)
  }

  unsubscribe(type, callback) {
    if (this.listeners.has(type)) {
      const callbacks = this.listeners.get(type)
      const index = callbacks.indexOf(callback)
      if (index > -1) {
        callbacks.splice(index, 1)
      }
    }
  }

  handleMessage(data) {
    const { type } = data
    if (this.listeners.has(type)) {
      this.listeners.get(type).forEach(callback => callback(data))
    }
  }
}

// Export singleton instance
export const wsService = new WebSocketService()

// Default export
export default api
