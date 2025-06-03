<template>
  <div>
    <v-row>
      <v-col cols="12">
        <v-card class="mb-4">
          <v-card-title class="text-h4 text-center py-6">
            <v-icon large left color="primary">mdi-football</v-icon>
            Welcome to TouchdownTally
          </v-card-title>
          <v-card-text class="text-center">
            <p class="text-h6 mb-4">Your Family Football Pool Dashboard</p>
            <p class="text-body-1">Track your picks, compete with family, and follow the action!</p>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <!-- Pools Section -->
    <v-row>
      <v-col cols="12">
        <PoolsList />
      </v-col>
    </v-row>

    <v-row>
      <!-- Quick Stats -->
      <v-col cols="12" md="3">
        <v-card color="primary" dark>
          <v-card-text>
            <div class="text-h3 font-weight-bold">{{ stats.totalPicks }}</div>
            <div class="text-body-1">Total Picks</div>
          </v-card-text>
          <v-card-text>
            <v-icon large>mdi-clipboard-check</v-icon>
          </v-card-text>
        </v-card>
      </v-col>

      <v-col cols="12" md="3">
        <v-card color="success" dark>
          <v-card-text>
            <div class="text-h3 font-weight-bold">{{ stats.correctPicks }}</div>
            <div class="text-body-1">Correct Picks</div>
          </v-card-text>
          <v-card-text>
            <v-icon large>mdi-check-circle</v-icon>
          </v-card-text>
        </v-card>
      </v-col>

      <v-col cols="12" md="3">
        <v-card color="info" dark>
          <v-card-text>
            <div class="text-h3 font-weight-bold">{{ stats.winPercentage }}%</div>
            <div class="text-body-1">Win Rate</div>
          </v-card-text>
          <v-card-text>
            <v-icon large>mdi-percent</v-icon>
          </v-card-text>
        </v-card>
      </v-col>

      <v-col cols="12" md="3">
        <v-card color="warning" dark>
          <v-card-text>
            <div class="text-h3 font-weight-bold">{{ stats.currentRank }}</div>
            <div class="text-body-1">Current Rank</div>
          </v-card-text>
          <v-card-text>
            <v-icon large>mdi-trophy</v-icon>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <v-row class="mt-4">
      <!-- This Week's Games -->
      <v-col cols="12" lg="8">
        <v-card>
          <v-card-title>
            <v-icon left>mdi-football</v-icon>
            This Week's Games
          </v-card-title>
          <v-card-text>
            <v-data-table
              :headers="gameHeaders"
              :items="upcomingGames"
              :loading="loading"
              class="elevation-1"
            >
              <template v-slot:item.game_time="{ item }">
                {{ formatGameTime(item.game_time) }}
              </template>
              <template v-slot:item.pick="{ item }">
                <v-chip
                  v-if="item.pick"
                  :color="item.pick === item.home_team ? 'primary' : 'secondary'"
                  small
                >
                  {{ item.pick }}
                </v-chip>
                <v-btn
                  v-else
                  color="primary"
                  size="small"
                  @click="makePick(item)"
                >
                  Make Pick
                </v-btn>
              </template>
            </v-data-table>
          </v-card-text>
        </v-card>
      </v-col>

      <!-- Recent Activity -->
      <v-col cols="12" lg="4">
        <v-card>
          <v-card-title>
            <v-icon left>mdi-timeline</v-icon>
            Recent Activity
          </v-card-title>
          <v-card-text>
            <v-timeline dense>
              <v-timeline-item
                v-for="activity in recentActivity"
                :key="activity.id"
                :color="activity.color"
                small
              >
                <template v-slot:opposite>
                  <span class="text-caption">{{ formatTime(activity.time) }}</span>
                </template>
                <div class="text-body-2">{{ activity.message }}</div>
              </v-timeline-item>
            </v-timeline>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <v-row class="mt-4">
      <!-- Pool Standings Preview -->
      <v-col cols="12">
        <v-card>
          <v-card-title>
            <v-icon left>mdi-trophy</v-icon>
            Pool Standings
            <v-spacer></v-spacer>
            <v-btn color="primary" variant="text" @click="viewFullStandings">
              View All
              <v-icon right>mdi-arrow-right</v-icon>
            </v-btn>
          </v-card-title>
          <v-card-text>
            <v-data-table
              :headers="standingsHeaders"
              :items="topStandings"
              :items-per-page="5"
              class="elevation-1"
            >
              <template v-slot:item.rank="{ item }">
                <v-chip
                  :color="getRankColor(item.rank)"
                  small
                >
                  #{{ item.rank }}
                </v-chip>
              </template>
            </v-data-table>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import PoolsList from './PoolsList.vue'

