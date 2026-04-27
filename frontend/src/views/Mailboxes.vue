<template>
  <div class="space-y-6 animate-in fade-in slide-in-from-bottom-2 duration-500">
    <!-- Header: Selector & Global Search -->
    <div class="glass-panel p-6 rounded-[32px] border border-slate-200 dark:border-slate-800 shadow-xl flex flex-col md:flex-row items-center gap-6">
      <div class="w-full md:w-1/3 relative">
        <label class="block text-[10px] font-black uppercase tracking-widest text-slate-400 mb-2 ml-1">{{ t('mailboxes.select_domain') }}</label>
        <select v-model="selectedDomain" @change="fetchData" 
          class="w-full pl-4 pr-10 py-3 rounded-2xl border-2 border-slate-100 dark:border-slate-800 bg-white dark:bg-slate-900 text-slate-900 dark:text-white font-bold focus:border-mail-blue-500 transition-all outline-none appearance-none cursor-pointer">
          <option value="">{{ t('mailboxes.select_prompt') }}</option>
          <option v-for="d in domains" :key="d.domain" :value="d.domain">{{ d.domain }}</option>
        </select>
        <div class="absolute right-4 bottom-3.5 pointer-events-none text-slate-400">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" /></svg>
        </div>
      </div>

      <div class="w-full md:w-2/3 relative group">
        <label class="block text-[10px] font-black uppercase tracking-widest text-slate-400 mb-2 ml-1">{{ t('common.search') }}</label>
        <div class="relative">
          <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none text-slate-400 group-focus-within:text-mail-blue-500 transition-colors">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" /></svg>
          </div>
          <input v-model="searchQuery" type="text" :placeholder="t('mailboxes.search_placeholder')"
            class="w-full pl-11 pr-4 py-3 rounded-2xl border-2 border-slate-100 dark:border-slate-800 bg-white dark:bg-slate-900 text-slate-900 dark:text-white focus:border-mail-blue-500 transition-all outline-none" />
        </div>
      </div>
    </div>

    <!-- Domain Stats (скрываем при поиске) -->
    <div v-if="selectedDomain && !searchQuery" class="mb-10 animate-in fade-in slide-in-from-top-4 duration-700">
      <div class="glass-panel p-10 text-center relative overflow-hidden group">
        <h2 class="text-xl font-extrabold text-slate-900 dark:text-white tracking-tight mb-4">{{ t('mailboxes.overview_for') }} {{ selectedDomain }} :</h2>
        <div class="flex flex-col items-center gap-1 text-sm font-bold text-slate-600 dark:text-slate-400">
          <p>{{ t('domains.table.aliases') }}: {{ stats.aliases_count }} <span class="text-slate-400">/</span> {{ stats.aliases === 0 ? t('common.unlimited') : stats.aliases }}</p>
          <p>{{ t('domains.table.mailboxes') }}: {{ stats.mailboxes_count }} <span class="text-slate-400">/</span> {{ stats.mailboxes === 0 ? t('common.unlimited') : stats.mailboxes }}</p>
        </div>

        <!-- Filters Navigation -->
        <div class="flex flex-wrap justify-center items-center gap-2 pt-4 text-xs font-black uppercase tracking-widest border-t border-slate-100 dark:border-slate-800/50 mt-4">
          <span class="text-slate-400">{{ t('mailboxes.filters.show') }}</span>
          <button @click="filterType = 'all'" :class="filterType === 'all' ? 'text-mail-blue-600' : 'text-slate-400 hover:text-slate-600 dark:hover:text-white'" class="transition-colors px-2">{{ t('mailboxes.filters.all') }}</button>
          <span class="text-slate-200 dark:text-slate-800">::</span>
          <button @click="filterType = 'boxes'" :class="filterType === 'boxes' ? 'text-mail-blue-600' : 'text-slate-400 hover:text-slate-600 dark:hover:text-white'" class="transition-colors px-2">{{ t('mailboxes.filters.boxes') }}</button>
          <span class="text-slate-200 dark:text-slate-800">::</span>
          <button @click="filterType = 'aliases'" :class="filterType === 'aliases' ? 'text-mail-blue-600' : 'text-slate-400 hover:text-slate-600 dark:hover:text-white'" class="transition-colors px-2">{{ t('mailboxes.filters.aliases') }}</button>
          <span class="text-slate-200 dark:text-slate-800">::</span>
          <button @click="filterType = 'domain_aliases'" :class="filterType === 'domain_aliases' ? 'text-mail-blue-600' : 'text-slate-400 hover:text-slate-600 dark:hover:text-white'" class="transition-colors px-2">{{ t('mailboxes.filters.domain_aliases') }}</button>
        </div>
      </div>
    </div>

    <!-- Actions Bar -->
    <div v-if="selectedDomain || searchQuery" class="flex flex-wrap items-center justify-center gap-4">
      <button @click="openMailboxModal()" class="px-6 py-3 bg-mail-blue-600 text-white rounded-2xl font-black uppercase tracking-widest text-[10px] hover:bg-mail-blue-700 shadow-xl shadow-mail-blue-500/20 transition-all active:scale-95 flex items-center gap-2">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M12 4v16m8-8H4" /></svg>
        {{ t('mailboxes.actions.create_box') }}
      </button>
      <button @click="openAliasModal()" class="px-6 py-3 bg-indigo-600 text-white rounded-2xl font-black uppercase tracking-widest text-[10px] hover:bg-indigo-700 shadow-xl shadow-indigo-500/20 transition-all active:scale-95 flex items-center gap-2">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M12 4v16m8-8H4" /></svg>
        {{ t('mailboxes.actions.create_alias') }}
      </button>
      <button @click="downloadCSV" class="px-6 py-3 bg-white dark:bg-slate-800 text-slate-800 dark:text-slate-200 rounded-2xl font-black uppercase tracking-widest text-[10px] hover:bg-slate-50 dark:hover:bg-slate-700 border border-slate-200 dark:border-slate-700 transition-all active:scale-95 flex items-center gap-2">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" /></svg>
        {{ t('mailboxes.actions.download_csv') }}
      </button>
    </div>

    <!-- Data Table -->
    <div v-if="selectedDomain || searchQuery" class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 shadow-2xl rounded-[32px] overflow-hidden">
      <div class="flex items-center justify-between px-8 py-6 border-b border-slate-100 dark:border-slate-800">
        <h2 class="text-xl font-extrabold text-slate-900 dark:text-white tracking-tight">
          {{ searchQuery ? t('mailboxes.table.results_search') : t('mailboxes.table.results_domain') }}
        </h2>
        <div v-if="searchQuery" class="px-3 py-1 bg-mail-blue-500/10 text-mail-blue-500 rounded-full text-[10px] font-black uppercase tracking-widest border border-mail-blue-500/20">
          {{ t('mailboxes.table.found') }}: {{ filteredItems.length }}
        </div>
      </div>
      <div class="overflow-x-auto">
        <table class="w-full text-left border-collapse">
          <thead>
            <tr class="bg-slate-50/50 dark:bg-slate-800/30 border-b border-slate-200 dark:border-slate-700/50">
              <th class="px-8 py-5 text-[10px] font-black text-slate-400 uppercase tracking-widest">{{ t('mailboxes.table.address_type') }}</th>
              <th class="px-8 py-5 text-[10px] font-black text-slate-400 uppercase tracking-widest">{{ t('mailboxes.table.target_info') }}</th>
              <th class="px-8 py-5 text-[10px] font-black text-slate-400 uppercase tracking-widest text-center">{{ t('mailboxes.table.status') }}</th>
              <th class="px-8 py-5 text-[10px] font-black text-slate-400 uppercase tracking-widest">{{ t('mailboxes.table.modified') }}</th>
              <th class="px-8 py-5 text-[10px] font-black text-slate-400 uppercase tracking-widest text-center">{{ t('common.actions') }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 dark:divide-slate-800/50 font-bold">
            <template v-for="item in filteredItems" :key="item.type + item.id">
              <tr class="group hover:bg-slate-50/50 dark:hover:bg-slate-800/30 transition-all duration-200">
                <td class="px-8 py-5">
                  <div class="flex items-center gap-3">
                    <div :class="getTypeColor(item.type)" class="w-2 h-8 rounded-full shadow-sm"></div>
                    <div class="flex flex-col">
                      <span class="text-sm font-bold text-slate-900 dark:text-white tracking-tight break-all">{{ item.address }}</span>
                      <div class="flex items-center gap-2 mt-1">
                        <span class="text-[9px] uppercase tracking-tighter text-slate-400 font-black">{{ item.typeName }}</span>
                        <span v-if="searchQuery" class="text-[9px] uppercase font-black text-mail-blue-500 bg-mail-blue-500/10 px-1.5 py-0.5 rounded border border-mail-blue-500/20">
                          {{ item.domain }}
                        </span>
                      </div>
                    </div>
                  </div>
                </td>
                <td class="px-8 py-5">
                  <div v-if="item.type === 'mailbox'" class="space-y-1">
                    <div class="text-xs text-slate-700 dark:text-slate-300">{{ item.name || '—' }}</div>
                    <div class="mt-2 space-y-1">
                      <div class="flex justify-between text-[9px] font-bold text-slate-500 uppercase tracking-tighter">
                        <span>{{ formatSize(item.quota_used) }} / {{ item.quota_mb > 0 ? item.quota_mb + ' МБ' : '∞' }}</span>
                        <span v-if="item.quota_mb > 0">{{ Math.round((item.quota_used / (item.quota_mb * 1024 * 1024)) * 100) }}%</span>
                        <span v-else>0%</span>
                      </div>
                      <div class="w-32 bg-slate-100 dark:bg-slate-800 rounded-full h-1 overflow-hidden">
                        <div 
                          class="h-full rounded-full transition-all duration-500" 
                          :class="(item.quota_used / (item.quota_mb * 1024 * 1024)) > 0.9 ? 'bg-red-500' : 'bg-mail-blue-500'"
                          :style="{ width: Math.min(100, (item.quota_mb > 0 ? (item.quota_used / (item.quota_mb * 1024 * 1024)) * 100 : 0)) + '%' }"
                        ></div>
                      </div>
                    </div>
                  </div>
                  <div v-else class="flex flex-col gap-1">
                    <div v-for="addr in item.goto_list" :key="addr" class="text-xs text-slate-600 dark:text-slate-400 truncate max-w-xs">
                      {{ addr }}
                    </div>
                  </div>
                </td>
              <td class="px-8 py-5 text-center">
                <button @click="toggleStatus(item)" :class="item.active ? 'bg-green-500 shadow-green-500/20' : 'bg-slate-200 dark:bg-slate-700'" class="relative inline-flex h-5 w-9 items-center rounded-full transition-all active:scale-95 mx-auto">
                  <span :class="item.active ? 'translate-x-5' : 'translate-x-1'" class="inline-block h-3.5 w-3.5 transform rounded-full bg-white transition-transform" />
                </button>
              </td>
              <td class="px-8 py-5">
                <span class="text-[11px] text-slate-500">{{ formatDate(item.modified) }}</span>
              </td>
              <td class="px-8 py-5">
                <div class="flex items-center gap-2 justify-center">
                  <button @click="editItem(item)" class="p-2 bg-slate-50 dark:bg-slate-800 text-slate-400 hover:text-mail-blue-500 hover:bg-white dark:hover:bg-slate-700 rounded-lg border border-slate-200 dark:border-slate-700 transition-all shadow-sm">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" /></svg>
                  </button>
                  <button @click="deleteItem(item)" class="p-2 bg-slate-50 dark:bg-slate-800 text-slate-400 hover:text-red-500 hover:bg-white dark:hover:bg-slate-700 rounded-lg border border-slate-200 dark:border-slate-700 transition-all shadow-sm">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
                  </button>
                </div>
              </td>
              </tr>
            </template>

            <!-- Load More -->
            <tr v-if="hasMore">
              <td colspan="5" class="px-8 py-8 text-center">
                <button @click="loadMore" :disabled="loading" class="px-10 py-3 bg-slate-100 dark:bg-slate-800 text-slate-500 rounded-2xl font-black uppercase tracking-widest text-[10px] hover:bg-mail-blue-600 hover:text-white transition-all shadow-sm">
                  {{ loading ? t('common.loading') : t('mailboxes.load_more') }}
                </button>
              </td>
            </tr>

            <tr v-if="filteredItems.length === 0 && !loading">
              <td colspan="5" class="px-8 py-20 text-center text-slate-400 font-bold italic">
                {{ searchQuery ? t('mailboxes.empty.search') : t('mailboxes.empty.section') }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Modals (Placeholder for now, implementation follows) -->
    <MailboxForm v-if="showMailboxModal" :domain="selectedDomain" :item="editingItem" @close="showMailboxModal = false" @save="onSaved" />
    <AliasForm v-if="showAliasModal" :domain="selectedDomain" :item="editingItem" @close="showAliasModal = false" @save="onSaved" />
  </div>
</template>

<script setup>
import { ref, onMounted, computed, watch, reactive } from 'vue'
import api from '@/api/axios'
import MailboxForm from '@/components/MailboxForm.vue'
import AliasForm from '@/components/AliasForm.vue'
import { useI18n } from 'vue-i18n'

const { t, locale } = useI18n()

const domains = ref([])
const selectedDomain = ref('')
const searchQuery = ref('')
const filterType = ref('all') // all, boxes, aliases, domain_aliases

const mailboxes = ref([])
const aliases = ref([])
const domainAliases = ref([])
const loading = ref(false)
const totalItems = ref(0)
const currentPage = ref(1)
const pageSize = ref(50)
const hasMore = ref(false)

const stats = reactive({
  aliases_count: 0,
  aliases: 0,
  mailboxes_count: 0,
  mailboxes: 0,
})

const showMailboxModal = ref(false)
const showAliasModal = ref(false)
const editingItem = ref(null)

const fetchDomains = async () => {
  try {
    const { data } = await api.get('/domains')
    domains.value = data
    if (data.length > 0 && !selectedDomain.value) {
      selectedDomain.value = data[0].domain
      fetchData()
    }
  } catch (err) {
    console.error('Error fetching domains:', err)
  }
}

const fetchData = async (reset = true) => {
  if (!selectedDomain.value && !searchQuery.value) return
  
  loading.value = true
  if (reset) {
    currentPage.value = 1
    mailboxes.value = []
    aliases.value = []
    domainAliases.value = []
  }

  // Обновляем статистику выбранного домена
  const domainInfo = domains.value.find(d => d.domain === selectedDomain.value)
  if (domainInfo) {
    Object.assign(stats, domainInfo)
  }

  try {
    const params = {
      page: currentPage.value,
      limit: pageSize.value,
      search: searchQuery.value || undefined,
      domain: searchQuery.value ? undefined : selectedDomain.value
    }

    const [boxesRes, aliasesRes, domAliasesRes] = await Promise.all([
      api.get('/mailboxes', { params }),
      api.get('/aliases', { params }),
      api.get('/aliases/domain-aliases', { params })
    ])

    if (reset) {
      mailboxes.value = boxesRes.data
      aliases.value = aliasesRes.data
      domainAliases.value = domAliasesRes.data
    } else {
      mailboxes.value = [...mailboxes.value, ...boxesRes.data]
      aliases.value = [...aliases.value, ...aliasesRes.data]
      domainAliases.value = [...domainAliases.value, ...domAliasesRes.data]
    }

    // Простая логика "есть еще"
    hasMore.value = boxesRes.data.length === pageSize.value || 
                    aliasesRes.data.length === pageSize.value
                    
  } catch (err) {
    console.error('Error fetching data:', err)
  } finally {
    loading.value = false
  }
}

const loadMore = () => {
  currentPage.value++
  fetchData(false)
}

// Следим за поиском с задержкой (debounce)
let searchTimeout
watch(searchQuery, () => {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    fetchData(true)
  }, 500)
})

