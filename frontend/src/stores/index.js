import { ref, computed, reactive } from 'vue'
import { authAPI, wsService } from './api.js'

// Auth store
export const useAuthStore = () => {
  const user = ref(null)
  const token = ref(localStorage.getItem('authToken'))
  const isAuthenticated = computed(() => !!token.value)
  const loading = ref(false)

  const login = async (credentials) => {
    loading.value = true
    try {
      const response = await authAPI.login(credentials)
      const { token: authToken, user: userData } = response.data
      
      token.value = authToken
      user.value = userData
      
      localStorage.setItem('authToken', authToken)
      localStorage.setItem('user', JSON.stringify(userData))
      
      // Connect WebSocket
      wsService.connect(authToken)
      
      return { success: true }
    } catch (error) {
      console.error('Login error:', error)
      return { 
        success: false, 
        error: error.response?.data?.message || 'Login failed' 
      }
    } finally {
      loading.value = false
    }
  }

  const register = async (userData) => {
    loading.value = true
    try {
      const response = await authAPI.register(userData)
      return { success: true, data: response.data }
    } catch (error) {
      console.error('Registration error:', error)
      return { 
        success: false, 
        error: error.response?.data?.message || 'Registration failed' 
      }
    } finally {
      loading.value = false
    }
  }

  const logout = () => {
    token.value = null
    user.value = null
    localStorage.removeItem('authToken')
    localStorage.removeItem('user')
    wsService.disconnect()
  }

  const initializeAuth = () => {
    const storedUser = localStorage.getItem('user')
    if (storedUser && token.value) {
      try {
        user.value = JSON.parse(storedUser)
        wsService.connect(token.value)
      } catch (error) {
        console.error('Error parsing stored user:', error)
        logout()
      }
    }
  }

  return {
    user,
    token,
    isAuthenticated,
    loading,
    login,
    register,
    logout,
    initializeAuth
  }
}

// Game store
export const useGameStore = () => {
  const games = ref([])
  const currentWeekGames = ref([])
  const loading = ref(false)
  const currentWeek = ref(1)
  const currentSeason = ref(2024)

  const fetchGames = async (params = {}) => {
    loading.value = true
    try {
      const { gamesAPI } = await import('./api.js')
      const response = await gamesAPI.getAll(params)
      games.value = response.data
      return { success: true, data: response.data }
    } catch (error) {
      console.error('Error fetching games:', error)
      return { success: false, error: error.message }
    } finally {
      loading.value = false
    }
  }

  const fetchCurrentWeekGames = async () => {
    loading.value = true
    try {
      const { gamesAPI } = await import('./api.js')
      const response = await gamesAPI.getCurrent()
      currentWeekGames.value = response.data
      return { success: true, data: response.data }
    } catch (error) {
      console.error('Error fetching current week games:', error)
      return { success: false, error: error.message }
    } finally {
      loading.value = false
    }
  }

  return {
    games,
    currentWeekGames,
    loading,
    currentWeek,
    currentSeason,
    fetchGames,
    fetchCurrentWeekGames
  }
}

// Picks store
export const usePicksStore = () => {
  const picks = ref([])
  const userPicks = reactive(new Map())
  const loading = ref(false)

  const fetchMyPicks = async (params = {}) => {
    loading.value = true
    try {
      const { picksAPI } = await import('./api.js')
      const response = await picksAPI.getMyPicks(params)
      picks.value = response.data
      
      // Update userPicks map for quick lookup
      response.data.forEach(pick => {
        userPicks.set(pick.game_id, pick)
      })
      
      return { success: true, data: response.data }
    } catch (error) {
      console.error('Error fetching picks:', error)
      return { success: false, error: error.message }
    } finally {
      loading.value = false
    }
  }

  const makePick = async (pickData) => {
    try {
      const { picksAPI } = await import('./api.js')
      const response = await picksAPI.createPick(pickData)
      
      // Update local state
      const newPick = response.data
      picks.value.push(newPick)
      userPicks.set(newPick.game_id, newPick)
      
      return { success: true, data: newPick }
    } catch (error) {
      console.error('Error making pick:', error)
      return { success: false, error: error.message }
    }
  }

  const updatePick = async (pickId, pickData) => {
    try {
      const { picksAPI } = await import('./api.js')
      const response = await picksAPI.updatePick(pickId, pickData)
      
      // Update local state
      const updatedPick = response.data
      const index = picks.value.findIndex(p => p.pick_id === pickId)
      if (index !== -1) {
        picks.value[index] = updatedPick
        userPicks.set(updatedPick.game_id, updatedPick)
      }
      
      return { success: true, data: updatedPick }
    } catch (error) {
      console.error('Error updating pick:', error)
      return { success: false, error: error.message }
    }
  }

  const getPickForGame = (gameId) => {
    return userPicks.get(gameId)
  }

  return {
    picks,
    userPicks,
    loading,
    fetchMyPicks,
    makePick,
    updatePick,
    getPickForGame
  }
}

