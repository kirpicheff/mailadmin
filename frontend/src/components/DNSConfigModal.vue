<script setup>
import { ref, watch } from 'vue'
import api from '@/api/axios'
import { useI18n } from 'vue-i18n'

const props = defineProps({
  show: Boolean,
  domain: String
})

const emit = defineEmits(['close'])
const { t } = useI18n()

const loading = ref(false)
const config = ref(null)

const fetchConfig = async () => {
  if (!props.domain) return
  loading.value = true
  try {
    const { data } = await api.get(`/domains/${props.domain}/dns`)
    config.value = data
  } catch (err) {
    console.error('Failed to fetch DNS config:', err)
  } finally {
    loading.value = false
  }
}

const copyToClipboard = (text) => {
  navigator.clipboard.writeText(text)
  // Можно добавить тост-уведомление здесь
}

watch(() => props.show, (newVal) => {
  if (newVal) fetchConfig()
})
</script>

<template>
  <div v-if="show" class="fixed inset-0 bg-slate-950/60 backdrop-blur-xl flex items-center justify-center p-4 z-[70]">
    <div class="bg-white dark:bg-slate-900 rounded-[32px] w-full max-w-2xl shadow-2xl border border-white/10 dark:border-slate-800/50 animate-in zoom-in duration-300 overflow-hidden flex flex-col">
      
      <!-- Header -->
      <div class="px-8 py-6 flex justify-between items-center bg-slate-50/50 dark:bg-slate-800/30 border-b border-slate-100 dark:border-slate-800/50">
        <div class="flex items-center gap-3">
          <div class="w-11 h-11 rounded-2xl bg-amber-500 text-white flex items-center justify-center shadow-lg shadow-amber-500/20">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>
          </div>
          <div>
            <h3 class="text-xl font-extrabold text-slate-900 dark:text-white tracking-tight">{{ t('dns.title') }}</h3>
            <p class="text-[10px] font-bold text-slate-400 uppercase tracking-widest">{{ domain }}</p>
          </div>
        </div>
        <button @click="$emit('close')" class="text-slate-400 hover:text-slate-600 dark:hover:text-white transition-all bg-white dark:bg-slate-800 p-2 rounded-xl shadow-sm border border-slate-100 dark:border-slate-700">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M6 18L18 6M6 6l12 12" /></svg>
        </button>
      </div>

      <div class="p-8 overflow-y-auto space-y-6">
        <div v-if="loading" class="flex flex-col items-center py-10">
          <div class="w-10 h-10 rounded-full border-4 border-slate-100 dark:border-slate-800 border-t-amber-500 animate-spin mb-4"></div>
        </div>

        <div v-else-if="config" class="space-y-6 animate-in fade-in duration-500">
          <!-- SPF -->
          <div class="space-y-2">
            <div class="flex justify-between items-center">
              <label class="text-[10px] font-black uppercase tracking-widest text-slate-400">{{ t('dns.spf') }}</label>
              <span class="text-[10px] font-bold text-slate-400">{{ t('dns.host') }}: @</span>
            </div>
            <div class="relative group">
              <div class="p-4 bg-slate-50 dark:bg-slate-950/50 rounded-2xl border border-slate-100 dark:border-slate-800 font-mono text-xs text-slate-600 dark:text-slate-300 break-all pr-12">
                {{ config.spf }}
              </div>
              <button @click="copyToClipboard(config.spf)" class="absolute right-3 top-1/2 -translate-y-1/2 p-2 bg-white dark:bg-slate-800 rounded-lg shadow-sm border border-slate-100 dark:border-slate-700 opacity-0 group-hover:opacity-100 transition-opacity">
                <svg class="w-4 h-4 text-amber-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3" /></svg>
              </button>
            </div>
          </div>

          <!-- DKIM -->
          <div class="space-y-2">
            <div class="flex justify-between items-center">
              <label class="text-[10px] font-black uppercase tracking-widest text-slate-400">{{ t('dns.dkim') }}</label>
              <span class="text-[10px] font-bold text-slate-400">{{ t('dns.host') }}: default._domainkey</span>
            </div>
            <div v-if="config.dkim" class="relative group">
              <div class="p-4 bg-slate-50 dark:bg-slate-950/50 rounded-2xl border border-slate-100 dark:border-slate-800 font-mono text-xs text-slate-600 dark:text-slate-300 break-all pr-12 max-h-32 overflow-y-auto">
                {{ config.dkim }}
              </div>
              <button @click="copyToClipboard(config.dkim)" class="absolute right-3 top-4 p-2 bg-white dark:bg-slate-800 rounded-lg shadow-sm border border-slate-100 dark:border-slate-700 opacity-0 group-hover:opacity-100 transition-opacity">
                <svg class="w-4 h-4 text-amber-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3" /></svg>
              </button>
            </div>
            <div v-else class="p-6 bg-amber-50 dark:bg-amber-900/10 rounded-2xl border border-amber-100 dark:border-amber-900/20 text-center">
              <p class="text-xs font-bold text-amber-700 dark:text-amber-400">{{ t('dns.dkim_not_found') }}</p>
              <p class="text-[10px] text-amber-600/70 mt-1 uppercase tracking-tight">{{ t('dns.dkim_path') }}: /etc/opendkim/keys/{{ domain }}/default.txt</p>
            </div>
          </div>

          <!-- DMARC -->
          <div class="space-y-2">
            <div class="flex justify-between items-center">
              <label class="text-[10px] font-black uppercase tracking-widest text-slate-400">{{ t('dns.dmarc') }}</label>
              <span class="text-[10px] font-bold text-slate-400">{{ t('dns.host') }}: _dmarc</span>
            </div>
            <div class="relative group">
              <div class="p-4 bg-slate-50 dark:bg-slate-950/50 rounded-2xl border border-slate-100 dark:border-slate-800 font-mono text-xs text-slate-600 dark:text-slate-300 break-all pr-12">
                {{ config.dmarc }}
              </div>
              <button @click="copyToClipboard(config.dmarc)" class="absolute right-3 top-1/2 -translate-y-1/2 p-2 bg-white dark:bg-slate-800 rounded-lg shadow-sm border border-slate-100 dark:border-slate-700 opacity-0 group-hover:opacity-100 transition-opacity">
                <svg class="w-4 h-4 text-amber-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3" /></svg>
              </button>
            </div>
          </div>

          <div class="p-4 bg-slate-50 dark:bg-slate-800/50 rounded-2xl flex items-start gap-3">
            <svg class="w-5 h-5 text-slate-400 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
            <p class="text-[10px] text-slate-500 leading-relaxed font-medium">{{ t('dns.hint') }}</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
