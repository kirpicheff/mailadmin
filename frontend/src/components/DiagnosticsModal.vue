<template>
  <div v-if="show" class="fixed inset-0 bg-slate-950/60 backdrop-blur-xl flex items-center justify-center p-4 z-[60]">
    <div class="bg-white dark:bg-slate-900 rounded-[32px] w-full max-w-3xl shadow-2xl border border-white/10 dark:border-slate-800/50 animate-in zoom-in duration-300 overflow-hidden max-h-[90vh] flex flex-col">
      
      <!-- Header -->
      <div class="px-8 py-6 flex justify-between items-center bg-slate-50/50 dark:bg-slate-800/30 border-b border-slate-100 dark:border-slate-800/50">
        <div class="flex items-center gap-3">
          <div class="w-11 h-11 rounded-2xl bg-indigo-600 text-white flex items-center justify-center shadow-lg shadow-indigo-500/20">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" /></svg>
          </div>
          <div>
            <h3 class="text-xl font-extrabold text-slate-900 dark:text-white tracking-tight">{{ t('diagnostics.title') }}</h3>
            <p class="text-[10px] font-bold text-slate-400 uppercase tracking-widest">{{ domain }}</p>
          </div>
        </div>
        <button @click="$emit('close')" class="text-slate-400 hover:text-slate-600 dark:hover:text-white transition-all bg-white dark:bg-slate-800 p-2 rounded-xl shadow-sm border border-slate-100 dark:border-slate-700 hover:scale-110 active:scale-95">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M6 18L18 6M6 6l12 12" /></svg>
        </button>
      </div>

      <!-- Content -->
      <div class="p-8 overflow-y-auto space-y-8">
        <div v-if="loading" class="flex flex-col items-center py-20">
          <div class="w-12 h-12 rounded-full border-4 border-slate-100 dark:border-slate-800 border-t-indigo-500 animate-spin mb-4"></div>
          <p class="text-xs font-bold text-slate-400 uppercase tracking-widest">{{ t('diagnostics.checking') }}</p>
        </div>

        <div v-else-if="result" class="space-y-8 animate-in fade-in duration-500">
          
          <!-- DNS Section -->
          <div class="space-y-4">
            <h4 class="text-[10px] font-black uppercase tracking-[0.2em] text-slate-400 flex items-center gap-2">
              <span class="w-1.5 h-1.5 rounded-full bg-indigo-500"></span>
              {{ t('diagnostics.dns_records') }}
            </h4>
            
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <!-- MX -->
              <div class="p-5 rounded-3xl bg-slate-50/50 dark:bg-slate-800/30 border border-slate-100 dark:border-slate-700/50">
                <div class="flex items-center justify-between mb-3">
                  <span class="text-xs font-bold text-slate-800 dark:text-slate-200">{{ t('diagnostics.mx.title') }}</span>
                  <div v-if="result.mx?.length" class="px-2 py-0.5 bg-green-500/10 text-green-600 dark:text-green-400 text-[10px] font-black rounded-lg uppercase">OK</div>
                  <div v-else class="px-2 py-0.5 bg-red-500/10 text-red-600 dark:text-red-400 text-[10px] font-black rounded-lg uppercase">MISSING</div>
                </div>
                <div class="space-y-1">
                  <div v-for="mx in result.mx" :key="mx" class="text-[11px] font-medium text-slate-500 truncate">{{ mx }}</div>
                  <div v-if="!result.mx?.length" class="text-[11px] text-slate-400 italic">No records found</div>
                </div>
              </div>

              <!-- SPF -->
              <div class="p-5 rounded-3xl bg-slate-50/50 dark:bg-slate-800/30 border border-slate-100 dark:border-slate-700/50">
                <div class="flex items-center justify-between mb-3">
                  <span class="text-xs font-bold text-slate-800 dark:text-slate-200">{{ t('diagnostics.spf.title') }}</span>
                  <div v-if="result.spf.valid" class="px-2 py-0.5 bg-green-500/10 text-green-600 dark:text-green-400 text-[10px] font-black rounded-lg uppercase">OK</div>
                  <div v-else class="px-2 py-0.5 bg-amber-500/10 text-amber-600 dark:text-amber-400 text-[10px] font-black rounded-lg uppercase">WARN</div>
                </div>
                <p class="text-[11px] font-medium text-slate-500 flex-1 break-all line-clamp-2" :title="result.spf.value">{{ result.spf.value || t('diagnostics.spf.invalid') }}</p>
              </div>

              <!-- DMARC -->
              <div class="p-5 rounded-3xl bg-slate-50/50 dark:bg-slate-800/30 border border-slate-100 dark:border-slate-700/50">
                <div class="flex items-center justify-between mb-3">
                  <span class="text-xs font-bold text-slate-800 dark:text-slate-200">{{ t('diagnostics.dmarc.title') }}</span>
                  <div v-if="result.dmarc.valid" class="px-2 py-0.5 bg-green-500/10 text-green-600 dark:text-green-400 text-[10px] font-black rounded-lg uppercase">OK</div>
                  <div v-else class="px-2 py-0.5 bg-amber-500/10 text-amber-600 dark:text-amber-400 text-[10px] font-black rounded-lg uppercase">WARN</div>
                </div>
                <p class="text-[11px] font-medium text-slate-500 break-all">{{ result.dmarc.value || t('diagnostics.dmarc.invalid') }}</p>
              </div>

              <!-- DKIM -->
              <div class="p-5 rounded-3xl bg-slate-50/50 dark:bg-slate-800/30 border border-slate-100 dark:border-slate-700/50">
                <div class="flex items-center justify-between mb-3">
                  <span class="text-xs font-bold text-slate-800 dark:text-slate-200">{{ t('diagnostics.dkim.title') }}</span>
                  <div v-if="result.dkim?.length" class="px-2 py-0.5 bg-green-500/10 text-green-600 dark:text-green-400 text-[10px] font-black rounded-lg uppercase">FOUND</div>
                  <div v-else class="px-2 py-0.5 bg-slate-500/10 text-slate-500 text-[10px] font-black rounded-lg uppercase">N/A</div>
                </div>
                <div class="space-y-1">
                  <div v-for="dkim in result.dkim" :key="dkim.selector" class="flex items-center justify-between">
                    <span class="text-[10px] font-bold text-slate-400 uppercase tracking-tight">{{ dkim.selector }}</span>
                    <span class="text-[10px] text-green-500 font-bold">READY</span>
                  </div>
                  <div v-if="!result.dkim?.length" class="text-[11px] text-slate-400 italic">{{ t('diagnostics.dkim.not_found') }}</div>
                </div>
              </div>
            </div>
          </div>

          <!-- SSL Section -->
          <div class="space-y-4">
            <h4 class="text-[10px] font-black uppercase tracking-[0.2em] text-slate-400 flex items-center gap-2">
              <span class="w-1.5 h-1.5 rounded-full bg-cyan-500"></span>
              {{ t('diagnostics.ssl_status') }}
            </h4>
            <div class="grid grid-cols-1 gap-3">
              <div v-for="ssl in result.ssl" :key="ssl.host + ssl.port" class="flex items-center justify-between p-4 bg-slate-50/50 dark:bg-slate-800/30 rounded-2xl border border-slate-100 dark:border-slate-700/50">
                <div class="flex items-center gap-4">
                  <div :class="ssl.valid ? 'bg-cyan-500/10 text-cyan-600' : 'bg-red-500/10 text-red-600'" class="w-10 h-10 rounded-xl flex items-center justify-center">
                    <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" /></svg>
                  </div>
                  <div>
                    <p class="text-xs font-bold text-slate-800 dark:text-slate-200">{{ ssl.host }}:{{ ssl.port }}</p>
                    <p class="text-[10px] text-slate-400 font-medium">{{ ssl.issuer || 'Unknown issuer' }}</p>
                  </div>
                </div>
                <div class="text-right">
                  <p :class="ssl.valid ? 'text-green-600 dark:text-green-400' : 'text-red-600'" class="text-[10px] font-black uppercase">{{ ssl.valid ? t('diagnostics.ssl.valid') : t('diagnostics.ssl.expired') }}</p>
                  <p v-if="ssl.valid" class="text-[10px] text-slate-400 mt-1">{{ ssl.days_left }} {{ t('common.days_short') }} {{ t('diagnostics.ssl.expires') }}</p>
                </div>
              </div>
            </div>
          </div>

          <!-- RBL Section -->
          <div class="space-y-4">
            <h4 class="text-[10px] font-black uppercase tracking-[0.2em] text-slate-400 flex items-center gap-2">
              <span class="w-1.5 h-1.5 rounded-full bg-red-500"></span>
              {{ t('diagnostics.rbl_status') }}
            </h4>
            <div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
              <div v-for="rbl in result.rbl" :key="rbl.server" :class="rbl.listed ? 'border-red-200 bg-red-50/50' : 'border-slate-100 bg-slate-50/50'" class="p-4 rounded-2xl border dark:bg-slate-800/30 dark:border-slate-700/50 flex flex-col items-center gap-2 text-center">
                <span class="text-[10px] font-bold text-slate-400 truncate w-full">{{ rbl.server }}</span>
                <span :class="rbl.listed ? 'text-red-600' : 'text-green-600'" class="text-xs font-black uppercase tracking-tight">{{ rbl.listed ? t('diagnostics.rbl.listed') : t('diagnostics.rbl.clean') }}</span>
              </div>
            </div>
          </div>

        </div>
      </div>

    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import api from '@/api/axios'
import { useI18n } from 'vue-i18n'

const props = defineProps({
  show: Boolean,
  domain: String
})

const emit = defineEmits(['close'])
const { t } = useI18n()

const loading = ref(false)
const result = ref(null)

const runDiagnostics = async () => {
  if (!props.domain) return
  loading.value = true
  try {
    const { data } = await api.get(`/diagnostics/${props.domain}`)
    result.value = data
  } catch (err) {
    console.error('Diagnostics failed:', err)
  } finally {
    loading.value = false
  }
}

watch(() => props.show, (newVal) => {
  if (newVal) {
    runDiagnostics()
  } else {
    result.value = null
  }
})
</script>