// Reactive data
const loading = ref(false)
const stats = ref({
  totalPicks: 0,
  correctPicks: 0,
  winPercentage: 0,
  currentRank: 0
})

const upcomingGames = ref([])
const recentActivity = ref([])
const topStandings = ref([])

// Table headers
const gameHeaders = [
  { title: 'Time', key: 'game_time' },
  { title: 'Away Team', key: 'away_team' },
  { title: 'Home Team', key: 'home_team' },
  { title: 'Your Pick', key: 'pick' },
]

const standingsHeaders = [
  { title: 'Rank', key: 'rank' },
  { title: 'Player', key: 'display_name' },
  { title: 'Correct', key: 'correct_picks' },
  { title: 'Total', key: 'total_picks' },
  { title: 'Win %', key: 'win_percentage' },
]

// Mock data - will be replaced with API calls
onMounted(() => {
  loadDashboardData()
})

const loadDashboardData = () => {
  // Mock stats
  stats.value = {
    totalPicks: 47,
    correctPicks: 32,
    winPercentage: 68,
    currentRank: 3
  }

  // Mock upcoming games
  upcomingGames.value = [
    {
      id: 1,
      game_time: '2025-06-08T17:00:00Z',
      away_team: 'Bills',
      home_team: 'Patriots',
      pick: 'Bills'
    },
    {
      id: 2,
      game_time: '2025-06-08T20:30:00Z',
      away_team: 'Cowboys',
      home_team: 'Giants',
      pick: null
    },
    {
      id: 3,
      game_time: '2025-06-09T13:00:00Z',
      away_team: 'Packers',
      home_team: 'Bears',
      pick: 'Packers'
    }
  ]

  // Mock recent activity
  recentActivity.value = [
    {
      id: 1,
      message: 'Made pick: Bills over Patriots',
      time: '2025-06-03T10:30:00Z',
      color: 'success'
    },
    {
      id: 2,
      message: 'Joined Moore Family Pool',
      time: '2025-06-02T15:45:00Z',
      color: 'info'
    },
    {
      id: 3,
      message: 'Correct pick: Chiefs over Raiders',
      time: '2025-06-01T22:15:00Z',
      color: 'success'
    }
  ]

  // Mock standings
  topStandings.value = [
    {
      rank: 1,
      display_name: 'Sarah M.',
      correct_picks: 38,
      total_picks: 47,
      win_percentage: 81
    },
    {
      rank: 2,
      display_name: 'Mike R.',
      correct_picks: 35,
      total_picks: 47,
      win_percentage: 74
    },
    {
      rank: 3,
      display_name: 'You',
      correct_picks: 32,
      total_picks: 47,
      win_percentage: 68
    },
    {
      rank: 4,
      display_name: 'Dad',
      correct_picks: 30,
      total_picks: 47,
      win_percentage: 64
    },
    {
      rank: 5,
      display_name: 'Mom',
      correct_picks: 28,
      total_picks: 47,
      win_percentage: 60
    }
  ]
}

// Utility functions
const formatGameTime = (gameTime) => {
  return new Date(gameTime).toLocaleDateString('en-US', {
    weekday: 'short',
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit'
  })
}

const formatTime = (time) => {
  return new Date(time).toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit'
  })
}

const getRankColor = (rank) => {
  if (rank === 1) return 'yellow'
  if (rank <= 3) return 'grey'
  return 'blue-grey'
}

const makePick = (game) => {
  console.log('Making pick for game:', game)
  // TODO: Implement pick making logic
}

const viewFullStandings = () => {
  console.log('Viewing full standings')
  // TODO: Navigate to standings page
}
</script>

<style scoped>
.v-card {
  transition: transform 0.2s;
}

.v-card:hover {
  transform: translateY(-2px);
}
</style>
