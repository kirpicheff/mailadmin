<script setup>
import { ref, onMounted } from 'vue'
import api from '@/api/axios'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const stats = ref({
  domains_count: 0,
  mailboxes_count: 0,
  aliases_count: 0,
  quota_limit: 0,
  quota_used: 0,
  recent_logs: []
})

const loading = ref(true)

const formatBytes = (bytes) => {
  if (!bytes || bytes === 0) return '0 Bytes'
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const translateAction = (action) => {
  const dictionary = {
    'create_domain': t('actions.create_domain'),
    'update_domain': t('actions.update_domain'),
    'delete_domain': t('actions.delete_domain'),
    'create_mailbox': t('actions.create_mailbox'),
    'update_mailbox': t('actions.update_mailbox'),
    'delete_mailbox': t('actions.delete_mailbox'),
    'create_alias': t('actions.create_alias'),
    'update_alias': t('actions.update_alias'),
    'delete_alias': t('actions.delete_alias'),
    'create_admin': t('actions.create_admin'),
    'delete_admin': t('actions.delete_admin'),
    'login': t('actions.login')
  }
  return dictionary[action.toLowerCase()] || action
}

const formatDate = (dateStr) => {
  return new Date(dateStr).toLocaleString()
}

const fetchStats = async () => {
  try {
    const response = await api.get('/stats/dashboard')
    stats.value = response.data
  } catch (error) {
    console.error('Failed to fetch stats:', error)
  } finally {
    loading.value = false
  }
}

onMounted(fetchStats)
</script>

<template>
  <div class="space-y-8 animate-in fade-in duration-500">
    <header class="flex flex-col gap-2">
      <h1 class="text-3xl font-bold tracking-tight text-slate-900 dark:text-white">{{ t('dashboard.title') }}</h1>
      <p class="text-slate-500 dark:text-slate-400">{{ t('dashboard.subtitle') }}</p>
    </header>

    <!-- Основные показатели -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
      <!-- Домены -->
      <div class="glass-panel p-6 group hover:border-mail-blue-500/50 transition-colors">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider">{{ t('dashboard.stats.total_domains') }}</p>
            <p class="text-3xl font-bold mt-2 text-slate-900 dark:text-white">{{ stats.domains_count || 0 }}</p>
          </div>
          <div class="w-12 h-12 rounded-2xl bg-mail-blue-500/10 flex items-center justify-center text-mail-blue-500 group-hover:scale-110 transition-transform">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9h18" />
            </svg>
          </div>
        </div>
      </div>

      <!-- Ящики -->
      <div class="glass-panel p-6 group hover:border-emerald-500/50 transition-colors">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider">{{ t('dashboard.stats.mailboxes') }}</p>
            <p class="text-3xl font-bold mt-2 text-slate-900 dark:text-white">{{ stats.mailboxes_count || 0 }}</p>
          </div>
          <div class="w-12 h-12 rounded-2xl bg-emerald-500/10 flex items-center justify-center text-emerald-500 group-hover:scale-110 transition-transform">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
            </svg>
          </div>
        </div>
      </div>

      <!-- Алиасы -->
      <div class="glass-panel p-6 group hover:border-amber-500/50 transition-colors">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider">{{ t('dashboard.stats.aliases') }}</p>
            <p class="text-3xl font-bold mt-2 text-slate-900 dark:text-white">{{ stats.aliases_count || 0 }}</p>
          </div>
          <div class="w-12 h-12 rounded-2xl bg-amber-500/10 flex items-center justify-center text-amber-500 group-hover:scale-110 transition-transform">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
            </svg>
          </div>
        </div>
      </div>

      <!-- Квота (Общая) -->
      <div class="glass-panel p-6 group hover:border-purple-500/50 transition-colors">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider">{{ t('dashboard.stats.disk_usage') }}</p>
            <p class="text-3xl font-bold mt-2 text-slate-900 dark:text-white text-base">
              {{ formatBytes(stats.quota_used) }}
            </p>
          </div>
          <div class="w-12 h-12 rounded-2xl bg-purple-500/10 flex items-center justify-center text-purple-500 group-hover:scale-110 transition-transform">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
            </svg>
          </div>
        </div>
        <div class="mt-4">
          <div class="flex justify-between text-[10px] mb-1 font-bold text-slate-500 uppercase tracking-tighter">
            <span>{{ stats.quota_limit > 0 ? Math.round((stats.quota_used / stats.quota_limit) * 100) : 0 }}% {{ t('dashboard.stats.used_of') }}</span>
            <span>{{ stats.quota_limit > 0 ? formatBytes(stats.quota_limit) : t('common.unlimited') }}</span>
          </div>
          <div class="w-full bg-slate-100 dark:bg-slate-800 rounded-full h-1.5 overflow-hidden">
            <div 
              class="bg-purple-500 h-full rounded-full transition-all duration-1000" 
              :style="{ width: (stats.quota_limit > 0 ? (stats.quota_used / stats.quota_limit) * 100 : 0) + '%' }"
            ></div>
          </div>
        </div>
      </div>
    </div>

    <!-- Таблица последних действий -->
    <div class="grid grid-cols-1 gap-8">
      <div class="glass-panel p-0 overflow-hidden flex flex-col min-h-[400px]">
        <div class="p-6 border-b border-slate-100 dark:border-slate-800 flex justify-between items-center bg-slate-50/50 dark:bg-slate-800/20">
          <h2 class="font-bold text-lg">{{ t('dashboard.recent_actions.title') }}</h2>
          <router-link to="/logs" class="text-xs font-bold text-mail-blue-500 hover:bg-mail-blue-500/10 px-4 py-2 rounded-xl transition-colors">
            {{ t('dashboard.recent_actions.view_all') }}
          </router-link>
        </div>
        <div class="flex-1 overflow-y-auto">
          <div v-if="!stats.recent_logs || stats.recent_logs.length === 0" class="h-full flex items-center justify-center p-20 text-slate-400 italic">
            <div class="text-center">
              <svg class="mx-auto h-12 w-12 text-slate-200" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <p class="mt-4">{{ t('dashboard.recent_actions.no_activity') }}</p>
            </div>
          </div>
          <div v-else class="divide-y divide-slate-50 dark:divide-slate-800">
            <div v-for="log in stats.recent_logs || []" :key="log.id" class="p-5 hover:bg-slate-50 dark:hover:bg-slate-800/30 transition-colors">
              <div class="flex items-start gap-5">
                <div class="mt-1">
                  <!-- Иконка в зависимости от действия -->
                  <div v-if="log.action?.includes('create')" class="w-10 h-10 rounded-xl bg-emerald-500/10 text-emerald-500 flex items-center justify-center">
                    <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M12 4v16m8-8H4" /></svg>
                  </div>
                  <div v-else-if="log.action?.includes('delete')" class="w-10 h-10 rounded-xl bg-red-500/10 text-red-500 flex items-center justify-center">
                    <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
                  </div>
                  <div v-else class="w-10 h-10 rounded-xl bg-mail-blue-500/10 text-mail-blue-500 flex items-center justify-center">
                    <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" /></svg>
                  </div>
                </div>
                <div class="flex-1 min-w-0">
                  <div class="flex justify-between items-baseline mb-1">
                    <p class="text-sm font-extrabold text-slate-800 dark:text-slate-100 tracking-tight">
                      {{ log.username }}
                    </p>
                    <span class="text-[10px] text-slate-400 font-black uppercase tracking-widest">{{ formatDate(log.timestamp) }}</span>
                  </div>
                  <p class="text-xs text-slate-600 dark:text-slate-400 leading-relaxed capitalize">
                    {{ translateAction(log.action) }}: <span class="text-slate-500 font-bold decoration-dotted">{{ log.data }}</span>
                  </p>
                  <p class="text-[10px] text-slate-400 mt-1 font-bold lowercase">{{ t('dashboard.recent_actions.object') }}: {{ log.domain }}</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
</style>
