<script setup>
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import api from '@/api/axios'

const { t } = useI18n()
const loading = ref(true)
const activeTab = ref('raw') // 'raw' или 'analysis'
const logs = ref([])
const search = ref('')
const lines = ref(200)
const autoRefresh = ref(true)
const logContainer = ref(null)
let timer = null

// Данные для анализа
const analysisData = ref({
  total_transactions: 0,
  sent_count: 0,
  deferred_count: 0,
  bounced_count: 0,
  reject_count: 0,
  transactions: [],
  rejects: [],
  top_senders: [],
  top_recipients: [],
  top_clients: []
})

const quickFilters = [
  { label: 'Rejects', value: 'reject', color: 'bg-red-500/10 text-red-500 border-red-500/20' },
  { label: 'Errors', value: 'error', color: 'bg-orange-500/10 text-orange-500 border-orange-500/20' },
  { label: 'Sent', value: 'status=sent', color: 'bg-emerald-500/10 text-emerald-500 border-emerald-500/20' },
  { label: 'Auth', value: 'sasl', color: 'bg-purple-500/10 text-purple-500 border-purple-500/20' }
]

const fetchLogs = async () => {
  if (activeTab.value === 'raw') {
    try {
      const response = await api.get('/system/logs', {
        params: {
          lines: lines.value,
          search: search.value?.trim() || ''
        }
      })
      const rawLogs = response.data.logs || ''
      logs.value = rawLogs.split('\n').filter(l => l.trim() !== '')
      
      if (autoRefresh.value) {
        await nextTick()
        scrollToBottom()
      }
    } catch (error) {
      console.error('Failed to fetch logs:', error)
    } finally {
      loading.value = false
    }
  } else {
    try {
      loading.value = true
      const response = await api.get('/system/logs/analysis', {
        params: {
          lines: lines.value
        }
      })
      const data = response.data || {}
      analysisData.value = {
        total_transactions: data.total_transactions || 0,
        sent_count: data.sent_count || 0,
        deferred_count: data.deferred_count || 0,
        bounced_count: data.bounced_count || 0,
        reject_count: data.reject_count || 0,
        transactions: data.transactions || [],
        rejects: data.rejects || [],
        top_senders: data.top_senders || [],
        top_recipients: data.top_recipients || [],
        top_clients: data.top_clients || []
      }
    } catch (error) {
      console.error('Failed to fetch log analysis:', error)
    } finally {
      loading.value = false
    }
  }
}

const scrollToBottom = () => {
  if (logContainer.value) {
    logContainer.value.scrollTop = logContainer.value.scrollHeight
  }
}

const toggleFilter = (val) => {
  if (search.value === val) {
    search.value = ''
  } else {
    search.value = val
  }
  fetchLogs()
}

// Следим за поиском, строками и вкладками
watch([lines, activeTab], () => fetchLogs())
let searchDebounce = null
watch(search, () => {
  if (searchDebounce) clearTimeout(searchDebounce)
  searchDebounce = setTimeout(fetchLogs, 500)
})

const startTimer = () => {
  if (timer) clearInterval(timer)
  timer = setInterval(() => {
    if (autoRefresh.value && !search.value && activeTab.value === 'raw') fetchLogs()
  }, 5000)
}

onMounted(() => {
  fetchLogs()
  startTimer()
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})

const getLineStyle = (line) => {
  const l = line.toLowerCase()
  if (l.includes('reject') || l.includes('error') || l.includes('fatal') || l.includes('failed') || l.includes('bounced')) {
    return 'bg-red-500/5 text-red-200 border-l-2 border-red-500'
  }
  if (l.includes('warning') || l.includes('deferred') || l.includes('timeout')) {
    return 'bg-amber-500/5 text-amber-200 border-l-2 border-amber-500'
  }
  if (l.includes('status=sent') || l.includes('passed') || l.includes('authenticated')) {
    return 'text-emerald-300'
  }
  return 'text-slate-400'
}

const highlightMatch = (line) => {
  const q = search.value?.trim()
  if (!q) return line
  const regex = new RegExp(`(${q})`, 'gi')
  return line.replace(regex, '<mark class="bg-mail-blue-500/40 text-white rounded px-0.5">$1</mark>')
}

