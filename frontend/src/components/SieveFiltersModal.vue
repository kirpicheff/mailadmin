<script setup>
import { ref, watch, computed } from 'vue'
import api from '@/api/axios'
import { useI18n } from 'vue-i18n'

const props = defineProps({
  show: Boolean,
  username: String // Email ящика или "GLOBAL"
})

const emit = defineEmits(['close', 'saved'])
const { t } = useI18n()

const loading = ref(false)
const saving = ref(false)
const filters = ref([])

const fields = computed(() => [
  { value: 'Subject', label: t('sieve.fields.subject') },
  { value: 'From', label: t('sieve.fields.from') },
  { value: 'To', label: t('sieve.fields.to') },
  { value: 'X-Spam-Flag', label: t('sieve.fields.spam') },
  { value: 'Body', label: t('sieve.fields.body') }
])

const operators = computed(() => [
  { value: 'contains', label: t('sieve.operators.contains') },
  { value: 'not_contains', label: t('sieve.operators.not_contains') },
  { value: 'is', label: t('sieve.operators.is') },
  { value: 'not_is', label: t('sieve.operators.not_is') },
  { value: 'matches', label: t('sieve.operators.matches') },
  { value: 'regex', label: t('sieve.operators.regex') }
])

const actionTypes = computed(() => [
  { value: 'fileinto', label: t('sieve.actions.fileinto') },
  { value: 'redirect', label: t('sieve.actions.redirect') },
  { value: 'discard', label: t('sieve.actions.discard') },
  { value: 'reject', label: t('sieve.actions.reject') },
  { value: 'setflag', label: t('sieve.actions.mark_read') },
  { value: 'vacation', label: t('sieve.actions.vacation') }
])

const standardFolders = computed(() => [
  { value: 'INBOX', label: t('folders.inbox') },
  { value: 'Sent', label: t('folders.sent') },
  { value: 'Junk', label: t('folders.junk') },
  { value: 'Trash', label: t('folders.trash') },
  { value: 'Drafts', label: t('folders.drafts') },
  { value: 'Archive', label: t('folders.archive') }
])

const isCustomFolder = (target) => {
  return target && !standardFolders.value.some(f => f.value === target)
}

const fetchFilters = async () => {
  if (!props.username) return
  loading.value = true
  try {
    const { data } = await api.get(`/sieve/${props.username}`)
    const raw = JSON.parse(data.rules_json || '[]')
    
    // Проверка формата (новый формат содержит массив conditions)
    if (raw.length > 0 && !Object.prototype.hasOwnProperty.call(raw[0], 'conditions')) {
      filters.value = raw.map(r => ({
        name: 'Правило: ' + (r.field || 'New'),
        match_all: true,
        active: true,
        conditions: [{ field: r.field || 'Subject', operator: r.operator || 'contains', value: r.value || '' }],
        actions: [{ type: r.action || 'fileinto', target: r.target || 'INBOX' }]
      }))
    } else {
      filters.value = raw.map(r => ({
        ...r,
        name: r.title || r.name || r.label || t('sieve.add_filter')
      }))
    }
  } catch (err) {
    console.error('Failed to fetch sieve filters:', err)
  } finally {
    loading.value = false
  }
}

const addFilter = () => {
  filters.value.push({
    name: t('sieve.add_filter'),
    match_all: true,
    active: true,
    conditions: [{ field: 'Subject', operator: 'contains', value: '' }],
    actions: [{ type: 'fileinto', target: 'INBOX' }]
  })
}

const addCondition = (filterIndex) => {
  filters.value[filterIndex].conditions.push({ field: 'Subject', operator: 'contains', value: '' })
}

const removeCondition = (filterIndex, condIndex) => {
  filters.value[filterIndex].conditions.splice(condIndex, 1)
}

const addAction = (filterIndex) => {
  filters.value[filterIndex].actions.push({ type: 'fileinto', target: 'INBOX' })
}

const removeAction = (filterIndex, actionIndex) => {
  filters.value[filterIndex].actions.splice(actionIndex, 1)
}

const removeFilter = (index) => {
  filters.value.splice(index, 1)
}

const saveFilters = async () => {
  saving.value = true
  try {
    await api.post(`/sieve/${props.username}`, {
      rules_json: JSON.stringify(filters.value),
      active: true
    })
    emit('saved')
    emit('close')
  } catch (err) {
    alert('Ошибка: ' + (err.response?.data?.error || err.message))
  } finally {
    saving.value = false
  }
}

