<script setup>
import { ref, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import api from '@/api/axios'

const { t } = useI18n()
const loading = ref(true)
const queue = ref([])
const refreshing = ref(false)

const fetchQueue = async () => {
  refreshing.ref = true
  try {
    const response = await api.get('/system/queue')
    queue.value = response.data
  } catch (error) {
    console.error('Failed to fetch queue:', error)
  } finally {
    loading.value = false
    refreshing.value = false
  }
}

const flushQueue = async () => {
  if (!confirm('Вы уверены, что хотите принудительно отправить всю очередь?')) return
  try {
    await api.post('/system/queue/flush')
    alert('Сигнал отправки подан (flush)')
    fetchQueue()
  } catch (error) {
    alert('Ошибка при выполнении flush')
  }
}

const deleteItem = async (id) => {
  if (!confirm(`Удалить письмо ${id} из очереди?`)) return
  try {
    await api.delete(`/system/queue/${id}`)
    fetchQueue()
  } catch (error) {
    alert('Ошибка при удалении')
  }
}

const clearQueue = async () => {
  if (!confirm('ВНИМАНИЕ: Это удалит ВСЕ письма из очереди. Продолжить?')) return
  try {
    await api.delete('/system/queue/all')
    fetchQueue()
  } catch (error) {
    alert('Ошибка при очистке очереди')
  }
}

const formatSize = (kb) => {
  if (kb > 1024) return (kb / 1024).toFixed(1) + ' MB'
  return kb + ' KB'
}

onMounted(fetchQueue)
</script>

<template>
  <div class="space-y-8 animate-in fade-in duration-500">
    <header class="flex justify-between items-end">
      <div>
        <h1 class="text-4xl font-black text-slate-900 dark:text-white tracking-tight">
          {{ t('menu.queue') }}
        </h1>
        <p class="text-slate-500 dark:text-slate-400 mt-2 font-medium">
          Управление зависшими сообщениями в Postfix (MailQ)
        </p>
      </div>
      <div class="flex gap-4">
        <button @click="fetchQueue" :disabled="refreshing" class="px-6 py-3 bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl font-bold text-xs uppercase tracking-widest hover:bg-slate-50 transition-all flex items-center gap-2">
          <svg class="w-4 h-4" :class="{'animate-spin': refreshing}" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" /></svg>
          Обновить
        </button>
        <button @click="flushQueue" class="px-6 py-3 bg-emerald-600 text-white rounded-2xl font-bold text-xs uppercase tracking-widest shadow-lg shadow-emerald-500/20 hover:bg-emerald-700 transition-all">
          Flush Queue
        </button>
        <button @click="clearQueue" class="px-6 py-3 bg-red-500 text-white rounded-2xl font-bold text-xs uppercase tracking-widest shadow-lg shadow-red-500/20 hover:bg-red-600 transition-all">
          Delete All
        </button>
      </div>
    </header>

    <div v-if="loading" class="space-y-4">
      <div v-for="i in 5" :key="i" class="h-16 bg-slate-100 dark:bg-slate-800 animate-pulse rounded-2xl"></div>
    </div>

    <div v-else-if="queue.length === 0" class="glass-panel p-20 text-center text-slate-400">
       <svg class="mx-auto h-16 w-16 opacity-10 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
      </svg>
      <p class="text-xl font-bold italic">Почтовая очередь пуста</p>
      <p class="mt-2 text-sm">Все сообщения доставлены или очередь еще не заполнена</p>
    </div>

    <div v-else class="glass-panel overflow-hidden">
      <table class="w-full text-left border-collapse">
        <thead>
          <tr class="bg-slate-50/50 dark:bg-slate-800/10 border-b border-slate-100 dark:border-slate-800/50">
            <th class="px-8 py-5 text-[10px] font-black uppercase tracking-widest text-slate-400">ID / Дата</th>
            <th class="px-8 py-5 text-[10px] font-black uppercase tracking-widest text-slate-400">Отправитель</th>
            <th class="px-8 py-5 text-[10px] font-black uppercase tracking-widest text-slate-400">Получатели</th>
            <th class="px-8 py-5 text-[10px] font-black uppercase tracking-widest text-slate-400 text-right">Размер</th>
            <th class="px-8 py-5 text-[10px] font-black uppercase tracking-widest text-slate-400 text-center">Действия</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-50 dark:divide-slate-800/50">
          <tr v-for="item in queue" :key="item.id" class="group hover:bg-slate-50/50 dark:hover:bg-slate-800/20 transition-colors">
            <td class="px-8 py-6">
              <div class="flex flex-col">
                <span class="font-black text-slate-900 dark:text-white tracking-tight break-all">{{ item.id }}</span>
                <span class="text-[10px] font-bold text-slate-400 mt-1">{{ item.arrival }}</span>
              </div>
            </td>
            <td class="px-8 py-6">
              <span class="text-sm font-bold text-slate-700 dark:text-slate-300 break-all">{{ item.sender }}</span>
            </td>
            <td class="px-8 py-6">
              <div class="flex flex-col gap-1">
                <span v-for="rcp in item.recipient" :key="rcp" class="text-xs font-medium text-slate-600 dark:text-slate-400 break-all">{{ rcp }}</span>
                <div v-if="item.reason" class="mt-2 p-3 bg-red-50 dark:bg-red-900/10 border border-red-100 dark:border-red-900/20 rounded-xl text-[10px] font-bold text-red-600 dark:text-red-400 leading-relaxed italic">
                  原因: {{ item.reason }}
                </div>
              </div>
            </td>
            <td class="px-8 py-6 text-right">
              <span class="text-xs font-black text-slate-500 bg-slate-100 dark:bg-slate-800 px-2.5 py-1 rounded-lg">{{ formatSize(item.size) }}</span>
            </td>
            <td class="px-8 py-6 text-center">
              <button @click="deleteItem(item.id)" class="p-2.5 bg-red-50 dark:bg-red-900/20 text-red-500 hover:bg-red-500 hover:text-white rounded-xl transition-all shadow-sm">
                <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