// Управление разворачиванием транзакций
const expandedTx = ref({})
const toggleTx = (id) => {
  expandedTx.value[id] = !expandedTx.value[id]
}

const showSpam = ref(false)
const txSearch = ref('')
const filteredTransactions = computed(() => {
  let list = analysisData.value.transactions
  if (!showSpam.value) {
    list = list.filter(tx => tx.from && tx.from !== 'unknown' && tx.size > 0)
  }
  
  const q = txSearch.value.toLowerCase().trim()
  if (q) {
    list = list.filter(tx => 
      (tx.queue_id && tx.queue_id.toLowerCase().includes(q)) ||
      (tx.from && tx.from.toLowerCase().includes(q)) ||
      (tx.client_host && tx.client_host.toLowerCase().includes(q)) ||
      (tx.client_ip && tx.client_ip.toLowerCase().includes(q)) ||
      (tx.message_id && tx.message_id.toLowerCase().includes(q)) ||
      (tx.deliveries && tx.deliveries.some(d => 
        (d.to && d.to.toLowerCase().includes(q)) ||
        (d.status_msg && d.status_msg.toLowerCase().includes(q)) ||
        (d.relay_host && d.relay_host.toLowerCase().includes(q)) ||
        (d.dsn && d.dsn.toLowerCase().includes(q))
      ))
    )
  }
  return list
})
</script>

