<script setup>
import { ref, onMounted, onUnmounted, computed, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import api from '@/api/axios'

const { t } = useI18n()
const loading = ref(true)
const logs = ref('')
const filter = ref('')
const lines = ref(200)
const autoRefresh = ref(true)
const logContainer = ref(null)
let timer = null

const fetchLogs = async () => {
  try {
    const response = await api.get(`/system/logs?lines=${lines.value}`)
    logs.value = response.data.logs
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

const filteredLogs = computed(() => {
  if (!filter.value) return logs.value
  const search = filter.value.toLowerCase()
  return logs.value.split('\n')
    .filter(line => line.toLowerCase().includes(search))
    .join('\n')
})

const startTimer = () => {
  if (timer) clearInterval(timer)
  timer = setInterval(() => {
    if (autoRefresh.value) fetchLogs()
  }, 10000)
}

onMounted(() => {
  fetchLogs()
  startTimer()
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<template>
  <div class="space-y-6 h-[calc(100vh-140px)] flex flex-col animate-in fade-in duration-500">
    <header class="flex justify-between items-end shrink-0">
      <div>
        <h1 class="text-4xl font-black text-slate-900 dark:text-white tracking-tight">
          {{ t('menu.server_logs') }}
        </h1>
        <p class="text-slate-500 dark:text-slate-400 mt-2 font-medium italic">
          /var/log/mail.log
        </p>
      </div>
      <div class="flex gap-4 items-center">
        <div class="flex items-center gap-2 px-4 py-2 bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl">
          <input type="checkbox" v-model="autoRefresh" class="rounded border-slate-300 text-mail-blue-600 focus:ring-mail-blue-500">
          <span class="text-xs font-bold text-slate-500 uppercase tracking-widest">{{ t('common.auto_refresh') || 'Auto-refresh' }}</span>
        </div>
        <select v-model="lines" @change="fetchLogs" class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl px-4 py-2 text-xs font-bold text-slate-500 outline-none focus:border-mail-blue-500 transition-colors">
          <option :value="100">100 строк</option>
          <option :value="500">500 строк</option>
          <option :value="1000">1000 строк</option>
        </select>
        <button @click="fetchLogs" class="p-3 bg-mail-blue-600 text-white rounded-xl hover:bg-mail-blue-700 transition-all shadow-lg shadow-mail-blue-500/20">
          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" /></svg>
        </button>
      </div>
    </header>

    <div class="flex-1 flex flex-col glass-panel overflow-hidden border-2 border-slate-200 dark:border-slate-800 shadow-2xl shadow-black/5">
      <!-- Search/Filter Bar -->
      <div class="p-4 bg-slate-50 dark:bg-slate-900/50 border-b border-slate-200 dark:border-slate-800 flex items-center gap-4">
        <svg class="w-4 h-4 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" /></svg>
        <input 
          v-model="filter" 
          type="text" 
          placeholder="Фильтр по отправителю, ID или тексту ошибки..." 
          class="bg-transparent border-none outline-none text-sm font-medium w-full text-slate-700 dark:text-slate-300"
        >
      </div>

      <!-- Terminal window -->
      <div 
        ref="logContainer"
        class="flex-1 bg-slate-950 p-6 font-mono text-[13px] leading-relaxed overflow-auto scrollbar-thin scrollbar-thumb-slate-800"
      >
        <div v-if="loading" class="text-slate-500 animate-pulse">Загрузка системных логов...</div>
        <div v-else-if="!logs" class="text-red-400 font-bold">ОШИБКА: Логи недоступны. Проверьте права доступа к /var/log/mail.log</div>
        <pre v-else class="text-slate-300 whitespace-pre-wrap"><code v-html="highlightLogs(filteredLogs)"></code></pre>
      </div>
    </div>
  </div>
</template>

<script>
// Добавляем хелпер для подсветки
export default {
  methods: {
    highlightLogs(text) {
      if (!text) return ''
      // Подсвечиваем ключевые слова
      return text
        .replace(/status=sent/g, '<span class="text-emerald-400 font-bold">status=sent</span>')
        .replace(/status=deferred/g, '<span class="text-amber-400 font-bold">status=deferred</span>')
        .replace(/status=bounced/g, '<span class="text-red-400 font-bold">status=bounced</span>')
        .replace(/CONNECT/g, '<span class="text-sky-400 font-bold">CONNECT</span>')
        .replace(/DISCONNECT/g, '<span class="text-slate-500 font-bold">DISCONNECT</span>')
        .replace(/SASL [A-Z]+ authentication/g, '<span class="text-purple-400 font-bold">$&</span>')
        .replace(/client=[^, ]+/g, '<span class="text-mail-blue-400">$&</span>')
    }
  }
}
</script>

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
}
.scrollbar-thin::-webkit-scrollbar-track {
  background-color: #0f172a; /* b-slate-900 */
}
.scrollbar-thin::-webkit-scrollbar-thumb {
  background-color: #1e293b; /* bg-slate-800 */
  border-radius: 9999px;
}
.scrollbar-thin::-webkit-scrollbar-thumb:hover {
  background-color: #334155; /* bg-slate-700 */
}
</style>
