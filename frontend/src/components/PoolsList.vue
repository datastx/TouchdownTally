<template>
  <v-card>
    <v-card-title class="d-flex justify-space-between align-center">
      <span class="text-h5">My Pools</span>
      <CreatePoolDialog @pool-created="handlePoolCreated" />
    </v-card-title>
    
    <v-card-text>
      <v-progress-linear v-if="loading" indeterminate></v-progress-linear>
      
      <div v-else-if="pools.length === 0" class="text-center py-8">
        <v-icon size="64" color="grey-lighten-1">mdi-account-group-outline</v-icon>
        <p class="text-h6 mt-4 mb-2">No pools yet</p>
        <p class="text-body-2 text-grey">Create your first pool to get started!</p>
      </div>
      
      <v-row v-else>
        <v-col 
          v-for="pool in pools" 
          :key="pool.id" 
          cols="12" 
          md="6" 
          lg="4"
        >
          <v-card variant="outlined" hover>
            <v-card-title class="text-h6">
              {{ pool.name }}
              <v-chip 
                v-if="pool.user_role === 'commissioner'"
                color="primary" 
                size="small" 
                class="ml-2"
              >
                Commissioner
              </v-chip>
            </v-card-title>
            
            <v-card-text>
              <div class="d-flex justify-space-between mb-2">
                <span class="text-body-2 text-grey">Type:</span>
                <span class="text-body-2 text-capitalize">{{ pool.pool_type }}</span>
              </div>
              
              <div class="d-flex justify-space-between mb-2">
                <span class="text-body-2 text-grey">Entry Fee:</span>
                <span class="text-body-2">${{ pool.entry_fee }}</span>
              </div>
              
              <div class="d-flex justify-space-between mb-2">
                <span class="text-body-2 text-grey">Members:</span>
                <span class="text-body-2">{{ pool.current_members }}/{{ pool.max_players }}</span>
              </div>
              
              <div class="d-flex justify-space-between mb-2">
                <span class="text-body-2 text-grey">Season:</span>
                <span class="text-body-2">{{ pool.season }}</span>
              </div>
              
              <div class="d-flex justify-space-between">
                <span class="text-body-2 text-grey">Status:</span>
                <v-chip 
                  :color="pool.is_active === 'active' ? 'success' : 'warning'"
                  size="small"
                >
                  {{ pool.is_active }}
                </v-chip>
              </div>
            </v-card-text>
            
            <v-card-actions>
              <v-btn 
                color="primary" 
                variant="text"
                @click="viewPool(pool)"
              >
                View Details
              </v-btn>
              <v-spacer></v-spacer>
              <v-btn 
                icon
                size="small"
                @click="refreshPool(pool.id)"
              >
                <v-icon>mdi-refresh</v-icon>
              </v-btn>
            </v-card-actions>
          </v-card>
        </v-col>
      </v-row>
    </v-card-text>
  </v-card>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores'
import api from '@/services/api'
import CreatePoolDialog from './CreatePoolDialog.vue'

const authStore = useAuthStore()
const pools = ref([])
const loading = ref(false)

const fetchPools = async () => {
  loading.value = true
  try {
    const response = await api.get('/pools')
    if (response.data.success) {
      // Combine available and member pools
      const memberPools = response.data.data.member_pools || []
      const availablePools = response.data.data.available_pools || []
      pools.value = [...memberPools, ...availablePools]
    }
  } catch (error) {
    console.error('Error fetching pools:', error.response?.data || error.message)
  } finally {
    loading.value = false
  }
}

const handlePoolCreated = (newPool) => {
  pools.value.unshift(newPool)
  console.log('New pool added to list:', newPool.name)
}

const viewPool = (pool) => {
  console.log('Viewing pool:', pool)
  // TODO: Navigate to pool details or open modal
}

const refreshPool = async (poolId) => {
  try {
    const response = await api.get(`/pools/${poolId}`)
    if (response.data.success) {
      const poolIndex = pools.value.findIndex(p => p.id === poolId)
      if (poolIndex !== -1) {
        pools.value[poolIndex] = response.data.data
      }
    }
  } catch (error) {
    console.error('Error refreshing pool:', error.response?.data || error.message)
  }
}

onMounted(() => {
  fetchPools()
})
</script>