const filteredItems = computed(() => {
  let items = []
  
  if (filterType.value === 'all' || filterType.value === 'boxes') {
    items.push(...mailboxes.value.map(box => ({
      ...box,
      id: box.username,
      address: box.username,
      type: 'mailbox',
      typeName: t('mailboxes.types.mailbox'),
      quota_mb: box.quota / (1024 * 1024)
    })))
  }

  if (filterType.value === 'all' || filterType.value === 'aliases') {
    items.push(...aliases.value
      .filter(al => al.address !== al.goto) // Скрываем алиасы, указывающие на самих себя
      .map(al => ({
        ...al,
        id: al.address,
        address: al.address,
        type: 'alias',
        typeName: t('mailboxes.types.alias'),
        goto_list: al.goto.split(',').map(s => s.trim()).filter(s => s)
      })))
  }

  if (filterType.value === 'all' || filterType.value === 'domain_aliases') {
    items.push(...domainAliases.value.map(dal => ({
      ...dal,
      id: dal.alias_domain,
      address: dal.alias_domain,
      type: 'domain_alias',
      typeName: t('mailboxes.types.domain_alias'),
      goto_list: [dal.target_domain]
    })))
  }

  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase()
    items = items.filter(i => 
      i.address.toLowerCase().includes(q) || 
      (i.name && i.name.toLowerCase().includes(q)) ||
      (i.goto && i.goto.toLowerCase().includes(q))
    )
  }

  return items.sort((a, b) => a.address.localeCompare(b.address))
})

