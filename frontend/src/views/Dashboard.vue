<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
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

const health = ref(null)
const loading = ref(true)
const bannedIPs = ref([])
const showF2BModal = ref(false)
const processingUnban = ref(null)
let timer = null

const formatBytes = (bytes) => {
  if (!bytes || bytes === 0) return '0 Bytes'
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const translateAction = (action) => {
  if (!action) return ''
  const tKey = action.toLowerCase().replace(/ /g, '_')
  return t(`actions.${tKey}`) || action
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
  }
}

const fetchHealth = async () => {
  try {
    const response = await api.get('/system/health')
    health.value = response.data
  } catch (err) {
    console.log('Metrics unavailable (remote)')
  } finally {
    loading.value = false
  }
}

const fetchBannedIPs = async () => {
  try {
    const response = await api.get('/system/fail2ban')
    bannedIPs.value = response.data || []
    showF2BModal.value = true
  } catch (error) {
    console.error('Failed to fetch banned IPs:', error)
  }
}

const unbanIP = async (ip, jail) => {
  if (!confirm(t('fail2ban.confirm_unban', { ip, jail }))) return
  processingUnban.value = ip
  try {
    await api.delete('/system/fail2ban/unban', { params: { ip, jail } })
    await fetchBannedIPs() // Рефреш списка
    await fetchHealth()    // Рефреш счетчика на дашборде
  } catch (error) {
    alert(t('fail2ban.messages.unban_error'))
  } finally {
    processingUnban.value = null
  }
}

const getStatusClass = (status) => {
  status = status.toUpperCase()
  if (status === 'RUNNING' || status === 'UP') return 'bg-emerald-500 shadow-emerald-500/50'
  if (status === 'STOPPED' || status === 'FATAL' || status === 'EXITED') return 'bg-red-500 shadow-red-500/50'
  return 'bg-amber-500 shadow-amber-500/50'
}

