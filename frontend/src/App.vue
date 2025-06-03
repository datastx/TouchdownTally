<template>
  <v-app>
    <!-- Show login form if not authenticated -->
    <LoginForm 
      v-if="!authStore.isAuthenticated"
      @login-success="handleLoginSuccess"
    />
    
    <!-- Main app if authenticated -->
    <template v-else>
      <v-app-bar
        app
        color="primary"
        dark
        flat
      >
        <v-app-bar-nav-icon @click="drawer = !drawer"></v-app-bar-nav-icon>
        <v-toolbar-title class="text-h5 font-weight-bold">
          <v-icon left>mdi-football</v-icon>
          TouchdownTally
        </v-toolbar-title>
        <v-spacer></v-spacer>
        
        <!-- User menu -->
        <v-menu offset-y>
          <template v-slot:activator="{ props }">
            <v-btn icon v-bind="props">
              <v-avatar size="32">
                <v-icon>mdi-account</v-icon>
              </v-avatar>
            </v-btn>
          </template>
          <v-list>
            <v-list-item>
              <v-list-item-title>{{ authStore.user?.display_name || 'User' }}</v-list-item-title>
              <v-list-item-subtitle>{{ authStore.user?.email }}</v-list-item-subtitle>
            </v-list-item>
            <v-divider></v-divider>
            <v-list-item @click="logout">
              <template v-slot:prepend>
                <v-icon>mdi-logout</v-icon>
              </template>
              <v-list-item-title>Logout</v-list-item-title>
            </v-list-item>
          </v-list>
        </v-menu>
        
        <v-btn icon @click="toggleTheme" class="ml-2">
          <v-icon>{{ $vuetify.theme.global.name === 'dark' ? 'mdi-weather-sunny' : 'mdi-weather-night' }}</v-icon>
        </v-btn>
      </v-app-bar>

      <v-navigation-drawer
        v-model="drawer"
        app
        temporary
      >
        <v-list>
          <v-list-item
            v-for="item in menuItems"
            :key="item.title"
            :to="item.route"
            link
          >
            <template v-slot:prepend>
              <v-icon>{{ item.icon }}</v-icon>
            </template>
            <v-list-item-title>{{ item.title }}</v-list-item-title>
          </v-list-item>
        </v-list>
      </v-navigation-drawer>

      <v-main>
        <v-container fluid>
          <!-- Router view will go here when we add Vue Router -->
          <DashboardView />
        </v-container>
      </v-main>

      <v-footer app>
        <v-spacer></v-spacer>
        <span>&copy; {{ new Date().getFullYear() }} TouchdownTally - Family Football Pool</span>
        <v-spacer></v-spacer>
      </v-footer>
    </template>
  </v-app>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useTheme } from 'vuetify'
import { useAuthStore } from '@/stores'
import DashboardView from './components/DashboardView.vue'
import LoginForm from './components/LoginForm.vue'

const theme = useTheme()
const authStore = useAuthStore()
const drawer = ref(false)

const menuItems = [
  { title: 'Dashboard', icon: 'mdi-view-dashboard', route: '/' },
  { title: 'My Picks', icon: 'mdi-clipboard-list', route: '/picks' },
  { title: 'Pools', icon: 'mdi-account-group', route: '/pools' },
  { title: 'Standings', icon: 'mdi-trophy', route: '/standings' },
  { title: 'Games', icon: 'mdi-football', route: '/games' },
  { title: 'Chat', icon: 'mdi-chat', route: '/chat' },
]

const toggleTheme = () => {
  theme.global.name.value = theme.global.current.value.dark ? 'light' : 'dark'
}

const handleLoginSuccess = () => {
  console.log('Login successful, showing dashboard')
}

const logout = () => {
  authStore.logout()
}

onMounted(() => {
  // Initialize auth state from localStorage
  authStore.initializeAuth()
})
</script>

<style>
.v-application {
  font-family: 'Roboto', sans-serif !important;
}
</style>
