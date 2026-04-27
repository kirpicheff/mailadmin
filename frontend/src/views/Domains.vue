<template>
  <div class="space-y-6 animate-in fade-in slide-in-from-bottom-2 duration-500">
    <!-- Header -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-extrabold text-slate-900 dark:text-white tracking-tight">Почтовые домены</h1>
        <p class="text-sm text-slate-500 dark:text-slate-400 mt-1">Управление ресурсами, лимитами и маршрутизацией</p>
      </div>
      <button @click="openCreateModal" class="px-5 py-2.5 bg-mail-blue-600 text-white rounded-xl font-bold hover:bg-mail-blue-700 shadow-lg shadow-mail-blue-500/20 transition-all active:scale-95 text-sm">
        Добавить домен
      </button>
    </div>

    <!-- Domains Table -->
    <div class="bg-white dark:bg-slate-900 overflow-hidden border border-slate-200 dark:border-slate-800 shadow-sm rounded-2xl">
      <div class="overflow-x-auto">
        <table class="w-full text-left border-collapse">
          <thead>
            <tr class="bg-slate-50/50 dark:bg-slate-800/30 border-b border-slate-200 dark:border-slate-700/50">
              <th class="px-6 py-4 text-[10px] font-bold text-slate-400 uppercase tracking-widest whitespace-nowrap">Домен</th>
              <th class="px-6 py-4 text-[10px] font-bold text-slate-400 uppercase tracking-widest whitespace-nowrap text-center">Алиасы</th>
              <th class="px-6 py-4 text-[10px] font-bold text-slate-400 uppercase tracking-widest whitespace-nowrap text-center">Ящики</th>
              <th class="px-6 py-4 text-[10px] font-bold text-slate-400 uppercase tracking-widest whitespace-nowrap text-center">Квота</th>
              <th class="px-6 py-4 text-[10px] font-bold text-slate-400 uppercase tracking-widest whitespace-nowrap text-center">Резерв MX</th>
              <th class="px-6 py-4 text-[10px] font-bold text-slate-400 uppercase tracking-widest whitespace-nowrap">Изменение</th>
              <th class="px-6 py-4 text-[10px] font-bold text-slate-400 uppercase tracking-widest text-center whitespace-nowrap">Активен</th>
              <th class="px-6 py-4 text-[10px] font-bold text-slate-400 uppercase tracking-widest whitespace-nowrap">Действия</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 dark:divide-slate-800/50">
            <tr v-for="d in domains" :key="d.domain" class="group hover:bg-slate-50/50 dark:hover:bg-slate-800/30 transition-all duration-200">
              <td class="px-6 py-4">
                <div class="flex flex-col">
                  <span class="text-sm font-bold text-slate-800 dark:text-slate-200">{{ d.domain }}</span>
                  <span class="text-[10px] text-slate-400 mt-0.5 font-medium uppercase tracking-tight">{{ d.description || 'Нет описания' }}</span>
                </div>
              </td>
              
              <td class="px-6 py-4 text-center">
                <span class="text-xs font-bold text-slate-700 dark:text-slate-300">
                  {{ d.aliases_count }} <span class="text-slate-400 font-normal">/</span> {{ d.aliases === 0 ? '∞' : d.aliases }}
                </span>
                <div class="w-16 h-1 bg-slate-100 dark:bg-slate-800 rounded-full mt-1.5 mx-auto overflow-hidden">
                  <div class="h-full bg-amber-400 rounded-full" :style="{ width: d.aliases === 0 ? '5%' : Math.min((d.aliases_count / d.aliases) * 100, 100) + '%' }"></div>
                </div>
              </td>

              <td class="px-6 py-4 text-center">
                <span class="text-xs font-bold text-slate-700 dark:text-slate-300">
                  {{ d.mailboxes_count }} <span class="text-slate-400 font-normal">/</span> {{ d.mailboxes === 0 ? '∞' : d.mailboxes }}
                </span>
                <div class="w-16 h-1 bg-slate-100 dark:bg-slate-800 rounded-full mt-1.5 mx-auto overflow-hidden">
                  <div class="h-full bg-mail-blue-500 rounded-full" :style="{ width: d.mailboxes === 0 ? '5%' : Math.min((d.mailboxes_count / d.mailboxes) * 100, 100) + '%' }"></div>
                </div>
              </td>

              <td class="px-6 py-4 text-center">
                <div class="flex flex-col items-center">
                  <span class="text-[10px] text-slate-400 font-bold uppercase mb-0.5">Лимит: {{ d.quota === 0 ? '∞' : d.quota }}</span>
                  <span class="text-xs font-bold text-slate-700 dark:text-slate-300">{{ d.quota_used }} МБ</span>
                </div>
              </td>

              <td class="px-6 py-4 text-center">
                <span :class="d.backupmx ? 'text-amber-500 font-bold' : 'text-slate-300'" class="text-[10px] uppercase font-black tracking-widest">
                  {{ d.backupmx ? 'Да' : 'Нет' }}
                </span>
              </td>

              <td class="px-6 py-4 whitespace-nowrap">
                <div class="flex flex-col">
                  <span class="text-[11px] font-bold text-slate-600 dark:text-slate-400">{{ formatDate(d.modified) }}</span>
                  <span class="text-[9px] font-bold uppercase text-slate-400 mt-0.5">Пароль: {{ d.password_expiry === 0 ? '∞' : d.password_expiry + 'дн' }}</span>
                </div>
              </td>

              <td class="px-6 py-4 text-center">
                <button @click="toggleDomainStatus(d)" :class="d.active ? 'bg-green-500 shadow-green-500/20' : 'bg-slate-200 dark:bg-slate-700'" class="relative inline-flex h-5 w-9 items-center rounded-full transition-all active:scale-95">
                  <span :class="d.active ? 'translate-x-5' : 'translate-x-1'" class="inline-block h-3.5 w-3.5 transform rounded-full bg-white transition-transform" />
                </button>
              </td>

              <td class="px-6 py-4">
                <div class="flex items-center gap-2 justify-center">
                  <button @click="editDomain(d)" class="p-2 bg-slate-50 dark:bg-slate-800 text-slate-400 hover:text-mail-blue-500 hover:bg-white dark:hover:bg-slate-700 rounded-lg border border-slate-200 dark:border-slate-700 transition-all shadow-sm">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" /></svg>
                  </button>
                  <button @click="deleteDomain(d.domain)" class="p-2 bg-slate-50 dark:bg-slate-800 text-slate-400 hover:text-red-500 hover:bg-white dark:hover:bg-slate-700 rounded-lg border border-slate-200 dark:border-slate-700 transition-all shadow-sm">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Create/Edit Modal -->
    <div v-if="showModal" class="fixed inset-0 bg-slate-950/40 backdrop-blur-md flex items-center justify-center p-4 z-50">
      <div class="bg-white dark:bg-slate-900 rounded-[32px] w-full max-w-2xl shadow-2xl border border-white/10 dark:border-slate-800/50 animate-in zoom-in duration-300 overflow-hidden">
        <!-- Header -->
        <div class="px-8 py-6 flex justify-between items-center bg-slate-50/50 dark:bg-slate-800/30 border-b border-slate-100 dark:border-slate-800/50">
          <div class="flex items-center gap-3">
            <div class="w-11 h-11 rounded-2xl bg-mail-blue-600 text-white flex items-center justify-center shadow-lg shadow-mail-blue-500/20">
              <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" /></svg>
            </div>
            <div>
              <h3 class="text-xl font-extrabold text-slate-900 dark:text-white tracking-tight">
                {{ isEdit ? 'Настройка домена' : 'Добавить домен' }}
              </h3>
              <p class="text-[10px] font-bold text-slate-400 uppercase tracking-widest">{{ isEdit ? form.domain : 'Новое почтовое пространство' }}</p>
            </div>
          </div>
          <button @click="showModal = false" class="text-slate-400 hover:text-slate-600 dark:hover:text-white transition-all bg-white dark:bg-slate-800 p-2 rounded-xl shadow-sm border border-slate-100 dark:border-slate-700 hover:scale-110 active:scale-95">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M6 18L18 6M6 6l12 12" /></svg>
          </button>
        </div>

        <form @submit.prevent="saveDomain" class="p-8 space-y-6">
          <!-- Основное -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-5">
            <div class="md:col-span-2 space-y-4">
              <div class="relative">
                <label class="block text-[10px] font-black uppercase tracking-widest text-slate-400 mb-1.5 ml-1">Имя домена</label>
                <div class="relative group">
                  <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none text-slate-400 group-focus-within:text-mail-blue-500 transition-colors">
                    <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 002 2 2 2 0 012 2v.627m3.232-1.359A9 9 0 113.732 3.732" /></svg>
                  </div>
                  <input v-model="form.domain" :disabled="isEdit" type="text" placeholder="example.com" required
                    class="w-full pl-11 pr-4 py-3.5 rounded-2xl border-2 border-slate-100 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/30 text-slate-900 dark:text-white focus:border-mail-blue-500 focus:ring-4 focus:ring-mail-blue-500/5 transition-all outline-none font-bold placeholder:font-normal" />
                </div>
              </div>

              <div class="relative">
                <label class="block text-[10px] font-black uppercase tracking-widest text-slate-400 mb-1.5 ml-1">Описание</label>
                <input v-model="form.description" type="text" placeholder="Например: Корпоративная почта"
                  class="w-full px-5 py-3.5 rounded-2xl border-2 border-slate-100 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/30 text-slate-900 dark:text-white focus:border-mail-blue-500 transition-all outline-none" />
              </div>
            </div>

            <!-- Группа: Ресурсы -->
            <div class="bg-slate-50/50 dark:bg-slate-800/20 p-6 rounded-[28px] border border-slate-100 dark:border-slate-800/50 space-y-5">
              <div class="flex items-center gap-2 text-slate-400">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" /></svg>
                <h4 class="text-[10px] font-black uppercase tracking-widest text-slate-400">Лимиты ресурсов</h4>
              </div>
              <div class="space-y-4">
                <div class="flex items-center justify-between">
                  <span class="text-xs font-bold text-slate-600 dark:text-slate-400">Алиасы</span>
                  <input v-model.number="form.aliases" type="number" class="w-16 h-9 text-center rounded-xl bg-white dark:bg-slate-900 border-2 border-slate-100 dark:border-slate-800 font-bold text-slate-900 dark:text-white text-xs shadow-sm focus:border-mail-blue-500 outline-none transition-all" />
                </div>
                <div class="flex items-center justify-between">
                  <span class="text-xs font-bold text-slate-600 dark:text-slate-400">Ящики</span>
                  <input v-model.number="form.mailboxes" type="number" class="w-16 h-9 text-center rounded-xl bg-white dark:bg-slate-900 border-2 border-slate-100 dark:border-slate-800 font-bold text-slate-900 dark:text-white text-xs shadow-sm focus:border-mail-blue-500 outline-none transition-all" />
                </div>
                <p class="text-[9px] text-center text-slate-400 font-bold uppercase tracking-tighter opacity-60">0 = ∞ | -1 = Выкл</p>
              </div>
            </div>

            <!-- Группа: Диск -->
            <div class="bg-slate-50/50 dark:bg-slate-800/20 p-6 rounded-[28px] border border-slate-100 dark:border-slate-800/50 space-y-5">
              <div class="flex items-center gap-2 text-slate-400">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" /></svg>
                <h4 class="text-[10px] font-black uppercase tracking-widest text-slate-400">Квоты (МБ)</h4>
              </div>
              <div class="space-y-4">
                <div class="flex items-center justify-between">
                  <span class="text-xs font-bold text-slate-600 dark:text-slate-400">Общая</span>
                  <input v-model.number="form.quota" type="number" class="w-20 h-9 text-center rounded-xl bg-white dark:bg-slate-900 border-2 border-slate-100 dark:border-slate-800 font-bold text-slate-900 dark:text-white text-xs shadow-sm focus:border-mail-blue-500 outline-none transition-all" />
                </div>
                <div class="flex items-center justify-between">
                  <span class="text-xs font-bold text-slate-600 dark:text-slate-400">На ящик</span>
                  <input v-model.number="form.maxquota" type="number" class="w-20 h-9 text-center rounded-xl bg-white dark:bg-slate-900 border-2 border-slate-100 dark:border-slate-800 font-bold text-slate-900 dark:text-white text-xs shadow-sm focus:border-mail-blue-500 outline-none transition-all" />
                </div>
                <p class="text-[9px] text-center text-slate-400 font-bold uppercase tracking-tighter opacity-60">0 = Без ограничений</p>
              </div>
            </div>

            <!-- Доп. настройки -->
            <div class="md:col-span-2 flex items-center justify-between bg-slate-50/30 dark:bg-slate-800/10 p-4 rounded-2xl border border-slate-100 dark:border-slate-800/50">
              <div class="flex flex-col gap-1 text-left">
                <label class="text-[10px] font-black uppercase tracking-widest text-slate-400 ml-1">Срок действия пароля (дней)</label>
                <p class="text-[9px] text-slate-400 ml-1">Автоматическое истечение доступа к ящикам</p>
              </div>
              <input v-model.number="form.password_expiry" type="number" class="w-24 h-10 px-4 rounded-xl border-2 border-slate-100 dark:border-slate-800 bg-white dark:bg-slate-900 font-bold outline-none focus:border-mail-blue-500 text-center" />
            </div>

            <!-- Тумблеры -->
            <div class="md:col-span-2 flex gap-4 pt-4 border-t border-slate-100 dark:border-slate-800/50 mt-2">
              <button type="button" @click="form.backupmx = !form.backupmx" :class="form.backupmx ? 'border-amber-500 bg-amber-500/5' : 'border-slate-200 dark:border-slate-800'" class="flex-1 flex flex-col items-center gap-2 p-4 rounded-[24px] border-2 transition-all">
                <span class="text-[10px] font-black uppercase tracking-widest" :class="form.backupmx ? 'text-amber-600' : 'text-slate-400'">Backup MX</span>
                <div :class="form.backupmx ? 'bg-amber-500' : 'bg-slate-300 dark:bg-slate-700'" class="relative inline-flex h-5 w-9 items-center rounded-full transition-colors">
                  <span :class="form.backupmx ? 'translate-x-5' : 'translate-x-1'" class="inline-block h-3 w-3 transform rounded-full bg-white transition-transform" />
                </div>
              </button>

              <button type="button" @click="form.active = !form.active" :class="form.active ? 'border-green-500 bg-green-500/5' : 'border-slate-200 dark:border-slate-800'" class="flex-1 flex flex-col items-center gap-2 p-4 rounded-[24px] border-2 transition-all">
                <span class="text-[10px] font-black uppercase tracking-widest" :class="form.active ? 'text-green-600' : 'text-slate-400'">Домен активен</span>
                <div :class="form.active ? 'bg-green-500' : 'bg-slate-300 dark:bg-slate-700'" class="relative inline-flex h-5 w-9 items-center rounded-full transition-colors">
                  <span :class="form.active ? 'translate-x-5' : 'translate-x-1'" class="inline-block h-3 w-3 transform rounded-full bg-white transition-transform" />
                </div>
              </button>
            </div>
          </div>

          <!-- Кнопки -->
          <div class="flex gap-4 pt-4">
            <button type="submit" class="flex-1 py-4.5 bg-mail-blue-600 text-white rounded-2xl font-black uppercase tracking-widest text-xs hover:bg-mail-blue-700 shadow-xl shadow-mail-blue-500/30 transition-all hover:-translate-y-0.5 active:translate-y-0">
              {{ isEdit ? 'Сохранить изменения' : 'Создать домен' }}
            </button>
            <button type="button" @click="showModal = false" class="px-8 py-4.5 bg-slate-100 dark:bg-slate-800 text-slate-500 rounded-2xl font-black uppercase tracking-widest text-xs hover:bg-slate-200 dark:hover:bg-slate-700 transition-all">
              Отмена
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, reactive } from 'vue'
import api from '@/api/axios'

