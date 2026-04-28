<template>
  <div class="space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-700">
    <!-- Header -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div>
        <h1 class="text-4xl font-black text-slate-900 dark:text-white tracking-tight uppercase">
          {{ t('tools.title') }}
        </h1>
        <p class="text-slate-500 font-medium mt-1">{{ t('tools.send_mail.subtitle') }}</p>
      </div>
    </div>

    <!-- Main Content -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
      <!-- Individual Email Form -->
      <div class="glass-panel p-8 rounded-[32px] border border-slate-200 dark:border-slate-800 shadow-2xl space-y-6">
        <div class="flex items-center gap-4 mb-2">
          <div class="w-12 h-12 rounded-2xl bg-mail-blue-600 text-white flex items-center justify-center shadow-lg shadow-mail-blue-500/20">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" /></svg>
          </div>
          <h2 class="text-xl font-extrabold text-slate-900 dark:text-white">{{ t('tools.send_mail.title') }}</h2>
        </div>

        <form @submit.prevent="sendIndividual" class="space-y-4">
          <div class="space-y-1.5">
            <label class="text-[10px] font-black uppercase tracking-widest text-slate-400 ml-1">{{ t('tools.send_mail.from') }}</label>
            <input v-model="individual.from" type="email" :placeholder="authStore.user?.username" required
              class="w-full px-5 py-3.5 rounded-2xl border-2 border-slate-100 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/30 text-slate-900 dark:text-white focus:border-mail-blue-500 transition-all outline-none font-bold" />
          </div>

          <div class="space-y-1.5">
            <label class="text-[10px] font-black uppercase tracking-widest text-slate-400 ml-1">{{ t('tools.send_mail.to') }}</label>
            <input v-model="individual.to" type="email" required
              class="w-full px-5 py-3.5 rounded-2xl border-2 border-slate-100 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/30 text-slate-900 dark:text-white focus:border-mail-blue-500 transition-all outline-none font-bold" />
          </div>

          <div class="space-y-1.5">
            <label class="text-[10px] font-black uppercase tracking-widest text-slate-400 ml-1">{{ t('tools.send_mail.subject') }}</label>
            <input v-model="individual.subject" type="text" required
              class="w-full px-5 py-3.5 rounded-2xl border-2 border-slate-100 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/30 text-slate-900 dark:text-white focus:border-mail-blue-500 transition-all outline-none font-bold" />
          </div>

          <div class="space-y-1.5">
            <label class="text-[10px] font-black uppercase tracking-widest text-slate-400 ml-1">{{ t('tools.send_mail.body') }}</label>
            <textarea v-model="individual.body" rows="6" required
              class="w-full px-5 py-3.5 rounded-2xl border-2 border-slate-100 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/30 text-slate-900 dark:text-white focus:border-mail-blue-500 transition-all outline-none font-bold resize-none"></textarea>
          </div>

          <button type="submit" :disabled="loading" class="w-full py-4 bg-mail-blue-600 text-white rounded-2xl font-black uppercase tracking-widest text-xs hover:bg-mail-blue-700 shadow-xl shadow-mail-blue-500/30 transition-all active:scale-95 disabled:opacity-50">
            {{ loading ? t('common.loading') : t('tools.send_mail.send') }}
          </button>
        </form>
      </div>

      <!-- Broadcast Form -->
      <div class="glass-panel p-8 rounded-[32px] border border-slate-200 dark:border-slate-800 shadow-2xl space-y-6">
        <div class="flex items-center gap-4 mb-2">
          <div class="w-12 h-12 rounded-2xl bg-indigo-600 text-white flex items-center justify-center shadow-lg shadow-indigo-500/20">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5.882V19.24a1.76 1.76 0 01-3.417.592l-2.147-6.167M11 5.882c.443-.443 1.108-.592 1.69-.346l5.132 2.148c.582.246.963.81.963 1.441v5.66c0 .631-.381 1.195-.963 1.441l-5.132 2.148c-.582.246-1.247.097-1.69-.346M11 5.882V19.24" /></svg>
          </div>
          <h2 class="text-xl font-extrabold text-slate-900 dark:text-white">{{ t('tools.broadcast.title') }}</h2>
        </div>

        <form @submit.prevent="sendBroadcast" class="space-y-4">
          <div class="space-y-1.5">
            <label class="text-[10px] font-black uppercase tracking-widest text-slate-400 ml-1">{{ t('tools.send_mail.from') }}</label>
            <input v-model="broadcast.from" type="email" :placeholder="authStore.user?.username" required
              class="w-full px-5 py-3.5 rounded-2xl border-2 border-slate-100 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/30 text-slate-900 dark:text-white focus:border-mail-blue-500 transition-all outline-none font-bold" />
          </div>

          <div class="space-y-1.5">
            <label class="text-[10px] font-black uppercase tracking-widest text-slate-400 ml-1">{{ t('tools.broadcast.select_domains') }}</label>
            <div class="relative bg-slate-50/50 dark:bg-slate-800/30 rounded-2xl border-2 border-slate-100 dark:border-slate-800 p-2 max-h-32 overflow-y-auto">
              <label class="flex items-center gap-3 p-2 hover:bg-slate-100 dark:hover:bg-slate-700/50 rounded-xl cursor-pointer transition-colors">
                <input type="checkbox" v-model="allDomainsSelected" class="hidden" />
                <div class="w-5 h-5 rounded border-2 flex items-center justify-center transition-colors" :class="allDomainsSelected ? 'bg-indigo-600 border-indigo-600' : 'border-slate-300 dark:border-slate-600'">
                  <svg v-if="allDomainsSelected" class="w-3.5 h-3.5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7" /></svg>
                </div>
                <span class="text-xs font-bold">{{ t('tools.broadcast.all_domains') }}</span>
              </label>
              <template v-if="!allDomainsSelected">
                <label v-for="d in domains" :key="d.domain" class="flex items-center gap-3 p-2 hover:bg-slate-100 dark:hover:bg-slate-700/50 rounded-xl cursor-pointer transition-colors">
                  <input type="checkbox" :value="d.domain" v-model="broadcast.domains" class="hidden" />
                  <div class="w-5 h-5 rounded border-2 flex items-center justify-center transition-colors" :class="broadcast.domains.includes(d.domain) ? 'bg-indigo-600 border-indigo-600' : 'border-slate-300 dark:border-slate-600'">
                    <svg v-if="broadcast.domains.includes(d.domain)" class="w-3.5 h-3.5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7" /></svg>
                  </div>
                  <span class="text-xs font-bold">{{ d.domain }}</span>
                </label>
              </template>
            </div>
          </div>

          <div class="space-y-1.5">
            <label class="text-[10px] font-black uppercase tracking-widest text-slate-400 ml-1">{{ t('tools.send_mail.subject') }}</label>
            <input v-model="broadcast.subject" type="text" required
              class="w-full px-5 py-3.5 rounded-2xl border-2 border-slate-100 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/30 text-slate-900 dark:text-white focus:border-mail-blue-500 transition-all outline-none font-bold" />
          </div>

          <div class="space-y-1.5">
            <label class="text-[10px] font-black uppercase tracking-widest text-slate-400 ml-1">{{ t('tools.send_mail.body') }}</label>
            <textarea v-model="broadcast.body" rows="4" required
              class="w-full px-5 py-3.5 rounded-2xl border-2 border-slate-100 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/30 text-slate-900 dark:text-white focus:border-mail-blue-500 transition-all outline-none font-bold resize-none"></textarea>
          </div>

          <div class="flex items-center justify-between p-4 bg-slate-50/30 dark:bg-slate-800/10 rounded-2xl border border-slate-100 dark:border-slate-800/50">
            <span class="text-xs font-bold text-slate-700 dark:text-slate-300">{{ t('tools.broadcast.only_mailboxes') }}</span>
            <button type="button" @click="broadcast.only_mailboxes = !broadcast.only_mailboxes" :class="broadcast.only_mailboxes ? 'bg-indigo-600' : 'bg-slate-300 dark:bg-slate-700'" class="relative inline-flex h-5 w-9 items-center rounded-full transition-colors active:scale-95">
              <span :class="broadcast.only_mailboxes ? 'translate-x-5' : 'translate-x-1'" class="inline-block h-3.5 w-3.5 transform rounded-full bg-white transition-transform" />
            </button>
          </div>

          <button type="submit" :disabled="loadingBroadcast" class="w-full py-4 bg-indigo-600 text-white rounded-2xl font-black uppercase tracking-widest text-xs hover:bg-indigo-700 shadow-xl shadow-indigo-500/30 transition-all active:scale-95 disabled:opacity-50">
            {{ loadingBroadcast ? t('common.loading') : t('tools.broadcast.send') }}
          </button>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, watch } from 'vue'