<template>
  <div class="space-y-6 h-[calc(100vh-140px)] flex flex-col animate-in fade-in duration-500">
    <header class="flex flex-col md:flex-row justify-between items-start md:items-end gap-4 shrink-0">
      <div>
        <div class="flex items-center gap-3">
          <h1 class="text-4xl font-black text-slate-900 dark:text-white tracking-tight">
            {{ t('menu.server_logs') }}
          </h1>
          <div class="flex items-center gap-1.5 px-2 py-1 bg-emerald-500/10 text-emerald-500 rounded-lg border border-emerald-500/20">
            <div class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></div>
            <span class="text-[10px] font-black uppercase tracking-widest">Live</span>
          </div>
        </div>
        <p class="text-slate-500 dark:text-slate-400 mt-2 font-medium italic flex items-center gap-2">
           <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>
           {{ t('server_logs.subtitle') }}
        </p>
      </div>

      <div class="flex flex-wrap gap-3 items-center">
        <!-- Вкладки управления режимом логов -->
        <div class="flex bg-slate-100 dark:bg-slate-900 p-1 rounded-xl border border-slate-200 dark:border-slate-800">
          <button 
            @click="activeTab = 'raw'"
            class="px-4 py-1.5 rounded-lg text-xs font-black uppercase tracking-widest transition-all"
            :class="activeTab === 'raw' ? 'bg-white dark:bg-slate-800 text-mail-blue-600 dark:text-white shadow' : 'text-slate-500 hover:text-slate-800 dark:hover:text-slate-300'"
          >
            {{ t('server_logs.tab_raw') }}
          </button>
          <button 
            @click="activeTab = 'analysis'"
            class="px-4 py-1.5 rounded-lg text-xs font-black uppercase tracking-widest transition-all"
            :class="activeTab === 'analysis' ? 'bg-white dark:bg-slate-800 text-mail-blue-600 dark:text-white shadow' : 'text-slate-500 hover:text-slate-800 dark:hover:text-slate-300'"
          >
            {{ t('server_logs.tab_analysis') }}
          </button>
        </div>

        <!-- Quick Filters (только для сырых логов) -->
        <div v-if="activeTab === 'raw'" class="flex gap-2">
          <button 
            v-for="f in quickFilters" 
            :key="f.label"
            @click="toggleFilter(f.value)"
            class="px-3 py-1.5 rounded-xl border text-[10px] font-black uppercase tracking-widest transition-all"
            :class="[search === f.value ? 'ring-2 ring-mail-blue-500 scale-105' : 'opacity-70 hover:opacity-100', f.color]"
          >
            {{ f.label }}
          </button>
        </div>

        <div class="hidden lg:block w-px h-8 bg-slate-200 dark:bg-slate-800 mx-2"></div>

        <div v-if="activeTab === 'raw'" class="flex items-center gap-2 px-4 py-2 bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl">
          <input type="checkbox" v-model="autoRefresh" class="rounded border-slate-300 text-mail-blue-600 focus:ring-mail-blue-500">
          <span class="text-[10px] font-bold text-slate-500 uppercase tracking-widest">{{ t('common.auto_refresh') || 'Auto-refresh' }}</span>
        </div>

        <select v-model="lines" class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl px-4 py-2 text-[10px] font-black uppercase tracking-widest text-slate-500 outline-none focus:border-mail-blue-500 transition-colors">
          <option :value="100">{{ t('server_logs.rows', { count: 100 }) }}</option>
          <option :value="200" v-if="activeTab === 'raw'">{{ t('server_logs.rows', { count: 200 }) }}</option>
          <option :value="500">{{ t('server_logs.rows', { count: 500 }) }}</option>
          <option :value="1000">{{ t('server_logs.rows', { count: 1000 }) }}</option>
          <option :value="5000">{{ t('server_logs.rows', { count: 5000 }) }}</option>
          <option :value="10000">{{ t('server_logs.rows', { count: 10000 }) }}</option>
        </select>

        <button @click="fetchLogs" class="p-2.5 bg-mail-blue-600 text-white rounded-xl hover:bg-mail-blue-700 transition-all shadow-lg shadow-mail-blue-500/20 active:scale-95">
          <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" /></svg>
        </button>
      </div>
    </header>

    <!-- ВКЛАДКА: СЫРОЙ ЛОГ -->
    <div v-if="activeTab === 'raw'" class="flex-1 flex flex-col glass-panel overflow-hidden border-2 border-slate-200 dark:border-slate-800 shadow-2xl">
      <!-- Search Bar -->
      <div class="p-4 bg-slate-50 dark:bg-slate-900/50 border-b border-slate-200 dark:border-slate-800 flex items-center gap-4 group">
        <svg class="w-5 h-5 text-slate-400 group-focus-within:text-mail-blue-500 transition-colors" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" /></svg>
        <input 
          v-model="search" 
          type="text" 
          :placeholder="t('server_logs.placeholder')" 
          class="bg-transparent border-none outline-none text-sm font-bold w-full text-slate-700 dark:text-slate-200 placeholder-slate-400"
        >
        <button v-if="search" @click="search = ''" class="text-slate-400 hover:text-slate-600 dark:hover:text-white">
          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M6 18L18 6M6 6l12 12" /></svg>
        </button>
      </div>

      <!-- Terminal window -->
      <div 
        ref="logContainer"
        class="flex-1 bg-slate-950 font-mono text-[11px] md:text-[13px] leading-relaxed overflow-auto scrollbar-thin scrollbar-thumb-slate-800 p-2"
      >
        <div v-if="loading && logs.length === 0" class="flex items-center justify-center h-full">
           <div class="flex flex-col items-center gap-4">
             <div class="w-10 h-10 border-4 border-mail-blue-500/20 border-t-mail-blue-500 rounded-full animate-spin"></div>
             <p class="text-slate-500 font-bold uppercase tracking-widest text-[10px] animate-pulse">{{ t('server_logs.initializing') }}</p>
           </div>
        </div>
        
        <div v-else-if="logs.length === 0" class="flex items-center justify-center h-full opacity-50">
           <p class="text-slate-500 italic">{{ t('server_logs.empty') }}</p>
        </div>

        <div v-else class="space-y-px">
          <div 
            v-for="(line, idx) in logs" 
            :key="idx"
            class="px-4 py-0.5 hover:bg-slate-800/40 transition-colors whitespace-nowrap"
            :class="getLineStyle(line)"
          >
            <span class="opacity-20 mr-4 select-none inline-block w-8 text-right">{{ idx + 1 }}</span>
            <span v-html="highlightMatch(line)"></span>
          </div>
        </div>
      </div>
    </div>

    <!-- ВКЛАДКА: АНАЛИЗ ЛОГОВ -->
    <div v-else class="flex-1 overflow-y-auto pr-2 space-y-4">
      <div v-if="loading" class="flex flex-col items-center justify-center h-64">
        <div class="w-10 h-10 border-4 border-mail-blue-500/20 border-t-mail-blue-500 rounded-full animate-spin"></div>
        <p class="text-slate-500 font-bold uppercase tracking-widest text-[10px] mt-4 animate-pulse">{{ t('common.loading') }}</p>
      </div>

      <div v-else class="space-y-4">
        <!-- Карточки статистики -->
        <div class="grid grid-cols-2 md:grid-cols-5 gap-4">
          <div class="glass-panel p-4 flex flex-col justify-between border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-900/50">
            <span class="text-[10px] font-black uppercase tracking-wider text-slate-400">{{ t('server_logs.stat_total_tx') }}</span>
            <span class="text-3xl font-black mt-2 text-slate-800 dark:text-white">{{ analysisData.total_transactions }}</span>
          </div>
          <div class="glass-panel p-4 flex flex-col justify-between border border-slate-200 dark:border-slate-800 bg-emerald-500/5 dark:bg-emerald-500/10">
            <span class="text-[10px] font-black uppercase tracking-wider text-emerald-500">{{ t('server_logs.stat_sent') }}</span>
            <span class="text-3xl font-black mt-2 text-emerald-600 dark:text-emerald-400">{{ analysisData.sent_count }}</span>
          </div>
          <div class="glass-panel p-4 flex flex-col justify-between border border-slate-200 dark:border-slate-800 bg-amber-500/5 dark:bg-amber-500/10">
            <span class="text-[10px] font-black uppercase tracking-wider text-amber-500">{{ t('server_logs.stat_deferred') }}</span>
            <span class="text-3xl font-black mt-2 text-amber-600 dark:text-amber-400">{{ analysisData.deferred_count }}</span>
          </div>
          <div class="glass-panel p-4 flex flex-col justify-between border border-slate-200 dark:border-slate-800 bg-red-500/5 dark:bg-red-500/10">
            <span class="text-[10px] font-black uppercase tracking-wider text-red-500">{{ t('server_logs.stat_bounced') }}</span>
            <span class="text-3xl font-black mt-2 text-red-600 dark:text-red-400">{{ analysisData.bounced_count }}</span>
          </div>
          <div class="glass-panel p-4 flex flex-col justify-between border border-slate-200 dark:border-slate-800 bg-purple-500/5 dark:bg-purple-500/10 col-span-2 md:col-span-1">
            <span class="text-[10px] font-black uppercase tracking-wider text-purple-500">{{ t('server_logs.stat_rejected') }}</span>
            <span class="text-3xl font-black mt-2 text-purple-600 dark:text-purple-400">{{ analysisData.reject_count }}</span>
          </div>
        </div>

        <!-- Списки ТОПов -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
          <!-- Топ Отправителей -->
          <div class="glass-panel p-5 border border-slate-200 dark:border-slate-800">
            <h3 class="text-sm font-black uppercase tracking-wider text-slate-800 dark:text-slate-200 mb-4">{{ t('server_logs.top_senders') }}</h3>
            <div v-if="analysisData.top_senders.length === 0" class="text-xs text-slate-400 italic">{{ t('common.none') }}</div>
            <ul v-else class="space-y-3">
              <li v-for="item in analysisData.top_senders" :key="item.key" class="flex justify-between items-center text-xs">
                <span class="font-semibold text-slate-600 dark:text-slate-300 truncate max-w-[200px]" :title="item.key">{{ item.key }}</span>
                <span class="px-2 py-0.5 bg-slate-100 dark:bg-slate-800 rounded font-black text-slate-700 dark:text-slate-400">{{ item.value }}</span>
              </li>
            </ul>
          </div>

          <!-- Топ Получателей -->
          <div class="glass-panel p-5 border border-slate-200 dark:border-slate-800">
            <h3 class="text-sm font-black uppercase tracking-wider text-slate-800 dark:text-slate-200 mb-4">{{ t('server_logs.top_recipients') }}</h3>
            <div v-if="analysisData.top_recipients.length === 0" class="text-xs text-slate-400 italic">{{ t('common.none') }}</div>
            <ul v-else class="space-y-3">
              <li v-for="item in analysisData.top_recipients" :key="item.key" class="flex justify-between items-center text-xs">
                <span class="font-semibold text-slate-600 dark:text-slate-300 truncate max-w-[200px]" :title="item.key">{{ item.key }}</span>
                <span class="px-2 py-0.5 bg-slate-100 dark:bg-slate-800 rounded font-black text-slate-700 dark:text-slate-400">{{ item.value }}</span>
              </li>
            </ul>
          </div>

          <!-- Топ Клиентов -->
          <div class="glass-panel p-5 border border-slate-200 dark:border-slate-800">
            <h3 class="text-sm font-black uppercase tracking-wider text-slate-800 dark:text-slate-200 mb-4">{{ t('server_logs.top_clients') }}</h3>
            <div v-if="analysisData.top_clients.length === 0" class="text-xs text-slate-400 italic">{{ t('common.none') }}</div>
            <ul v-else class="space-y-3">
              <li v-for="item in analysisData.top_clients" :key="item.key" class="flex justify-between items-center text-xs">
                <span class="font-mono font-semibold text-slate-600 dark:text-slate-300">{{ item.key }}</span>
                <span class="px-2 py-0.5 bg-slate-100 dark:bg-slate-800 rounded font-black text-slate-700 dark:text-slate-400">{{ item.value }}</span>
              </li>
            </ul>
          </div>
        </div>

        <!-- Таблица NOQUEUE отказов -->
        <div class="glass-panel border border-slate-200 dark:border-slate-800 overflow-hidden">
          <div class="px-5 py-4 border-b border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-900/50">
            <h3 class="text-sm font-black uppercase tracking-wider text-slate-800 dark:text-slate-200">{{ t('server_logs.table_rejects') }}</h3>
          </div>
          <div class="overflow-x-auto">
            <table class="w-full text-left border-collapse">
              <thead>
                <tr class="border-b border-slate-200 dark:border-slate-800 text-[10px] font-black uppercase tracking-widest text-slate-400 bg-slate-50/30 dark:bg-slate-900/20">
                  <th class="px-5 py-3">{{ t('server_logs.col_time') }}</th>
                  <th class="px-5 py-3">{{ t('server_logs.col_client') }}</th>
                  <th class="px-5 py-3">{{ t('server_logs.col_from') }}</th>
                  <th class="px-5 py-3">{{ t('server_logs.col_to') }}</th>
                  <th class="px-5 py-3">{{ t('server_logs.col_reason') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-if="analysisData.rejects.length === 0">
                  <td colspan="5" class="px-5 py-6 text-center text-xs text-slate-400 italic">{{ t('common.none') }}</td>
                </tr>
                <tr 
                  v-for="(rej, idx) in analysisData.rejects" 
                  :key="idx" 
                  class="border-b border-slate-100 dark:border-slate-800/50 hover:bg-slate-50/50 dark:hover:bg-slate-900/10 transition-colors text-xs"
                >
                  <td class="px-5 py-3.5 font-mono text-slate-400 whitespace-nowrap">{{ rej.timestamp }}</td>
                  <td class="px-5 py-3.5 font-semibold text-slate-700 dark:text-slate-300 whitespace-nowrap">{{ rej.client }}</td>
                  <td class="px-5 py-3.5 font-semibold text-slate-600 dark:text-slate-400 truncate max-w-[150px]">{{ rej.from }}</td>
                  <td class="px-5 py-3.5 font-semibold text-slate-600 dark:text-slate-400 truncate max-w-[150px]">{{ rej.to }}</td>
                  <td class="px-5 py-3.5 text-red-500 font-medium font-mono select-all">{{ rej.reason }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- Список транзакций очереди / истории писем -->
        <div class="glass-panel border border-slate-200 dark:border-slate-800 overflow-hidden">
          <div class="px-5 py-4 border-b border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-900/50 flex flex-wrap justify-between items-center gap-4">
            <h3 class="text-sm font-black uppercase tracking-wider text-slate-800 dark:text-slate-200">{{ t('server_logs.queue_title') }}</h3>
            
            <div class="flex items-center gap-4">
              <div class="relative">
                <svg class="w-4 h-4 text-slate-400 absolute left-2.5 top-1/2 -translate-y-1/2" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" /></svg>
                <input 
                  type="text" 
                  v-model="txSearch" 
                  :placeholder="t('common.search')" 
                  class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-lg pl-9 pr-3 py-1.5 text-xs text-slate-700 dark:text-slate-300 outline-none focus:border-mail-blue-500 w-48 transition-colors"
                >
              </div>
              <label class="flex items-center gap-2 cursor-pointer">
                <input type="checkbox" v-model="showSpam" class="rounded border-slate-300 text-mail-blue-600 focus:ring-mail-blue-500">
                <span class="text-[10px] font-bold text-slate-500 uppercase tracking-widest">{{ t('server_logs.show_spam') }}</span>
              </label>
            </div>
          </div>
          <div class="divide-y divide-slate-100 dark:divide-slate-800">
            <div v-if="filteredTransactions.length === 0" class="px-5 py-8 text-center text-xs text-slate-400 italic">
              {{ t('server_logs.no_transactions') }}
            </div>
            <div 
              v-for="tx in filteredTransactions" 
              :key="tx.queue_id"
              class="p-4 hover:bg-slate-50/30 dark:hover:bg-slate-900/10 transition-colors"
              :class="{'opacity-60 grayscale': !tx.from || tx.from === 'unknown' || tx.size === 0}"
            >
              <div @click="toggleTx(tx.queue_id)" class="flex flex-wrap items-center justify-between gap-4 cursor-pointer">
                <div class="flex items-center gap-3">
                  <span class="font-mono text-xs font-black bg-slate-100 dark:bg-slate-800 px-2.5 py-1 rounded text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-slate-700">{{ tx.queue_id }}</span>
                  <div class="flex flex-col">
                    <span class="text-xs font-bold text-slate-700 dark:text-slate-300">
                      {{ tx.from || 'unknown' }} &rarr; 
                      <span class="text-slate-500 font-medium">
                        {{ (tx.deliveries || []).length > 0 ? tx.deliveries.map(d => d.to).join(', ') : '?' }}
                      </span>
                    </span>
                    <span class="text-[10px] text-slate-400 font-mono mt-0.5">{{ tx.timestamp }} | Размер: {{ tx.size }} B</span>
                  </div>
                </div>

                <div class="flex items-center gap-3">
                  <!-- Суммарный бейдж статуса -->
                  <span 
                    v-if="(tx.deliveries || []).length > 0"
                    class="px-2 py-0.5 rounded text-[10px] font-black uppercase tracking-widest"
                    :class="{
                      'bg-emerald-500/10 text-emerald-500 border border-emerald-500/20': tx.deliveries.every(d => d.status === 'sent'),
                      'bg-red-500/10 text-red-500 border border-red-500/20': tx.deliveries.some(d => d.status === 'bounced'),
                      'bg-amber-500/10 text-amber-500 border border-amber-500/20': tx.deliveries.some(d => d.status === 'deferred') && !tx.deliveries.some(d => d.status === 'bounced'),
                    }"
                  >
                    {{ tx.deliveries.some(d => d.status === 'bounced') ? 'Bounced' : tx.deliveries.some(d => d.status === 'deferred') ? 'Deferred' : 'Sent' }}
                  </span>
                  
                  <!-- Стрелочка развертывания -->
                  <svg 
                    class="w-5 h-5 text-slate-400 transition-transform duration-300"
                    :class="{ 'rotate-180': expandedTx[tx.queue_id] }"
                    fill="none" 
                    viewBox="0 0 24 24" 
                    stroke="currentColor"
                  >
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M19 9l-7 7-7-7" />
                  </svg>
                </div>
              </div>

              <!-- Детали транзакции (выпадающий блок) -->
              <div v-if="expandedTx[tx.queue_id]" class="mt-4 pt-4 border-t border-slate-100 dark:border-slate-800/80 space-y-3 animate-in slide-in-from-top-1 duration-200">
                <div class="grid grid-cols-1 md:grid-cols-2 gap-4 text-xs">
                  <div>
                    <span class="text-slate-400 font-bold uppercase text-[9px] tracking-wider">{{ t('server_logs.sender_client') }}</span>
                    <p class="font-semibold text-slate-700 dark:text-slate-300 mt-0.5">
                      {{ tx.client_host }} <span class="text-slate-400 font-mono">[{{ tx.client_ip }}]</span>
                    </p>
                  </div>
                  <div>
                    <span class="text-slate-400 font-bold uppercase text-[9px] tracking-wider">Message-ID</span>
                    <p class="font-mono text-slate-600 dark:text-slate-400 mt-0.5 truncate select-all" :title="tx.message_id">{{ tx.message_id || 'N/A' }}</p>
                  </div>
                </div>

                <!-- Попытки доставки -->
                <div class="space-y-1.5 mt-3">
                  <span class="text-slate-400 font-bold uppercase text-[9px] tracking-wider block">{{ t('server_logs.deliveries_title') }}</span>
                  <div 
                    v-for="(del, dIdx) in (tx.deliveries || [])" 
                    :key="dIdx"
                    class="bg-slate-50 dark:bg-slate-900/50 p-3 rounded-lg border border-slate-200/60 dark:border-slate-800/60 space-y-1.5 text-xs"
                  >
                    <div class="flex justify-between items-center">
                      <span class="font-bold text-slate-700 dark:text-slate-300">{{ del.to }}</span>
                      <span 
                        class="px-1.5 py-0.5 rounded text-[9px] font-black tracking-wider"
                        :class="{
                          'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20': del.status === 'sent',
                          'bg-red-500/10 text-red-600 dark:text-red-400 border border-red-500/20': del.status === 'bounced',
                          'bg-amber-500/10 text-amber-600 dark:text-amber-400 border border-amber-500/20': del.status === 'deferred',
                        }"
                      >
                        {{ del.status === 'sent' ? t('server_logs.delivery_status_sent') : del.status === 'bounced' ? t('server_logs.delivery_status_bounced') : del.status === 'deferred' ? t('server_logs.delivery_status_deferred') : del.status }}
                      </span>
                    </div>
                    <div class="grid grid-cols-1 md:grid-cols-2 gap-2 text-[11px] text-slate-500">
                      <div>
                        <span class="font-semibold">{{ t('server_logs.relay') }}:</span> {{ del.relay_host }} <span class="font-mono text-[10px] bg-slate-200 dark:bg-slate-800 px-1 rounded">[{{ del.relay_ip }}]</span>
                      </div>
                      <div>
                        <span class="font-semibold">{{ t('server_logs.dsn') }}:</span> <span class="font-mono font-bold">{{ del.dsn }}</span>
                      </div>
                    </div>
                    <div class="mt-1">
                      <span class="text-[9px] font-bold text-slate-400 uppercase tracking-widest">{{ t('server_logs.remote_response') }}</span>
                      <p class="font-mono text-[11px] text-slate-600 dark:text-slate-400 bg-white/50 dark:bg-slate-950 p-2 rounded border border-slate-200 dark:border-slate-800 select-all leading-normal mt-1">{{ del.status_msg }}</p>
                    </div>
                  </div>
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
pre {
  margin: 0;
}
code {
  font-family: 'JetBrains Mono', 'Fira Code', 'Courier New', monospace;
}
/* Кастомный скроллбар для терминала */
.scrollbar-thin::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}
.scrollbar-thin::-webkit-scrollbar-track {
  background-color: #020617; /* slate-950 */
}
.scrollbar-thin::-webkit-scrollbar-thumb {
  background-color: #1e293b; /* bg-slate-800 */
  border-radius: 9999px;
  border: 2px solid #020617;
}
.scrollbar-thin::-webkit-scrollbar-thumb:hover {
  background-color: #334155; /* bg-slate-700 */
}

mark {
  background-color: rgba(59, 130, 246, 0.4);
  color: white;
  border-radius: 2px;
}
</style>
