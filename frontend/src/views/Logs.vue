<script setup>
import { ref, onMounted, watch } from 'vue'
import api from '@/api/axios'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const logs = ref([])
const loading = ref(true)
const page = ref(1)
const totalLogs = ref(0)
const limit = 20

const fetchLogs = async () => {
  loading.value = true
  try {
    const response = await api.get('/logs', {
      params: {
        page: page.value,
        limit: limit
      }
    })
    logs.value = response.data
    totalLogs.value = parseInt(response.headers['x-total-count'] || 0)
  } catch (error) {
    console.error('Failed to fetch logs:', error)
  } finally {
    loading.value = false
  }
}

const translateAction = (action) => {
  if (!action) return ''
  const act = action.toLowerCase().replace(/ /g, '_')
  const dictionary = {
    'create_domain': t('actions.create_domain'),
    'update_domain': t('actions.update_domain'),
    'delete_domain': t('actions.delete_domain'),
    'create_mailbox': t('actions.create_mailbox'),
    'create_mailbox_+_alias': t('actions.create_mailbox_alias'),
    'update_mailbox': t('actions.update_mailbox'),
    'delete_mailbox': t('actions.delete_mailbox'),
    'create_alias': t('actions.create_alias'),
    'update_alias': t('actions.update_alias'),
    'delete_alias': t('actions.delete_alias'),
    'create_admin': t('actions.create_admin'),
    'delete_admin': t('actions.delete_admin'),
    'login': t('actions.login'),
    'enable_vacation': t('actions.enable_vacation'),
    'update_vacation': t('actions.update_vacation'),
    'broadcast': t('tools.broadcast.title'),
    'send_email': t('tools.send_mail.title'),
    'batch_create_mailboxes': t('actions.batch_create_mailboxes'),
    'batch_delete_mailboxes': t('actions.batch_delete_mailboxes'),
    'batch_status_update': t('actions.batch_status_update'),
    'create_domain_alias': t('actions.create_domain_alias'),
    'delete_domain_alias': t('actions.delete_domain_alias')
  }
  return dictionary[act] || action
}

const formatDate = (dateStr) => {
  return new Date(dateStr).toLocaleString()
}

watch(page, fetchLogs)

onMounted(fetchLogs)
</script>

<template>
  <div class="space-y-6">
    <header class="flex justify-between items-center">
      <div>
        <h1 class="text-3xl font-bold tracking-tight text-slate-900 dark:text-white text-left">{{ t('logs.title') }}</h1>
        <p class="text-slate-500 dark:text-slate-400 text-left">{{ t('logs.subtitle') }}</p>
      </div>
    </header>

    <div class="glass-panel overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left">
          <thead>
            <tr class="border-b border-slate-100 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/50">
              <th class="px-6 py-4 text-xs font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400">{{ t('logs.table.datetime') }}</th>
              <th class="px-6 py-4 text-xs font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400">{{ t('logs.table.admin') }}</th>
              <th class="px-6 py-4 text-xs font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400">{{ t('logs.table.domain') }}</th>
              <th class="px-6 py-4 text-xs font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400">{{ t('logs.table.action') }}</th>
              <th class="px-6 py-4 text-xs font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400">{{ t('logs.table.data') }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 dark:divide-slate-800">
            <tr v-if="loading" v-for="i in 5" :key="i" class="animate-pulse">
              <td colspan="5" class="px-6 py-4"><div class="h-4 bg-slate-100 dark:bg-slate-800 rounded w-full"></div></td>
            </tr>
            <tr v-else-if="logs.length === 0">
              <td colspan="5" class="px-6 py-12 text-center text-slate-400 italic">{{ t('logs.no_entries') }}</td>
            </tr>
            <tr v-for="log in logs" :key="log.id" class="hover:bg-slate-50 dark:hover:bg-slate-800/40 transition-colors">
              <td class="px-6 py-4 text-sm font-medium text-slate-500 dark:text-slate-400 whitespace-nowrap">
                {{ formatDate(log.timestamp) }}
              </td>
              <td class="px-6 py-4 text-sm font-bold text-slate-900 dark:text-white">
                {{ log.username }}
              </td>
              <td class="px-6 py-4 text-sm">
                <span class="px-2 py-1 rounded-md bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 font-medium">
                  {{ log.domain }}
                </span>
              </td>
              <td class="px-6 py-4">
                <span :class="{
                  'px-2 py-1 rounded-full text-[10px] font-bold uppercase tracking-wider': true,
                  'bg-emerald-100 text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-400': log.action?.includes('create'),
                  'bg-red-100 text-red-700 dark:bg-red-500/10 dark:text-red-400': log.action?.includes('delete'),
                  'bg-mail-blue-100 text-mail-blue-700 dark:bg-mail-blue-500/20 dark:text-mail-blue-400': log.action?.includes('update')
                }">
                  {{ translateAction(log.action) }}
                </span>
              </td>
              <td class="px-6 py-4 text-sm text-slate-600 dark:text-slate-400 font-medium">
                {{ log.data }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Пагинация -->
      <div class="px-6 py-4 border-t border-slate-100 dark:border-slate-800 flex items-center justify-between bg-slate-50/30 dark:bg-slate-800/30">
        <div class="text-sm text-slate-500">
          {{ t('logs.pagination.shown') }} {{ logs.length }} {{ t('logs.pagination.of') }} {{ totalLogs }} {{ t('logs.pagination.entries') }}
        </div>
        <div class="flex gap-2">
          <button 
            @click="page--" 
            :disabled="page === 1 || loading"
            class="px-4 py-2 rounded-xl bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-sm font-bold disabled:opacity-50 transition-all hover:bg-slate-50 dark:hover:bg-slate-700"
          >
            {{ t('logs.pagination.prev') }}
          </button>
          <button 
            @click="page++" 
            :disabled="page * limit >= totalLogs || loading"
            class="px-4 py-2 rounded-xl bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-sm font-bold disabled:opacity-50 transition-all hover:bg-slate-50 dark:hover:bg-slate-700"
          >
            {{ t('logs.pagination.next') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
</style>