const getTypeColor = (type) => {
  switch(type) {
    case 'mailbox': return 'bg-mail-blue-500'
    case 'alias': return 'bg-indigo-500'
    case 'domain_alias': return 'bg-amber-500'
    default: return 'bg-slate-400'
  }
}

const openMailboxModal = (item = null) => {
  editingItem.value = item
  showMailboxModal.value = true
}

const openAliasModal = (item = null) => {
  editingItem.value = item
  showAliasModal.value = true
}

const editItem = (item) => {
  if (item.type === 'mailbox') openMailboxModal(item)
  else openAliasModal(item)
}

const toggleStatus = async (item) => {
  try {
    const newStatus = !item.active
    const endpoint = item.type === 'mailbox' ? `/mailboxes/${item.address}` : 
                     (item.type === 'alias' ? `/aliases/${item.address}` : `/aliases/domain-aliases/${item.address}`)
    
    await api.put(endpoint, {
      ...item,
      active: newStatus
    })
    
    // Обновляем в исходном массиве для реактивности
    if (item.type === 'mailbox') {
      const target = mailboxes.value.find(m => m.username === item.address)
      if (target) target.active = newStatus
    } else if (item.type === 'alias') {
      const target = aliases.value.find(a => a.address === item.address)
      if (target) target.active = newStatus
    } else if (item.type === 'domain_alias') {
      const target = domainAliases.value.find(a => a.alias_domain === item.address)
      if (target) target.active = newStatus
    }
  } catch (err) {
    alert(t('mailboxes.errors.status_change'))
  }
}

