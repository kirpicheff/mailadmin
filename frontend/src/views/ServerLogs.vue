<script setup>
import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import api from '@/api/axios'

const { t } = useI18n()
const loading = ref(true)
const logs = ref([])
const search = ref('')
const lines = ref(200)
const autoRefresh = ref(true)
const logContainer = ref(null)
let timer = null

const quickFilters = [
  { label: 'Rejects', value: 'reject', color: 'bg-red-500/10 text-red-500 border-red-500/20' },
  { label: 'Errors', value: 'error', color: 'bg-orange-500/10 text-orange-500 border-orange-500/20' },
  { label: 'Sent', value: 'status=sent', color: 'bg-emerald-500/10 text-emerald-500 border-emerald-500/20' },
  { label: 'Auth', value: 'sasl', color: 'bg-purple-500/10 text-purple-500 border-purple-500/20' }
]

const fetchLogs = async () => {
  try {
    const response = await api.get('/system/logs', {
      params: {
        lines: lines.value,
        search: search.value
      }
    })
    
    // Разбиваем строку на массив для построчного рендеринга
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

// Следим за поиском и строками
watch([lines], () => fetchLogs())
let searchDebounce = null
watch(search, () => {
  if (searchDebounce) clearTimeout(searchDebounce)
  searchDebounce = setTimeout(fetchLogs, 500)
})

const startTimer = () => {
  if (timer) clearInterval(timer)
  timer = setInterval(() => {
    if (autoRefresh.value && !search.value) fetchLogs()
  }, 5000) // 5 секунд для активного мониторинга
}

onMounted(() => {
  fetchLogs()
  startTimer()
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})

// Функция определения стиля строки
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
  if (!search.value) return line
  const regex = new RegExp(`(${search.value})`, 'gi')
  return line.replace(regex, '<mark class="bg-mail-blue-500/40 text-white rounded px-0.5">$1</mark>')
}
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
        <!-- Quick Filters -->
        <div class="flex gap-2">
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

        <div class="flex items-center gap-2 px-4 py-2 bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl">
          <input type="checkbox" v-model="autoRefresh" class="rounded border-slate-300 text-mail-blue-600 focus:ring-mail-blue-500">
          <span class="text-[10px] font-bold text-slate-500 uppercase tracking-widest">{{ t('common.auto_refresh') || 'Auto-refresh' }}</span>
        </div>

        <select v-model="lines" class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl px-4 py-2 text-[10px] font-black uppercase tracking-widest text-slate-500 outline-none focus:border-mail-blue-500 transition-colors">
          <option :value="100">{{ t('server_logs.rows', { count: 100 }) }}</option>
          <option :value="500">{{ t('server_logs.rows', { count: 500 }) }}</option>
          <option :value="1000">{{ t('server_logs.rows', { count: 1000 }) }}</option>
          <option :value="5000">{{ t('server_logs.rows', { count: 5000 }) }}</option>
        </select>

        <button @click="fetchLogs" class="p-2.5 bg-mail-blue-600 text-white rounded-xl hover:bg-mail-blue-700 transition-all shadow-lg shadow-mail-blue-500/20 active:scale-95">
          <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" /></svg>
        </button>
      </div>
    </header>

    <div class="flex-1 flex flex-col glass-panel overflow-hidden border-2 border-slate-200 dark:border-slate-800 shadow-2xl">
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
