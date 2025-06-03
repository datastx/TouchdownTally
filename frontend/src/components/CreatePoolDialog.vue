<template>
  <v-dialog v-model="dialog" max-width="600px">
    <template v-slot:activator="{ props }">
      <v-btn
        color="primary"
        v-bind="props"
        prepend-icon="mdi-plus"
        size="large"
      >
        Create New Pool
      </v-btn>
    </template>

    <v-card>
      <v-card-title>
        <span class="text-h5">Create New Pool</span>
      </v-card-title>
      
      <v-card-text>
        <v-container>
          <v-form ref="form" v-model="valid">
            <v-row>
              <v-col cols="12">
                <v-text-field
                  v-model="poolData.pool_name"
                  label="Pool Name"
                  :rules="[rules.required]"
                  required
                ></v-text-field>
              </v-col>
              
              <v-col cols="12" sm="6">
                <v-text-field
                  v-model.number="poolData.season_year"
                  label="Season Year"
                  type="number"
                  :rules="[rules.required]"
                  required
                ></v-text-field>
              </v-col>
              
              <v-col cols="12" sm="6">
                <v-text-field
                  v-model.number="poolData.entry_fee"
                  label="Entry Fee ($)"
                  type="number"
                  step="0.01"
                  :rules="[rules.required]"
                  required
                ></v-text-field>
              </v-col>
              
              <v-col cols="12" sm="6">
                <v-text-field
                  v-model.number="poolData.max_members"
                  label="Max Members"
                  type="number"
                  :rules="[rules.required]"
                  required
                ></v-text-field>
              </v-col>
              
              <v-col cols="12" sm="6">
                <v-select
                  v-model="poolData.pool_type"
                  :items="poolTypes"
                  label="Pool Type"
                  :rules="[rules.required]"
                  required
                ></v-select>
              </v-col>
              
              <v-col cols="12">
                <v-textarea
                  v-model="poolData.description"
                  label="Description (Optional)"
                  rows="3"
                ></v-textarea>
              </v-col>
            </v-row>
          </v-form>
        </v-container>
      </v-card-text>
      
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn color="grey" variant="text" @click="close">
          Cancel
        </v-btn>
        <v-btn 
          color="primary" 
          :loading="loading"
          :disabled="!valid"
          @click="createPool"
        >
          Create Pool
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useAuthStore } from '@/stores'
import api from '@/services/api'

const emit = defineEmits(['pool-created'])

const authStore = useAuthStore()
const dialog = ref(false)
const valid = ref(false)
const loading = ref(false)
const form = ref(null)

const poolTypes = [
  { title: 'Weekly Pick\'em', value: 'weekly' },
  { title: 'Survivor', value: 'survivor' },
  { title: 'Confidence Points', value: 'confidence' }
]

const poolData = reactive({
  pool_name: '',
  season_year: new Date().getFullYear(),
  entry_fee: 25.00,
  max_members: 10,
  pool_type: 'weekly',
  description: '',
  prize_structure: {
    first: 70,
    second: 20,
    third: 10
  },
  settings: {
    allow_ties: true,
    deadline_hours: 2
  }
})

const rules = {
  required: value => !!value || 'This field is required',
}

const createPool = async () => {
  if (!valid.value) return
  
  loading.value = true
  try {
    const response = await api.post('/pools', poolData)
    
    if (response.data.success) {
      emit('pool-created', response.data.data)
      close()
      // Show success message
      console.log('Pool created successfully:', response.data.data)
    }
  } catch (error) {
    console.error('Error creating pool:', error.response?.data || error.message)
    // Show error message
  } finally {
    loading.value = false
  }
}

const close = () => {
  dialog.value = false
  // Reset form
  if (form.value) {
    form.value.reset()
  }
  // Reset data
  Object.assign(poolData, {
    pool_name: '',
    season_year: new Date().getFullYear(),
    entry_fee: 25.00,
    max_members: 10,
    pool_type: 'weekly',
    description: '',
    prize_structure: {
      first: 70,
      second: 20,
      third: 10
    },
    settings: {
      allow_ties: true,
      deadline_hours: 2
    }
  })
}
</script>