const importing = ref(false)

const importFromServer = async () => {
  if (!confirm('Это действие перезапишет текущие правила в базе данных. Продолжить?')) return
  importing.value = true
  try {
    const { data } = await api.post(`/sieve/${props.username}/import`)
    const raw = JSON.parse(data.rules_json || '[]')
    
    if (raw.length > 0 && !Object.prototype.hasOwnProperty.call(raw[0], 'conditions')) {
      filters.value = raw.map(r => ({
        name: 'Правило: ' + (r.field || 'New'),
        match_all: true,
        active: true,
        conditions: [{ field: r.field || 'Subject', operator: r.operator || 'contains', value: r.value || '' }],
        actions: [{ type: r.action || 'fileinto', target: r.target || 'INBOX' }]
      }))
    } else {
      filters.value = raw.map(r => ({
        ...r,
        name: r.title || r.name || r.label || t('sieve.add_filter')
      }))
    }
    alert(t('sieve.import_success'))
  } catch (err) {
    alert('Ошибка: ' + (err.response?.data?.error || err.message))
  } finally {
    importing.value = false
  }
}

watch(() => props.show, (newVal) => {
  if (newVal) fetchFilters()
})
</script>

<template>
  <div v-if="show" class="fixed inset-0 bg-slate-950/60 backdrop-blur-xl flex items-center justify-center p-4 z-[70]">
    <div class="bg-white dark:bg-slate-900 rounded-[32px] w-full max-w-4xl shadow-2xl border border-white/10 dark:border-slate-800/50 animate-in zoom-in duration-300 overflow-hidden flex flex-col max-h-[95vh]">
      
      <!-- Header -->
      <div class="px-8 py-6 flex justify-between items-center bg-slate-50/50 dark:bg-slate-800/30 border-b border-slate-100 dark:border-slate-800/50">
        <div class="flex items-center gap-4">
          <div :class="username === 'GLOBAL' ? 'bg-indigo-600' : 'bg-mail-blue-600'" class="w-12 h-12 rounded-2xl text-white flex items-center justify-center shadow-lg">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z" /></svg>
          </div>
          <div>
            <h3 class="text-xl font-extrabold text-slate-900 dark:text-white tracking-tight">
              {{ t('sieve.title') }}
            </h3>
            <p class="text-[10px] font-bold text-slate-400 uppercase tracking-widest">{{ username }}</p>
          </div>
        </div>
        <button @click="$emit('close')" class="text-slate-400 hover:text-slate-600 dark:hover:text-white transition-all bg-white dark:bg-slate-800 p-2 rounded-xl border border-slate-100 dark:border-slate-700">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M6 18L18 6M6 6l12 12" /></svg>
        </button>
      </div>

      <!-- Main Content -->
      <div class="p-8 overflow-y-auto space-y-8 flex-1 scrollbar-thin bg-slate-50/20 dark:bg-slate-900/50">
        <div v-if="loading" class="flex flex-col items-center py-20">
          <div class="w-12 h-12 rounded-full border-4 border-slate-100 dark:border-slate-800 border-t-mail-blue-500 animate-spin mb-4"></div>
        </div>

        <div v-else class="space-y-6">
          <div v-for="(filter, fIdx) in filters" :key="fIdx" class="bg-white dark:bg-slate-800/40 rounded-[32px] border border-slate-200/60 dark:border-slate-800/80 shadow-sm overflow-hidden animate-in slide-in-from-top-4 duration-500">
            
            <!-- Filter Header -->
            <div class="px-6 py-4 bg-slate-50/50 dark:bg-slate-800/60 border-b border-slate-100 dark:border-slate-700/50 flex items-center gap-4">
              <input v-model="filter.name" type="text" :placeholder="t('common.name')" 
                class="bg-transparent border-none focus:ring-0 text-sm font-black text-slate-700 dark:text-white flex-1 outline-none" />
              
              <div class="flex items-center gap-2">
                <span class="text-[10px] font-bold uppercase text-slate-400">{{ t('common.active') }}</span>
                <button @click="filter.active = !filter.active" :class="filter.active ? 'bg-green-500' : 'bg-slate-300 dark:bg-slate-700'" class="relative inline-flex h-4 w-8 items-center rounded-full transition-all">
                  <span :class="filter.active ? 'translate-x-4' : 'translate-x-1'" class="inline-block h-2.5 w-2.5 transform rounded-full bg-white transition-transform" />
                </button>
                <button @click="removeFilter(fIdx)" class="ml-2 text-slate-300 hover:text-red-500 transition-colors">
                  <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
                </button>
              </div>
            </div>

            <div class="p-6 space-y-6">
              <!-- Conditions -->
              <div>
                <div class="flex items-center justify-between mb-4">
                  <div class="flex items-center gap-3">
                    <span class="text-[10px] font-black uppercase tracking-widest text-slate-400">{{ t('mailboxes.filters.show') }}</span>
                    <select v-model="filter.match_all" class="bg-slate-100 dark:bg-slate-800 border-none rounded-lg text-[9px] font-black uppercase px-2 py-1 outline-none">
                      <option :value="true">{{ t('sieve.match_all') }}</option>
                      <option :value="false">{{ t('sieve.match_any') }}</option>
                    </select>
                  </div>
                  <button @click="addCondition(fIdx)" class="text-[9px] font-black uppercase py-1 px-2 border-2 border-dashed border-slate-200 dark:border-slate-700 rounded-lg text-slate-400 hover:text-mail-blue-500 hover:border-mail-blue-500 transition-all">
                    + {{ t('sieve.add_condition') }}
                  </button>
                </div>

                <div class="space-y-3">
                  <div v-for="(cond, cIdx) in filter.conditions" :key="cIdx" class="flex flex-wrap items-center gap-3 animate-in fade-in duration-300">
                    <select v-model="cond.field" class="h-9 px-3 bg-slate-50 dark:bg-slate-900/50 border border-slate-100 dark:border-slate-700 rounded-xl text-xs font-bold focus:border-mail-blue-500 outline-none transition-all">
                      <option v-for="f in fields" :key="f.value" :value="f.value">{{ f.label }}</option>
                    </select>
                    
                    <select v-if="cond.field !== 'X-Spam-Flag'" v-model="cond.operator" class="h-9 px-3 bg-slate-50 dark:bg-slate-900/50 border border-slate-100 dark:border-slate-700 rounded-xl text-xs font-bold focus:border-mail-blue-500 outline-none transition-all">
                      <option v-for="op in operators" :key="op.value" :value="op.value">{{ op.label }}</option>
                    </select>

                    <input v-if="cond.field !== 'X-Spam-Flag'" v-model="cond.value" type="text" placeholder="..." 
                      class="flex-1 min-w-[150px] h-9 px-4 bg-slate-50 dark:bg-slate-900/50 border border-slate-100 dark:border-slate-700 rounded-xl text-xs font-bold focus:border-mail-blue-500 outline-none" />

                    <button @click="removeCondition(fIdx, cIdx)" v-if="filter.conditions.length > 1" class="text-slate-300 hover:text-red-500 transition-colors">
                      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M6 18L18 6M6 6l12 12" /></svg>
                    </button>
                  </div>
                </div>
              </div>

              <!-- Actions -->
              <div class="pt-6 border-t border-slate-100 dark:border-slate-700/50">
                <div class="flex items-center justify-between mb-4">
                  <span class="text-[10px] font-black uppercase tracking-widest text-slate-400">{{ t('common.actions') }}</span>
                  <button @click="addAction(fIdx)" class="text-[9px] font-black uppercase py-1 px-2 border-2 border-dashed border-slate-200 dark:border-slate-700 rounded-lg text-slate-400 hover:text-mail-blue-500 hover:border-mail-blue-500 transition-all">
                    + {{ t('sieve.add_action') }}
                  </button>
                </div>

                <div class="space-y-3">
                  <div v-for="(action, aIdx) in filter.actions" :key="aIdx" class="flex flex-wrap items-center gap-3 animate-in fade-in duration-300">
                    <select v-model="action.type" class="h-9 px-3 bg-slate-100 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-xl text-xs font-bold text-slate-700 dark:text-slate-300 focus:border-mail-blue-500 outline-none transition-all">
                      <option v-for="a in actionTypes" :key="a.value" :value="a.value">{{ a.label }}</option>
                    </select>

                    <template v-if="action.type === 'fileinto'">
                       <select v-model="action.target" class="h-9 px-3 bg-slate-50 dark:bg-slate-900/50 border border-slate-100 dark:border-slate-700 rounded-xl text-xs font-bold focus:border-mail-blue-500 outline-none transition-all">
                          <option v-for="f in standardFolders" :key="f.value" :value="f.value">{{ f.label }}</option>
                          <option value="CUSTOM">-- {{ t('folders.custom') }} --</option>
                       </select>
                       <input v-if="action.target === 'CUSTOM' || isCustomFolder(action.target)" v-model="action.target" type="text" :placeholder="t('folders.custom')" 
                        class="flex-1 min-w-[150px] h-9 px-4 bg-slate-50 dark:bg-slate-900/50 border border-slate-100 dark:border-slate-700 rounded-xl text-xs font-bold focus:border-mail-blue-500 outline-none" />
                    </template>

                    <template v-else-if="['redirect', 'reject', 'setflag', 'vacation'].includes(action.type)">
                       <textarea v-if="action.type === 'vacation'" v-model="action.target" :placeholder="t('mailbox_form.vacation.body')" 
                        class="w-full mt-2 p-4 bg-slate-50 dark:bg-slate-900/50 border border-slate-100 dark:border-slate-700 rounded-2xl text-xs font-bold focus:border-mail-blue-500 outline-none scrollbar-thin min-h-[100px]"></textarea>
                       
                       <input v-else v-model="action.target" type="text" :placeholder="action.type === 'setflag' ? '\\Seen' : t('common.search')" 
                        class="flex-1 min-w-[200px] h-9 px-4 bg-slate-50 dark:bg-slate-900/50 border border-slate-100 dark:border-slate-700 rounded-xl text-xs font-bold focus:border-mail-blue-500 outline-none" />
                    </template>

                    <button @click="removeAction(fIdx, aIdx)" v-if="filter.actions.length > 1" class="text-slate-300 hover:text-red-500 transition-colors">
                      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M6 18L18 6M6 6l12 12" /></svg>
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <button @click="addFilter" class="w-full py-6 border-2 border-dashed border-slate-200 dark:border-slate-800 rounded-[32px] text-slate-400 font-bold hover:border-mail-blue-500 hover:text-mail-blue-500 transition-all flex items-center justify-center gap-3 active:scale-[0.99] bg-white/50 dark:bg-slate-800/20">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M12 6v6m0 0v6m0-6h6m-6 0H6" /></svg>
            {{ t('sieve.add_filter') }}
          </button>
        </div>
      </div>

      <!-- Footer Actions -->
      <div class="px-8 py-6 bg-slate-50 dark:bg-slate-800/30 border-t border-slate-100 dark:border-slate-800/50 flex gap-4">
        <button @click="importFromServer" :disabled="saving || loading || importing" class="px-6 py-4 bg-indigo-600 hover:bg-indigo-700 text-white rounded-2xl font-black uppercase tracking-widest text-xs shadow-xl shadow-indigo-500/30 transition-all hover:-translate-y-0.5 disabled:opacity-50 flex items-center justify-center">
            <span v-if="importing" class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin mr-2"></span>
            {{ t('sieve.import_from_server') }}
        </button>
        <button @click="saveFilters" :disabled="saving || loading || importing" class="flex-1 py-4 bg-mail-blue-600 text-white rounded-2xl font-black uppercase tracking-widest text-xs hover:bg-mail-blue-700 shadow-xl shadow-mail-blue-500/30 transition-all hover:-translate-y-0.5 disabled:opacity-50 flex items-center justify-center">
            <span v-if="saving" class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin mr-2"></span>
            {{ t('domains.modal.save_changes') }}
        </button>
        <button @click="$emit('close')" class="px-10 py-4 bg-white dark:bg-slate-800 text-slate-500 rounded-2xl font-black uppercase tracking-widest text-xs hover:bg-slate-100 dark:hover:bg-slate-700 transition-all">
            {{ t('common.cancel') }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.scrollbar-thin::-webkit-scrollbar { width: 6px; }
.scrollbar-thin::-webkit-scrollbar-track { background: transparent; }
.scrollbar-thin::-webkit-scrollbar-thumb { background: #cbd5e1; border-radius: 10px; }
.dark .scrollbar-thin::-webkit-scrollbar-thumb { background: #334155; }
</style>
