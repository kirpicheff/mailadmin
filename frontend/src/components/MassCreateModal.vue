<template>
  <div class="fixed inset-0 bg-slate-950/60 backdrop-blur-xl flex items-center justify-center p-4 z-[70]">
    <div class="bg-white dark:bg-slate-900 rounded-[40px] w-full max-w-2xl shadow-2xl border border-white/10 dark:border-slate-800/50 animate-in zoom-in duration-300 overflow-hidden flex flex-col">
      
      <!-- Header -->
      <div class="px-10 py-8 flex justify-between items-center bg-slate-50/50 dark:bg-slate-800/30 border-b border-slate-100 dark:border-slate-800/50">
        <div class="flex items-center gap-4">
          <div class="w-14 h-14 rounded-3xl bg-slate-900 text-white flex items-center justify-center shadow-2xl shadow-slate-900/20">
            <svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 14v6m-3-3h6M6 10h2a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v2a2 2 0 002 2zm10 0h2a2 2 0 002-2V6a2 2 0 00-2-2h-2a2 2 0 00-2 2v2a2 2 0 002 2zM6 20h2a2 2 0 002-2v-2a2 2 0 00-2-2H6a2 2 0 00-2 2v2a2 2 0 002 2z" /></svg>
          </div>
          <div>
            <h3 class="text-2xl font-black text-slate-900 dark:text-white tracking-tight">{{ t('mailboxes.mass_modal.title') }}</h3>
            <p class="text-xs font-bold text-slate-400 uppercase tracking-widest mt-1">{{ t('mailboxes.mass_modal.subtitle', { domain }) }}</p>
          </div>
        </div>
        <button @click="$emit('close')" class="text-slate-400 hover:text-slate-600 dark:hover:text-white transition-all bg-white dark:bg-slate-800 p-3 rounded-2xl shadow-sm border border-slate-100 dark:border-slate-700 hover:scale-110 active:scale-95">
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M6 18L18 6M6 6l12 12" /></svg>
        </button>
      </div>

      <!-- Form -->
      <form @submit.prevent="save" class="p-10 space-y-8 overflow-y-auto max-h-[70vh]">
        
        <div class="space-y-3">
          <label class="block text-[10px] font-black uppercase tracking-widest text-slate-400 ml-1">{{ t('mailboxes.mass_modal.prefixes_label') }}</label>
          <textarea v-model="rawPrefixes" :placeholder="t('mailboxes.mass_modal.prefixes_hint')" required rows="5"
            class="w-full px-6 py-4 rounded-3xl border-2 border-slate-100 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/30 text-slate-900 dark:text-white focus:border-indigo-500 transition-all outline-none font-bold text-sm leading-relaxed"></textarea>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div class="space-y-3">
            <label class="block text-[10px] font-black uppercase tracking-widest text-slate-400 ml-1">{{ t('mailboxes.mass_modal.password_label') }}</label>
            <input v-model="form.password" type="text" required
              class="w-full px-6 py-4 rounded-2xl border-2 border-slate-100 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/30 text-slate-900 dark:text-white focus:border-indigo-500 transition-all outline-none font-bold text-sm" />
          </div>

          <div class="space-y-3">
            <label class="block text-[10px] font-black uppercase tracking-widest text-slate-400 ml-1">{{ t('mailboxes.mass_modal.quota_label') }}</label>
            <input v-model.number="form.quota" type="number" required
              class="w-full px-6 py-4 rounded-2xl border-2 border-slate-100 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/30 text-slate-900 dark:text-white focus:border-indigo-500 transition-all outline-none font-bold text-sm" />
          </div>
        </div>

        <div class="flex items-center justify-between p-6 bg-slate-50/50 dark:bg-slate-800/20 rounded-3xl border border-slate-100 dark:border-slate-800">
          <div class="flex flex-col gap-1">
            <span class="text-xs font-black uppercase tracking-widest text-slate-900 dark:text-white">{{ t('mailboxes.mass_modal.active_label') }}</span>
            <span class="text-[10px] text-slate-400 font-bold">Аккаунты будут готовы к работе сразу после создания</span>
          </div>
          <button type="button" @click="form.active = !form.active" :class="form.active ? 'bg-green-500' : 'bg-slate-200 dark:bg-slate-700'" class="relative inline-flex h-6 w-11 items-center rounded-full transition-all">
            <span :class="form.active ? 'translate-x-6' : 'translate-x-1'" class="inline-block h-4 w-4 transform rounded-full bg-white transition-transform" />
          </button>
        </div>

        <div class="flex gap-4 pt-4">
          <button type="submit" :disabled="loading" class="flex-1 py-5 bg-slate-900 text-white rounded-[24px] font-black uppercase tracking-widest text-xs hover:bg-slate-800 shadow-2xl shadow-slate-900/20 transition-all hover:-translate-y-1 active:translate-y-0 disabled:opacity-50">
            {{ loading ? t('common.loading') : t('mailboxes.mass_modal.button') }}
          </button>
          <button type="button" @click="$emit('close')" class="px-10 py-5 bg-slate-100 dark:bg-slate-800 text-slate-500 rounded-[24px] font-black uppercase tracking-widest text-xs hover:bg-slate-200 dark:hover:bg-slate-700 transition-all">
            {{ t('common.cancel') }}
          </button>
        </div>

      </form>

    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import api from '@/api/axios'
import { useI18n } from 'vue-i18n'

const props = defineProps({
  domain: String
})

const emit = defineEmits(['close', 'save'])
const { t } = useI18n()

const loading = ref(false)
const rawPrefixes = ref('')

const form = reactive({
  password: Math.random().toString(36).slice(-10),
  quota: 1024,
  active: true
})

const save = async () => {
  if (!rawPrefixes.value.trim()) return
  
  loading.value = true
  try {
    // Парсим префиксы
    const prefixes = rawPrefixes.value
      .split(/[\n,;]/)
      .map(s => s.trim())
      .filter(s => s)

    if (prefixes.length === 0) return

    const { data } = await api.post('/mailboxes/batch/create', {
      domain: props.domain,
      prefixes,
      password: form.password,
      quota: form.quota * 1024 * 1024,
      active: form.active
    })

    alert(t('mailboxes.mass_modal.success', { count: data.created }))
    emit('save')
  } catch (err) {
    alert(err.response?.data?.error || err.message)
  } finally {
    loading.value = false
  }
}
</script>