onMounted(() => {
  fetchStats()
  fetchHealth()
  timer = setInterval(() => {
    fetchStats()
    fetchHealth()
  }, 30000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<template>
  <div class="space-y-8 animate-in fade-in duration-500 pb-10">
    <header class="flex justify-between items-end">
      <div class="flex flex-col gap-2">
        <h1 class="text-4xl font-black tracking-tight text-slate-900 dark:text-white">{{ t('dashboard.title') }}</h1>
        <p class="text-slate-500 dark:text-slate-400 font-medium">{{ t('dashboard.subtitle') }}</p>
      </div>
      <div v-if="health" class="hidden md:flex items-center gap-3">
        <div class="px-4 py-2 bg-white dark:bg-slate-900 rounded-2xl border border-slate-100 dark:border-slate-800 shadow-sm">
          <span class="text-[10px] font-black uppercase text-slate-400 tracking-widest mr-2">{{ t('system.uptime') }}:</span>
          <span class="text-xs font-bold text-mail-blue-600">{{ health.uptime }}</span>
        </div>
        <div class="px-4 py-2 bg-mail-blue-600 text-white rounded-2xl font-bold text-xs shadow-lg shadow-mail-blue-500/20">
          {{ health.hostname }}
        </div>
      </div>
    </header>

    <!-- Секция здоровья сервера (только если локально) -->
    <section v-if="health && health.ram_total" class="space-y-6">
      <div class="flex items-center gap-2">
        <div class="w-1.5 h-4 rounded-full bg-mail-blue-500"></div>
        <h2 class="text-xs font-black uppercase tracking-widest text-slate-400">{{ t('system.title') }}</h2>
      </div>
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <!-- RAM -->
        <div class="health-card">
          <div class="flex justify-between mb-2">
            <span class="text-[10px] font-black text-slate-400 uppercase tracking-widest">{{ t('system.ram') }}</span>
            <span class="text-xs font-bold text-slate-900 dark:text-white">{{ health.ram_perc }}%</span>
          </div>
          <div class="text-2xl font-black mb-3 text-slate-900 dark:text-white">{{ health.ram_used }} <span class="text-[10px] text-slate-500 font-medium">MB</span></div>
          <div class="h-1.5 w-full bg-slate-100 dark:bg-slate-800 rounded-full overflow-hidden">
            <div 
              class="h-full transition-all duration-1000"
              :class="health.ram_perc > 80 ? 'bg-red-500' : 'bg-indigo-500'"
              :style="{ width: health.ram_perc + '%' }"
            ></div>
          </div>
        </div>

        <!-- DISK -->
        <div class="health-card">
          <div class="flex justify-between mb-2">
            <span class="text-[10px] font-black text-slate-400 uppercase tracking-widest">{{ t('system.disk') }}</span>
            <span class="text-xs font-bold text-slate-900 dark:text-white">{{ health.disk_perc }}%</span>
          </div>
          <div class="text-2xl font-black mb-3 text-slate-900 dark:text-white">{{ health.disk_used }} <span class="text-[10px] text-slate-500 font-medium">/ {{ health.disk_total }}</span></div>
          <div class="h-1.5 w-full bg-slate-100 dark:bg-slate-800 rounded-full overflow-hidden">
            <div 
              class="h-full transition-all duration-1000"
              :class="health.disk_perc > 90 ? 'bg-red-500' : 'bg-emerald-500'"
              :style="{ width: health.disk_perc + '%' }"
            ></div>
          </div>
        </div>

        <!-- SSL Сертификат -->
        <div class="health-card">
          <div class="flex justify-between mb-2">
            <span class="text-[10px] font-black text-slate-400 uppercase tracking-widest">SSL Сертификат</span>
          </div>
          <div class="text-2xl font-black mb-3" :class="health.ssl_days < 10 ? 'text-red-500' : 'text-slate-900 dark:text-white'">{{ health.ssl_days }}</div>
          <p class="text-[9px] font-bold text-slate-400 uppercase">{{ t('system.ssl_days') }}</p>
        </div>

        <!-- Средняя нагрузка -->
        <div class="health-card">
          <div class="flex justify-between mb-2">
            <span class="text-[10px] font-black text-slate-400 uppercase tracking-widest">Средняя нагрузка</span>
          </div>
          <div class="text-2xl font-black mb-3 text-slate-900 dark:text-white">{{ health.load }}</div>
          <p class="text-[9px] font-bold text-slate-400 uppercase">Нагрузка на CPU (1 мин)</p>
        </div>

        <!-- Защита Fail2Ban -->
        <div class="health-card cursor-pointer hover:border-amber-500/50 group/f2b" @click="fetchBannedIPs">
          <div class="flex justify-between mb-2">
            <span class="text-[10px] font-black text-slate-400 uppercase tracking-widest group-hover/f2b:text-amber-500 transition-colors">Защита Fail2Ban</span>
            <svg class="w-3 h-3 text-slate-300 group-hover/f2b:translate-x-0.5 transition-transform" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M9 5l7 7-7 7" /></svg>
          </div>
          <div class="text-2xl font-black mb-3" :class="health.f2b_count > 0 ? 'text-amber-500' : 'text-slate-900 dark:text-white'">{{ health.f2b_count }}</div>
          <p class="text-[9px] font-bold text-slate-400 uppercase">Забанено IP</p>
        </div>

        <!-- Очередь писем -->
        <div class="health-card">
          <div class="flex justify-between mb-2">
            <span class="text-[10px] font-black text-slate-400 uppercase tracking-widest">Очередь писем</span>
          </div>
          <div class="text-2xl font-black mb-3 text-slate-900 dark:text-white">{{ health.queue }}</div>
          <p class="text-[9px] font-bold text-slate-400 uppercase">Писем в очереди Postfix</p>
        </div>

        <!-- IMAP Сессии -->
        <div class="health-card">
          <div class="flex justify-between mb-2">
            <span class="text-[10px] font-black text-slate-400 uppercase tracking-widest">IMAP Сессии</span>
          </div>
          <div class="text-2xl font-black mb-3 text-slate-900 dark:text-white">{{ health.imap_sessions }}</div>
          <p class="text-[9px] font-bold text-slate-400 uppercase">Активные подключения</p>
        </div>

        <!-- БД & Redis -->
        <div class="health-card">
          <div class="flex justify-between mb-2">
            <span class="text-[10px] font-black text-slate-400 uppercase tracking-widest">БД & Redis</span>
          </div>
          <div class="flex items-baseline gap-2 mb-3">
            <span class="text-2xl font-black text-slate-900 dark:text-white">{{ health.db_threads }}</span>
            <span class="text-[10px] text-slate-400 font-bold uppercase">SQL</span>
            <span class="text-slate-300 dark:text-slate-700 mx-1">|</span>
            <span class="text-2xl font-black text-slate-900 dark:text-white">{{ health.redis_memory }}</span>
            <span class="text-[10px] text-slate-400 font-bold uppercase">RDS</span>
          </div>
          <p class="text-[9px] font-bold text-slate-400 uppercase">Потоки БД и память Redis</p>
        </div>
      </div>

      <!-- Services Small Row -->
      <div v-if="health.services && health.services.length > 0" class="flex flex-wrap gap-4">
        <div v-for="srv in health.services" :key="srv.name" 
             class="flex items-center gap-2 px-3 py-1.5 bg-white dark:bg-slate-900 border border-slate-100 dark:border-slate-800 rounded-xl shadow-sm">
          <div class="w-2 h-2 rounded-full" :class="getStatusClass(srv.status)"></div>
          <span class="text-[10px] font-black uppercase text-slate-600 dark:text-slate-400">{{ srv.name }}</span>
        </div>
      </div>
    </section>

    <!-- Основные показатели -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
      <div class="glass-panel p-6 group hover:border-mail-blue-500/50 transition-colors">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest">{{ t('dashboard.stats.total_domains') }}</p>
            <p class="text-3xl font-black mt-2 text-slate-900 dark:text-white">{{ stats.domains_count || 0 }}</p>
          </div>
          <div class="w-12 h-12 rounded-2xl bg-mail-blue-500/10 flex items-center justify-center text-mail-blue-500 group-hover:rotate-12 transition-transform">
            <svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="2" y1="12" x2="22" y2="12"></line><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"></path></svg>
          </div>
        </div>
      </div>

      <div class="glass-panel p-6 group hover:border-emerald-500/50 transition-colors">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest">{{ t('dashboard.stats.mailboxes') }}</p>
            <p class="text-3xl font-black mt-2 text-slate-900 dark:text-white">{{ stats.mailboxes_count || 0 }}</p>
          </div>
          <div class="w-12 h-12 rounded-2xl bg-emerald-500/10 flex items-center justify-center text-emerald-500 group-hover:rotate-12 transition-transform">
            <svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" /></svg>
          </div>
        </div>
      </div>

      <div class="glass-panel p-6 group hover:border-amber-500/50 transition-colors">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest">{{ t('dashboard.stats.aliases') }}</p>
            <p class="text-3xl font-black mt-2 text-slate-900 dark:text-white">{{ stats.aliases_count || 0 }}</p>
          </div>
          <div class="w-12 h-12 rounded-2xl bg-amber-500/10 flex items-center justify-center text-amber-500 group-hover:rotate-12 transition-transform">
            <svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" /></svg>
          </div>
        </div>
      </div>

      <div class="glass-panel p-6 group hover:border-purple-500/50 transition-colors">
        <div class="flex items-center justify-between mb-4">
          <div>
            <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest">{{ t('dashboard.stats.disk_usage') }}</p>
            <p class="text-2xl font-black mt-1 text-slate-900 dark:text-white">{{ formatBytes(stats.quota_used) }}</p>
          </div>
          <div class="w-12 h-12 rounded-2xl bg-purple-500/10 flex items-center justify-center text-purple-500 group-hover:rotate-12 transition-transform">
            <svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" /></svg>
          </div>
        </div>
        <div class="w-full bg-slate-100 dark:bg-slate-800 rounded-full h-1.5 overflow-hidden">
          <div 
            class="bg-purple-500 h-full rounded-full transition-all duration-1000" 
            :style="{ width: (stats.quota_limit > 0 ? (stats.quota_used / stats.quota_limit) * 100 : 0) + '%' }"
          ></div>
        </div>
        <div class="flex justify-between text-[9px] mt-2 font-black text-slate-400 uppercase tracking-widest">
            <span>{{ stats.quota_limit > 0 ? Math.round((stats.quota_used / stats.quota_limit) * 100) : 0 }}% {{ t('dashboard.stats.used_of') }}</span>
            <span>{{ stats.quota_limit > 0 ? formatBytes(stats.quota_limit) : t('common.unlimited') }}</span>
        </div>
      </div>
    </div>

    <!-- Таблица последних действий -->
    <div class="glass-panel p-0 overflow-hidden flex flex-col">
      <div class="p-6 border-b border-slate-100 dark:border-slate-800 flex justify-between items-center bg-slate-50/50 dark:bg-slate-800/10">
        <div class="flex items-center gap-3">
          <div class="w-1.5 h-4 rounded-full bg-mail-blue-500"></div>
          <h2 class="font-black text-xs uppercase tracking-widest text-slate-400">{{ t('dashboard.recent_actions.title') }}</h2>
        </div>
        <router-link to="/logs" class="text-[10px] font-black uppercase tracking-widest text-mail-blue-500 hover:text-mail-blue-700 transition-colors">
          {{ t('dashboard.recent_actions.view_all') }} →
        </router-link>
      </div>
      <div class="overflow-y-auto max-h-[600px] divide-y divide-slate-50 dark:divide-slate-800/50">
        <div v-if="!stats.recent_logs || stats.recent_logs.length === 0" class="py-20 text-center text-slate-400 italic font-medium">
          {{ t('dashboard.recent_actions.no_activity') }}
        </div>
        <div v-for="log in stats.recent_logs" :key="log.id" class="p-6 hover:bg-slate-50/50 dark:hover:bg-slate-800/20 transition-all group">
          <div class="flex items-center gap-6">
            <div class="w-12 h-12 rounded-2xl bg-slate-100 dark:bg-slate-800 flex items-center justify-center text-slate-400 group-hover:bg-white dark:group-hover:bg-slate-700 shadow-sm transition-colors">
              <svg v-if="log.action?.includes('create')" class="w-6 h-6 text-emerald-500" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M12 4v16m8-8H4" /></svg>
              <svg v-else-if="log.action?.includes('delete')" class="w-6 h-6 text-red-500" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
              <svg v-else class="w-6 h-6 text-mail-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" /></svg>
            </div>
            <div class="flex-1 min-w-0">
              <div class="flex justify-between items-start mb-1">
                <span class="text-sm font-black text-slate-900 dark:text-white tracking-tight break-all uppercase">{{ log.username }}</span>
                <span class="text-[10px] text-slate-400 font-black uppercase tracking-widest">{{ formatDate(log.timestamp) }}</span>
              </div>
              <p class="text-xs text-slate-500 flex items-center gap-2">
                <span class="font-black text-slate-700 dark:text-slate-300 uppercase tracking-tighter">{{ translateAction(log.action) }}</span>
                <span class="w-1 h-1 rounded-full bg-slate-300"></span>
                <span class="font-bold text-mail-blue-600/70">{{ log.data }}</span>
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Fail2Ban Modal -->
    <div v-if="showF2BModal" class="fixed inset-0 z-[100] flex items-center justify-center p-6 animate-in fade-in duration-300">
      <div class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm" @click="showF2BModal = false"></div>
      
      <div class="glass-panel w-full max-w-2xl overflow-hidden relative animate-in zoom-in-95 duration-300">
        <div class="p-8 border-b border-slate-100 dark:border-slate-800 flex justify-between items-center bg-slate-50/50 dark:bg-slate-800/10">
          <div>
            <h2 class="text-2xl font-black text-slate-900 dark:text-white leading-none">{{ t('fail2ban.title') }}</h2>
            <p class="text-xs text-slate-500 font-bold uppercase tracking-widest mt-2">{{ t('fail2ban.banned_ips') }}</p>
          </div>
          <button @click="showF2BModal = false" class="p-2 hover:bg-slate-200 dark:hover:bg-slate-800 rounded-xl transition-colors">
            <svg class="w-6 h-6 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
          </button>
        </div>

        <div class="p-0 max-h-[60vh] overflow-y-auto">
          <table v-if="bannedIPs?.length > 0" class="w-full text-left">
            <thead class="sticky top-0 bg-white/90 dark:bg-slate-900/90 backdrop-blur-md z-10 border-b border-slate-100 dark:border-slate-800">
              <tr>
                <th class="px-8 py-4 text-[10px] font-black uppercase text-slate-400 tracking-widest">{{ t('fail2ban.ip_address') }}</th>
                <th class="px-8 py-4 text-[10px] font-black uppercase text-slate-400 tracking-widest">{{ t('fail2ban.jail') }}</th>
                <th class="px-8 py-4 text-[10px] font-black uppercase text-slate-400 tracking-widest text-right">{{ t('common.actions') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-800">
              <tr v-for="item in bannedIPs" :key="item.ip + item.jail" class="group transition-colors hover:bg-slate-50/50 dark:hover:bg-slate-800/30">
                <td class="px-8 py-5 text-sm font-black text-slate-900 dark:text-white tracking-tight">{{ item.ip }}</td>
                <td class="px-8 py-5">
                  <span class="px-2 py-1 bg-amber-100 dark:bg-amber-500/10 text-amber-600 dark:text-amber-400 text-[10px] font-black uppercase rounded-lg border border-amber-200 dark:border-amber-500/20">
                    {{ item.jail }}
                  </span>
                </td>
                <td class="px-8 py-5 text-right">
                  <button 
                    @click="unbanIP(item.ip, item.jail)" 
                    :disabled="processingUnban === item.ip"
                    class="px-5 py-2.5 bg-red-500 hover:bg-red-600 text-white rounded-xl text-[10px] font-black uppercase tracking-widest transition-all shadow-lg shadow-red-500/20 disabled:opacity-50"
                  >
                    {{ processingUnban === item.ip ? '...' : t('fail2ban.unban') }}
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
          <div v-else class="py-24 text-center">
            <div class="w-20 h-20 bg-emerald-100 dark:bg-emerald-500/10 text-emerald-500 rounded-3xl flex items-center justify-center mx-auto mb-6 shadow-xl shadow-emerald-500/10 rotate-3 group-hover:rotate-0 transition-transform">
              <svg class="w-10 h-10" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04M12 21.056c1.54 0 3.024-.317 4.382-.886L12.038 12 7.634 20.17A11.947 11.947 0 0112 21.056z" /></svg>
            </div>
            <p class="text-xl font-black text-slate-900 dark:text-white tracking-tight uppercase">{{ t('fail2ban.empty') }}</p>
            <p class="text-sm text-slate-500 font-medium mt-2">Сервер в безопасности, активных угроз не обнаружено</p>
          </div>
        </div>

        <div class="p-8 border-t border-slate-100 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/10 flex justify-end">
          <button @click="showF2BModal = false" class="px-8 py-3 bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl font-bold text-xs uppercase tracking-widest hover:bg-slate-50 dark:hover:bg-slate-800 transition-all text-slate-900 dark:text-white shadow-sm">
            {{ t('common.close') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.health-card {
  background-color: var(--card-bg, white);
  border: 1px solid rgba(0,0,0,0.05);
  border-radius: 2rem;
  padding: 1.5rem;
  transition: all 0.3s ease;
}

:border-dark .health-card {
  background-color: #0f172a;
  border-color: rgba(255,255,255,0.05);
}

.health-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.1);
}

/* Fallback if Tailwind context is lost */
.dark .health-card {
  background-color: #0f172a;
  border-color: #1e293b;
}
</style>