const domains = ref([])
const showModal = ref(false)
const isEdit = ref(false)

const form = reactive({
  domain: '',
  description: '',
  aliases: 0,
  mailboxes: 0,
  maxquota: 1024,
  quota: 0,
  transport: 'virtual',
  backupmx: false,
  active: true,
  password_expiry: 3650
})

const fetchDomains = async () => {
  try {
    const { data } = await api.get('/domains')
    domains.value = data
  } catch (err) {
    console.error('Error fetching domains:', err)
  }
}

const openCreateModal = () => {
  isEdit.value = false
  Object.assign(form, {
    domain: '',
    description: '',
    aliases: 0,
    mailboxes: 0,
    maxquota: 1024,
    quota: 0,
    transport: 'virtual',
    backupmx: false,
    active: true,
    password_expiry: 3650
  })
  showModal.value = true
}

const editDomain = (domainData) => {
  isEdit.value = true
  // Обрезаем технические поля статистики для формы
  const { mailboxes_count, aliases_count, quota_used, ...pureDomain } = domainData
  Object.assign(form, pureDomain)
  showModal.value = true
}

const saveDomain = async () => {
  try {
    if (isEdit.value) {
      await api.put(`/domains/${form.domain}`, form)
    } else {
      await api.post('/domains', form)
    }
    showModal.value = false
    fetchDomains()
  } catch (err) {
    alert('Ошибка: ' + (err.response?.data?.error || err.message))
  }
}

const toggleDomainStatus = async (domain) => {
  try {
    const newStatus = !domain.active
    await api.put(`/domains/${domain.domain}`, {
      ...domain,
      active: newStatus
    })
    domain.active = newStatus
  } catch (err) {
    alert('Ошибка при смене статуса')
  }
}

const deleteDomain = async (name) => {
  if (!confirm(`Удалить домен ${name} и ВСЕ ящики / алиасы этого домена?\nВНИМАНИЕ: Это действие необратимо!`)) return
  try {
    await api.delete(`/domains/${name}`)
    fetchDomains()
  } catch (err) {
    alert('Ошибка при удалении домена')
  }
}

const formatDate = (dateStr) => {
  if (!dateStr || dateStr.startsWith('0001') || dateStr.startsWith('2000-01-01')) return '—'
  return new Date(dateStr).toLocaleDateString('ru-RU', { 
    day: '2-digit', 
    month: '2-digit', 
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
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
.text-gradient {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}
</style>