const deleteItem = async (item) => {
  if (!confirm(t('mailboxes.delete_confirm', { address: item.address }))) return
  try {
    const endpoint = item.type === 'mailbox' ? `/mailboxes/${item.address}` : 
                     (item.type === 'alias' ? `/aliases/${item.address}` : `/aliases/domain-aliases/${item.address}`)
    await api.delete(endpoint)
    fetchData()
  } catch (err) {
    alert(t('mailboxes.errors.delete'))
  }
}

const onSaved = () => {
  showMailboxModal.value = false
  showAliasModal.value = false
  fetchData()
  fetchDomains() // Обновляем статистику в списке доменов
}

const formatDate = (dateStr) => {
  if (!dateStr || dateStr.startsWith('0001') || dateStr.startsWith('2000-01-01')) return '—'
  return new Date(dateStr).toLocaleDateString(locale.value === 'ru' ? 'ru-RU' : 'en-US', { 
    day: '2-digit', 
    month: '2-digit', 
    year: 'numeric'
  })
}

const formatSize = (bytes) => {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

const downloadCSV = () => {
  const headers = ['Address', 'Type', 'Target/Info', 'Status', 'Modified']
  const rows = filteredItems.value.map(item => [
    item.address,
    item.typeName,
    item.type === 'mailbox' ? item.name : item.goto_list.join('; '),
    item.active ? 'Active' : 'Inactive',
    formatDate(item.modified)
  ])

  let csvContent = "data:text/csv;charset=utf-8,\uFEFF" 
    + [headers, ...rows].map(e => e.map(cell => `"${cell}"`).join(",")).join("\n")

  const encodedUri = encodeURI(csvContent)
  const link = document.createElement("a")
  link.setAttribute("href", encodedUri)
  link.setAttribute("download", `mailadmin_${selectedDomain.value || 'export'}.csv`)
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}

onMounted(fetchDomains)
</script>

<style scoped>
.glass-panel {
  background: rgba(255, 255, 255, 0.7);
  backdrop-filter: blur(20px);
}
.dark .glass-panel {
  background: rgba(15, 23, 42, 0.6);
}
</style>