import api from '@/api/axios'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/store/auth'

const { t, locale } = useI18n()
const authStore = useAuthStore()

const domains = ref([])
const loading = ref(false)
const loadingBroadcast = ref(false)
const allDomainsSelected = ref(true)

const individual = reactive({
  from: authStore.user?.username || '',
  to: '',
  subject: '',
  body: ''
})

const broadcast = reactive({
  from: authStore.user?.username || '',
  domains: [],
  subject: '',
  body: '',
  only_mailboxes: true
})

onMounted(async () => {
  // Инициализация дефолтных значений
  initDefaults()

  try {
    const { data } = await api.get('/domains')
    domains.value = data
  } catch (e) {}
})

const initDefaults = () => {
  individual.subject = t('tools.send_mail.default_subject')
  individual.body = t('tools.send_mail.default_body')
  broadcast.subject = t('tools.broadcast.default_subject')
  broadcast.body = t('tools.broadcast.default_body')
}

// Следим за сменой языка, чтобы обновить дефолтные тексты, если форма пуста или не тронута
watch(locale, () => {
  initDefaults()
})

const sendIndividual = async () => {
  loading.value = true
  try {
    await api.post('/tools/send-email', individual)
    alert(t('tools.send_mail.success'))
    individual.to = ''
    // Не очищаем тему и тело, если пользователь хочет отправить еще одно похожее
  } catch (err) {
    alert(t('common.error') + ': ' + (err.response?.data?.error || err.message))
  } finally {
    loading.value = false
  }
}

const sendBroadcast = async () => {
  loadingBroadcast.value = true
  try {
    const data = { ...broadcast }
    if (allDomainsSelected.value) {
      data.domains = [] 
    }
    
    await api.post('/tools/broadcast', data)
    alert(t('tools.broadcast.success'))
  } catch (err) {
    alert(t('common.error') + ': ' + (err.response?.data?.error || err.message))
  } finally {
    loadingBroadcast.value = false
  }
}

watch(allDomainsSelected, (val) => {
  if (val) broadcast.domains = []
})
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
