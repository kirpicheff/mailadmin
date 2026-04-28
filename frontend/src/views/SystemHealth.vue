<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import axios from 'axios'

const { t } = useI18n()
const stats = ref(null)
const loading = ref(true)
const error = ref(null)
let timer = null

const fetchStats = async () => {
  try {
    const response = await axios.get('/api/system/health')
    stats.value = response.data
    error.value = null
  } catch (err) {
    console.error('Failed to fetch system health:', err)
    error.value = err.response?.data?.error || 'Failed to load system data'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchStats()
  timer = setInterval(fetchStats, 30000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})

const getStatusClass = (status) => {
  status = status.toUpperCase()
  if (status === 'RUNNING' || status === 'UP') return 'status-running'
  if (status === 'STOPPED' || status === 'FATAL' || status === 'EXITED') return 'status-danger'
  return 'status-warning'
}
</script>

<template>
  <div class="space-y-8">
    <!-- Header -->
    <div class="flex justify-between items-end">
      <div>
        <h1 class="text-4xl font-black text-slate-900 dark:text-white tracking-tight">
          {{ t('system.title') }}
        </h1>
        <p class="text-slate-500 dark:text-slate-400 mt-2 font-medium">
          {{ t('system.subtitle') }}
        </p>
      </div>
      <div v-if="stats" class="flex items-center gap-4 text-sm font-bold">
        <div class="px-4 py-2 bg-white dark:bg-slate-900 rounded-2xl border border-slate-200 dark:border-slate-800 shadow-sm flex items-center gap-2">
          <span class="text-slate-400 uppercase text-[10px] tracking-wider">{{ t('system.uptime') }}:</span>
          <span class="text-mail-blue-600 dark:text-mail-blue-400">{{ stats.uptime }}</span>
        </div>
        <div class="px-4 py-2 bg-mail-blue-600 text-white rounded-2xl shadow-lg shadow-mail-blue-500/20">
          {{ stats.hostname }}
        </div>
      </div>
    </div>

    <div v-if="loading && !stats" class="grid grid-cols-4 gap-6">
      <div v-for="i in 8" :key="i" class="h-32 bg-slate-100 dark:bg-slate-800 animate-pulse rounded-3xl"></div>
    </div>

    <div v-else-if="error" class="p-8 bg-red-50 dark:bg-red-900/10 border border-red-100 dark:border-red-900/30 rounded-3xl text-center">
      <div class="text-4xl mb-4">⚠️</div>
      <h3 class="text-lg font-bold text-red-800 dark:text-red-400">{{ t('common.error') }}</h3>
      <p class="text-red-600/70 dark:text-red-400/60 mt-1">{{ error }}</p>
      <button @click="fetchStats" class="mt-6 px-6 py-2 bg-red-600 text-white rounded-xl font-bold hover:bg-red-700 transition-colors">
        Try Again
      </button>
    </div>

    <div v-else-if="stats" class="space-y-6">
      <!-- Info Banner -->
      <div 
        class="p-5 rounded-[2rem] border transition-all duration-500 flex items-center gap-4"
        :class="!stats.ram_total ? 'bg-amber-50 dark:bg-amber-900/20 border-amber-200 dark:border-amber-800/50 text-amber-800 dark:text-amber-400' : 'bg-blue-50 dark:bg-blue-900/10 border-blue-100 dark:border-blue-900/20 text-blue-800 dark:text-blue-400'"
      >
        <div class="w-10 h-10 rounded-full bg-white dark:bg-slate-900 flex items-center justify-center shadow-sm shrink-0">
          <svg v-if="!stats.ram_total" xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-amber-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
          </svg>
          <svg v-else xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </div>
        <p class="text-xs font-semibold leading-relaxed">
          {{ t('system.local_server_warning') }}
        </p>
      </div>

      <!-- Main Stats Grid -->
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <!-- RAM -->
        <div class="health-card">
          <div class="flex justify-between items-start mb-4">
            <h3 class="health-card-title">{{ t('system.ram') }}</h3>
            <div class="w-8 h-8 rounded-lg bg-indigo-50 dark:bg-indigo-900/30 flex items-center justify-center text-indigo-600">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                <path fill-rule="evenodd" d="M2 5a2 2 0 012-2h12a2 2 0 012 2v10a2 2 0 01-2 2H4a2 2 0 01-2-2V5zm3 1h10v8H5V6z" clip-rule="evenodd" />
              </svg>
            </div>
          </div>
          <div class="text-3xl font-black mb-1">{{ stats.ram_used }} <span class="text-sm font-medium text-slate-400">MB</span></div>
          <p class="text-[11px] text-slate-500 mb-4">{{ stats.ram_perc }}% of {{ stats.ram_total }}MB used</p>
          <div class="h-2 w-full bg-slate-100 dark:bg-slate-800 rounded-full overflow-hidden">
            <div 
              class="h-full transition-all duration-1000 bg-gradient-to-r from-indigo-500 to-purple-500"
              :style="{ width: stats.ram_perc + '%', backgroundColor: stats.ram_perc > 85 ? '#ef4444' : '' }"
            ></div>
          </div>
        </div>

        <!-- DISK -->
        <div class="health-card">
          <div class="flex justify-between items-start mb-4">
            <h3 class="health-card-title">{{ t('system.disk') }}</h3>
            <div class="w-8 h-8 rounded-lg bg-emerald-50 dark:bg-emerald-900/30 flex items-center justify-center text-emerald-600">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                <path d="M7 3a1 1 0 000 2h6a1 1 0 100-2H7zM4 7a1 1 0 011-1h10a1 1 0 110 2H5a1 1 0 01-1-1zM2 11a2 2 0 012-2h12a2 2 0 012 2v4a2 2 0 01-2 2H4a2 2 0 01-2-2v-4z" />
              </svg>
            </div>
          </div>
          <div class="text-3xl font-black mb-1">{{ stats.disk_used }} <span class="text-sm font-medium text-slate-400">/ {{ stats.disk_total }}</span></div>
          <p class="text-[11px] text-slate-500 mb-4">{{ stats.disk_perc }}% utilized</p>
          <div class="h-2 w-full bg-slate-100 dark:bg-slate-800 rounded-full overflow-hidden">
            <div 
              class="h-full transition-all duration-1000 bg-gradient-to-r from-emerald-500 to-teal-500"
              :style="{ width: stats.disk_perc + '%', backgroundColor: stats.disk_perc > 90 ? '#ef4444' : '' }"
            ></div>
          </div>
        </div>

        <!-- SSL -->
        <div class="health-card">
          <div class="flex justify-between items-start mb-4">
            <h3 class="health-card-title">{{ t('system.ssl') }}</h3>
            <div class="w-8 h-8 rounded-lg bg-amber-50 dark:bg-amber-900/30 flex items-center justify-center text-amber-600">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                <path fill-rule="evenodd" d="M2.166 4.999A11.954 11.954 0 0010 1.944 11.954 11.954 0 0017.834 5c.11.65.166 1.32.166 2.001 0 4.946-2.567 9.29-6.433 11.754a.75.75 0 01-.834 0C6.9 16.29 4.333 11.946 4.333 7c0-.68.056-1.35.166-2.001zm11.53 3.32a.75.75 0 00-1.06-1.06l-3.386 3.385-1.06-1.06a.75.75 0 00-1.06 1.06l1.59 1.59a.75.75 0 001.06 0l3.916-3.915z" clip-rule="evenodd" />
              </svg>
            </div>
          </div>
          <div 
            class="text-3xl font-black mb-1"
            :class="stats.ssl_days < 10 ? 'text-red-500' : 'text-emerald-500 dark:text-emerald-400'"
          >
            {{ stats.ssl_days }} <span class="text-sm font-medium text-slate-400">{{ t('system.ssl_days') }}</span>
          </div>
          <p class="text-[11px] text-slate-500">{{ stats.ssl_days > 0 ? 'Auto-renewal active' : 'Check certificates' }}</p>
        </div>

        <!-- Fail2Ban -->
        <div class="health-card">
          <div class="flex justify-between items-start mb-4">
            <h3 class="health-card-title">{{ t('system.f2b') }}</h3>
            <div class="w-8 h-8 rounded-lg bg-red-50 dark:bg-red-900/30 flex items-center justify-center text-red-600">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                <path fill-rule="evenodd" d="M10 1.944A11.954 11.954 0 012.166 5C2.056 5.649 2 6.319 2 7c0 5.189 3.011 9.635 7.376 11.743.407.197.841.197 1.248 0C14.989 16.635 18 12.189 18 7c0-.681-.056-1.351-.166-2.001A11.954 11.954 0 0110 1.944zM11 14a1 1 0 11-2 0 1 1 0 012 0zm0-7a1 1 0 10-2 0v3a1 1 0 102 0V7z" clip-rule="evenodd" />
              </svg>
            </div>
          </div>
          <div 
            class="text-3xl font-black mb-1"
            :class="stats.f2b_count > 0 ? 'text-amber-500' : ''"
          >
            {{ stats.f2b_count }} <span class="text-sm font-medium text-slate-400">{{ t('system.banned_ip') }}</span>
          </div>
          <p class="text-[11px] text-slate-500">Security: Postfix / Dovecot</p>
        </div>
      </div>

      <!-- Secondary Stats Row -->
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <div class="health-card">
          <h3 class="health-card-title mb-2">{{ t('system.load') }}</h3>
          <div class="text-2xl font-bold">{{ stats.load }}</div>
          <p class="text-xs text-slate-500 mt-1">CPU Load (1 min)</p>
        </div>

        <div class="health-card">
          <h3 class="health-card-title mb-2">{{ t('system.imap') }}</h3>
          <div class="text-2xl font-bold">{{ stats.imap_sessions }}</div>
          <p class="text-xs text-slate-500 mt-1">{{ t('system.active_connections') }}</p>
        </div>

        <div class="health-card">
          <h3 class="health-card-title mb-2">{{ t('system.db_redis') }}</h3>
          <div class="text-2xl font-bold flex items-baseline gap-2">
            {{ stats.db_threads }} <span class="text-xs font-medium text-slate-400">SQL</span>
            <span class="text-slate-200 dark:text-slate-800">|</span>
            {{ stats.redis_memory }} <span class="text-xs font-medium text-slate-400">RDS</span>
          </div>
          <p class="text-xs text-slate-500 mt-1">Active threads & memory</p>
        </div>

        <div class="health-card">
          <h3 class="health-card-title mb-2">{{ t('system.queue') }}</h3>
          <div 
            class="text-2xl font-bold"
            :class="stats.queue > 50 ? 'text-red-500' : ''"
          >
            {{ stats.queue }}
          </div>
          <p class="text-xs text-slate-500 mt-1">{{ t('system.mail_queue') }}</p>
        </div>
      </div>

      <!-- Services List -->
      <div class="grid grid-cols-1 gap-6">
        <div class="health-card p-0 overflow-hidden">
          <div class="p-6 border-b border-slate-100 dark:border-slate-800">
            <h3 class="health-card-title">{{ t('system.services') }} (Supervisor)</h3>
          </div>
          <div class="p-6">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-x-12 gap-y-1">
              <div v-for="srv in stats.services" :key="srv.name" class="flex justify-between items-center py-3 border-b border-slate-50 dark:border-slate-800 last:border-0">
                <div class="flex items-center gap-3">
                  <span class="w-10 h-10 rounded-xl bg-slate-50 dark:bg-slate-800/50 flex items-center justify-center">
                    <div class="w-2 h-2 rounded-full shadow-lg" :class="getStatusClass(srv.status)"></div>
                  </span>
                  <span class="font-bold text-sm text-slate-700 dark:text-slate-300">{{ srv.name }}</span>
                </div>
                <div class="text-[10px] items-center flex gap-2 font-black uppercase px-3 py-1 rounded-lg" :class="getStatusClass(srv.status) + '-bg'">
                  {{ srv.status }}
                  <span class="text-slate-400 lowercase font-normal">{{ srv.info.split(',')[0] }}</span>
                </div>
              </div>
            </div>
            <div v-if="!stats.services || stats.services.length === 0" class="text-center py-12 text-slate-400">
              No services managed by Supervisor found.
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
@reference "../style.css";

.health-card {
  @apply bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-[2.5rem] p-7 transition-all duration-300 hover:shadow-2xl hover:shadow-slate-200 dark:hover:shadow-black hover:-translate-y-1;
}

.health-card-title {
  @apply text-[10px] font-black uppercase tracking-widest text-slate-400 dark:text-slate-500;
}

.status-running { @apply bg-emerald-500 shadow-emerald-500/50; }
.status-danger { @apply bg-red-500 shadow-red-500/50; }
.status-warning { @apply bg-amber-500 shadow-amber-500/50; }

.status-running-bg { @apply bg-emerald-50 text-emerald-700 dark:bg-emerald-900/20 dark:text-emerald-400; }
.status-danger-bg { @apply bg-red-50 text-red-700 dark:bg-red-900/20 dark:text-red-400; }
.status-warning-bg { @apply bg-amber-50 text-amber-700 dark:bg-amber-900/20 dark:text-amber-400; }
</style>