// Pools store
export const usePoolsStore = () => {
  const pools = ref([])
  const currentPool = ref(null)
  const poolMembers = ref([])
  const loading = ref(false)

  const fetchPools = async () => {
    loading.value = true
    try {
      const { poolsAPI } = await import('./api.js')
      const response = await poolsAPI.getAll()
      pools.value = response.data
      return { success: true, data: response.data }
    } catch (error) {
      console.error('Error fetching pools:', error)
      return { success: false, error: error.message }
    } finally {
      loading.value = false
    }
  }

  const joinPool = async (poolId, joinData = {}) => {
    try {
      const { poolsAPI } = await import('./api.js')
      const response = await poolsAPI.join(poolId, joinData)
      
      // Refresh pools list
      await fetchPools()
      
      return { success: true, data: response.data }
    } catch (error) {
      console.error('Error joining pool:', error)
      return { success: false, error: error.message }
    }
  }

  const createPool = async (poolData) => {
    try {
      const { poolsAPI } = await import('./api.js')
      const response = await poolsAPI.create(poolData)
      
      // Add to local pools
      pools.value.push(response.data)
      
      return { success: true, data: response.data }
    } catch (error) {
      console.error('Error creating pool:', error)
      return { success: false, error: error.message }
    }
  }

  return {
    pools,
    currentPool,
    poolMembers,
    loading,
    fetchPools,
    joinPool,
    createPool
  }
}

// Standings store
export const useStandingsStore = () => {
  const standings = ref([])
  const userStats = ref(null)
  const loading = ref(false)

  const fetchStandings = async (poolId, params = {}) => {
    loading.value = true
    try {
      const { standingsAPI } = await import('./api.js')
      const response = await standingsAPI.getPoolStandings(poolId, params)
      standings.value = response.data
      return { success: true, data: response.data }
    } catch (error) {
      console.error('Error fetching standings:', error)
      return { success: false, error: error.message }
    } finally {
      loading.value = false
    }
  }

  const fetchUserStats = async (userId, poolId) => {
    try {
      const { standingsAPI } = await import('./api.js')
      const response = await standingsAPI.getUserStats(userId, poolId)
      userStats.value = response.data
      return { success: true, data: response.data }
    } catch (error) {
      console.error('Error fetching user stats:', error)
      return { success: false, error: error.message }
    }
  }

  return {
    standings,
    userStats,
    loading,
    fetchStandings,
    fetchUserStats
  }
}

// Chat store
export const useChatStore = () => {
  const messages = ref([])
  const loading = ref(false)

  const fetchMessages = async (poolId, params = {}) => {
    loading.value = true
    try {
      const { chatAPI } = await import('./api.js')
      const response = await chatAPI.getMessages(poolId, params)
      messages.value = response.data
      return { success: true, data: response.data }
    } catch (error) {
      console.error('Error fetching messages:', error)
      return { success: false, error: error.message }
    } finally {
      loading.value = false
    }
  }

  const sendMessage = async (messageData) => {
    try {
      const { chatAPI } = await import('./api.js')
      const response = await chatAPI.sendMessage(messageData)
      
      // Add to local messages
      messages.value.push(response.data)
      
      return { success: true, data: response.data }
    } catch (error) {
      console.error('Error sending message:', error)
      return { success: false, error: error.message }
    }
  }

  const addMessage = (message) => {
    messages.value.push(message)
  }

  // WebSocket message handler
  const handleNewMessage = (data) => {
    if (data.type === 'chat_message') {
      addMessage(data.message)
    }
  }

  return {
    messages,
    loading,
    fetchMessages,
    sendMessage,
    addMessage,
    handleNewMessage
  }
}
