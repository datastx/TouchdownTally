<template>
  <v-container fluid class="fill-height">
    <v-row align="center" justify="center">
      <v-col cols="12" sm="8" md="6" lg="4">
        <v-card class="elevation-12">
          <v-card-title class="text-center pa-8">
            <div class="text-h3 font-weight-bold primary--text">
              <v-icon large left color="primary">mdi-football</v-icon>
              TouchdownTally
            </div>
            <div class="text-subtitle-1 mt-2">Family Football Pool</div>
          </v-card-title>

          <v-card-text class="px-8 pb-8">
            <v-form @submit.prevent="handleSubmit" ref="form">
              <v-tabs v-model="tab" centered>
                <v-tab value="login">Login</v-tab>
                <v-tab value="register">Register</v-tab>
              </v-tabs>

              <v-window v-model="tab" class="mt-6">
                <!-- Login Tab -->
                <v-window-item value="login">
                  <v-text-field
                    v-model="loginForm.email"
                    label="Email"
                    type="email"
                    prepend-inner-icon="mdi-email"
                    variant="outlined"
                    :rules="[rules.required, rules.email]"
                    required
                  ></v-text-field>

                  <v-text-field
                    v-model="loginForm.password"
                    label="Password"
                    :type="showPassword ? 'text' : 'password'"
                    prepend-inner-icon="mdi-lock"
                    :append-inner-icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'"
                    @click:append-inner="showPassword = !showPassword"
                    variant="outlined"
                    :rules="[rules.required]"
                    required
                  ></v-text-field>

                  <v-btn
                    type="submit"
                    color="primary"
                    size="large"
                    block
                    class="mt-4"
                    :loading="loading"
                    @click="currentAction = 'login'"
                  >
                    Login
                  </v-btn>
                </v-window-item>

                <!-- Register Tab -->
                <v-window-item value="register">
                  <v-text-field
                    v-model="registerForm.email"
                    label="Email"
                    type="email"
                    prepend-inner-icon="mdi-email"
                    variant="outlined"
                    :rules="[rules.required, rules.email]"
                    required
                  ></v-text-field>

                  <v-text-field
                    v-model="registerForm.username"
                    label="Username"
                    prepend-inner-icon="mdi-account"
                    variant="outlined"
                    :rules="[rules.required, rules.username]"
                    required
                  ></v-text-field>

                  <v-text-field
                    v-model="registerForm.display_name"
                    label="Display Name"
                    prepend-inner-icon="mdi-card-account-details"
                    variant="outlined"
                    :rules="[rules.required]"
                    required
                  ></v-text-field>

                  <v-text-field
                    v-model="registerForm.password"
                    label="Password"
                    :type="showPassword ? 'text' : 'password'"
                    prepend-inner-icon="mdi-lock"
                    :append-inner-icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'"
                    @click:append-inner="showPassword = !showPassword"
                    variant="outlined"
                    :rules="[rules.required, rules.password]"
                    required
                  ></v-text-field>

                  <v-text-field
                    v-model="registerForm.confirmPassword"
                    label="Confirm Password"
                    :type="showPassword ? 'text' : 'password'"
                    prepend-inner-icon="mdi-lock-check"
                    variant="outlined"
                    :rules="[rules.required, rules.passwordMatch]"
                    required
                  ></v-text-field>

                  <v-btn
                    type="submit"
                    color="primary"
                    size="large"
                    block
                    class="mt-4"
                    :loading="loading"
                    @click="currentAction = 'register'"
                  >
                    Register
                  </v-btn>
                </v-window-item>
              </v-window>
            </v-form>

            <!-- Demo Login Button -->
            <v-divider class="my-6"></v-divider>
            <v-btn
              color="secondary"
              variant="outlined"
              block
              @click="demoLogin"
              :loading="loading"
            >
              <v-icon left>mdi-account-star</v-icon>
              Demo Login
            </v-btn>
          </v-card-text>
        </v-card>

        <!-- Error/Success Messages -->
        <v-alert
          v-if="message"
          :type="messageType"
          class="mt-4"
          closable
          @click:close="message = ''"
        >
          {{ message }}
        </v-alert>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useAuthStore } from '@/stores'

const emit = defineEmits(['login-success'])

const authStore = useAuthStore()

const tab = ref('login')
const showPassword = ref(false)
const loading = ref(false)
const currentAction = ref('login')
const message = ref('')
const messageType = ref('success')
const form = ref(null)

const loginForm = reactive({
  email: '',
  password: ''
})

const registerForm = reactive({
  email: '',
  username: '',
  display_name: '',
  password: '',
  confirmPassword: ''
})

const rules = {
  required: (value) => !!value || 'This field is required',
  email: (value) => {
    const pattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
    return pattern.test(value) || 'Invalid email'
  },
  username: (value) => {
    const pattern = /^[a-zA-Z0-9_]{3,20}$/
    return pattern.test(value) || 'Username must be 3-20 characters, letters, numbers, and underscores only'
  },
  password: (value) => {
    return value.length >= 6 || 'Password must be at least 6 characters'
  },
  passwordMatch: (value) => {
    return value === registerForm.password || 'Passwords do not match'
  }
}

const handleSubmit = async () => {
  const { valid } = await form.value.validate()
  if (!valid) return

  loading.value = true
  message.value = ''

  try {
    if (currentAction.value === 'login') {
      const result = await authStore.login(loginForm)
      if (result.success) {
        message.value = 'Login successful!'
        messageType.value = 'success'
        emit('login-success')
      } else {
        message.value = result.error
        messageType.value = 'error'
      }
    } else {
      const result = await authStore.register({
        email: registerForm.email,
        username: registerForm.username,
        display_name: registerForm.display_name,
        password: registerForm.password
      })
      
      if (result.success) {
        message.value = 'Registration successful! Please login.'
        messageType.value = 'success'
        tab.value = 'login'
        // Reset form
        Object.keys(registerForm).forEach(key => {
          registerForm[key] = ''
        })
      } else {
        message.value = result.error
        messageType.value = 'error'
      }
    }
  } catch (error) {
    message.value = 'An unexpected error occurred'
    messageType.value = 'error'
    console.error('Auth error:', error)
  } finally {
    loading.value = false
  }
}

const demoLogin = async () => {
  loading.value = true
  message.value = ''

  // Demo credentials
  const demoCredentials = {
    email: 'demo@touchdowntally.com',
    password: 'demo123'
  }

  try {
    const result = await authStore.login(demoCredentials)
    if (result.success) {
      message.value = 'Demo login successful!'
      messageType.value = 'success'
      emit('login-success')
    } else {
      message.value = 'Demo login failed. Please try manual login.'
      messageType.value = 'warning'
    }
  } catch (error) {
    message.value = 'Demo login not available. Please register or login manually.'
    messageType.value = 'info'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.fill-height {
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.v-card {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
}
</style>
