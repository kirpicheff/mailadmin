<template>
  <div class="fixed inset-0 bg-slate-950/60 backdrop-blur-md flex items-center justify-center p-4 z-[100]">
    <div class="bg-white dark:bg-slate-900 rounded-[32px] w-full max-w-xl shadow-2xl border border-white/10 dark:border-slate-800/50 animate-in zoom-in duration-300 overflow-hidden flex flex-col max-h-[90vh]">
      <!-- Header -->
      <div class="px-8 py-6 flex justify-between items-center bg-slate-50/50 dark:bg-slate-800/30 border-b border-slate-100 dark:border-slate-800/50 flex-shrink-0">
        <div class="flex items-center gap-3">
          <div class="w-11 h-11 rounded-2xl bg-mail-blue-600 text-white flex items-center justify-center shadow-lg shadow-mail-blue-500/20">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" /></svg>
          </div>
          <div>
            <h3 class="text-xl font-extrabold text-slate-900 dark:text-white tracking-tight">
              {{ isEdit ? t('mailbox_form.title_edit') : t('mailbox_form.title_add') }}
            </h3>
            <p class="text-[10px] font-bold text-slate-400 uppercase tracking-widest">{{ isEdit ? form.username : t('mailbox_form.subtitle_add') }}</p>
          </div>
        </div>
        <button @click="$emit('close')" class="text-slate-400 hover:text-slate-600 dark:hover:text-white transition-all bg-white dark:bg-slate-800 p-2 rounded-xl shadow-sm border border-slate-100 dark:border-slate-700 hover:scale-110 active:scale-95">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M6 18L18 6M6 6l12 12" /></svg>
        </button>
      </div>

      <!-- Content -->
      <div class="overflow-y-auto flex-1 p-8">
        <form @submit.prevent="save" class="space-y-6">
          <div class="space-y-6">
            <!-- Адрес -->
            <div class="relative">
              <label class="block text-[10px] font-black uppercase tracking-widest text-slate-400 mb-1.5 ml-1">{{ t('mailbox_form.address_label') }}</label>
              <div class="flex items-center gap-3">
                <div class="relative flex-1 group">
                  <input v-model="localPart" :disabled="isEdit" type="text" :placeholder="t('mailbox_form.address_placeholder')" required
                    class="w-full px-5 py-3.5 rounded-2xl border-2 border-slate-100 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/30 text-slate-900 dark:text-white focus:border-mail-blue-500 transition-all outline-none font-bold" />
                </div>
                <div class="relative w-48">
                  <CustomSelect 
                    v-model="targetDomain" 
                    :disabled="isEdit"
                    :options="domains.map(d => ({ label: '@' + d.domain, value: d.domain }))"
                  />
                </div>
              </div>
            </div>

            <!-- Имя -->
            <div class="relative">
              <label class="block text-[10px] font-black uppercase tracking-widest text-slate-400 mb-1.5 ml-1">{{ t('mailbox_form.name_label') }}</label>
              <input v-model="form.name" type="text" :placeholder="t('mailbox_form.name_placeholder')"
                class="w-full px-5 py-3.5 rounded-2xl border-2 border-slate-100 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/30 text-slate-900 dark:text-white focus:border-mail-blue-500 transition-all outline-none font-bold" />
            </div>

            <!-- Пароль -->
            <div class="relative">
              <label class="block text-[10px] font-black uppercase tracking-widest text-slate-400 mb-1.5 ml-1">{{ t('mailbox_form.password_label') }}</label>
              <div class="flex gap-2">
                <div class="relative flex-1 group">
                  <input v-model="form.password" :type="showPassword ? 'text' : 'password'" :placeholder="isEdit ? t('mailbox_form.password_placeholder_edit') : t('mailbox_form.password_placeholder_add')" :required="!isEdit"
                    class="w-full px-5 py-3.5 rounded-2xl border-2 border-slate-100 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/30 text-slate-900 dark:text-white focus:border-mail-blue-500 transition-all outline-none font-bold" />
                  <button type="button" @click="showPassword = !showPassword" class="absolute right-4 top-3.5 text-slate-400 hover:text-mail-blue-500 transition-colors">
                    <svg v-if="showPassword" class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" /></svg>
                    <svg v-else class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.542-7a10.025 10.025 0 014.132-5.411m0 0L4 4m5.352 5.352a3 3 0 004.293 4.293M11.25 11.25l.041-.041m-2.141 2.141L12 12m4.242 4.242L18 18" /></svg>
                  </button>
                </div>
                <button type="button" @click="generatePassword" class="px-4 bg-slate-100 dark:bg-slate-800 text-slate-600 rounded-2xl hover:bg-slate-200 transition-all border border-slate-200 dark:border-slate-700 active:scale-95 shadow-sm" :title="t('mailbox_form.password_generate')">
                  <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357-2H15" /></svg>
                </button>
              </div>
              <PasswordStrength :password="form.password" />
            </div>

            <!-- Квота -->
            <div class="relative">
              <label class="block text-[10px] font-black uppercase tracking-widest text-slate-400 mb-1.5 ml-1">{{ t('mailbox_form.quota_label') }}</label>
              <div class="flex items-center gap-4 bg-slate-50/50 dark:bg-slate-800/30 p-4 rounded-2xl border-2 border-slate-100 dark:border-slate-800">
                <input v-model.number="quotaMb" type="number" 
                  class="w-24 px-4 py-2 rounded-xl border-2 border-white dark:border-slate-900 bg-white dark:bg-slate-900 font-bold outline-none focus:border-mail-blue-500 text-center" />
                <div class="flex-1">
                  <input type="range" v-model.number="quotaMb" min="10" max="51200" step="100" class="w-full accent-mail-blue-600 h-1.5 bg-slate-200 dark:bg-slate-700 rounded-full appearance-none cursor-pointer" />
                </div>
              </div>
              <p class="text-[9px] text-slate-400 mt-1.5 ml-1 uppercase font-bold tracking-tighter opacity-70 italic">{{ t('mailbox_form.quota_hint') }}</p>
            </div>

            <!-- Статус -->
            <div class="flex items-center justify-between bg-slate-50/30 dark:bg-slate-800/10 p-4 rounded-2xl border border-slate-100 dark:border-slate-800/50">
              <div class="flex flex-col">
                <span class="text-sm font-bold text-slate-700 dark:text-slate-300">{{ t('mailbox_form.status_label') }}</span>
                <span class="text-[10px] text-slate-400 uppercase font-black">{{ t('mailbox_form.status_hint') }}</span>
              </div>
              <button type="button" @click="form.active = !form.active" :class="form.active ? 'bg-green-500' : 'bg-slate-300 dark:bg-slate-700'" class="relative inline-flex h-6 w-11 items-center rounded-full transition-colors active:scale-95">
                <span :class="form.active ? 'translate-x-6' : 'translate-x-1'" class="inline-block h-4 w-4 transform rounded-full bg-white transition-transform" />
              </button>
            </div>
          </div>

          <!-- Buttons -->
          <div class="flex gap-4 pt-4 sticky bottom-0 bg-white dark:bg-slate-900 pb-2">
            <button type="submit" :disabled="loading" class="flex-1 py-4 bg-mail-blue-600 text-white rounded-2xl font-black uppercase tracking-widest text-xs hover:bg-mail-blue-700 shadow-xl shadow-mail-blue-500/30 transition-all active:scale-95 disabled:opacity-50">
              {{ loading ? t('common.loading') : (isEdit ? t('common.save') : t('common.create')) }}
            </button>
            <button type="button" @click="$emit('close')" class="px-8 py-4 bg-slate-100 dark:bg-slate-800 text-slate-500 rounded-2xl font-black uppercase tracking-widest text-xs hover:bg-slate-200 dark:hover:bg-slate-700 transition-all">
              {{ t('common.cancel') }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, reactive, computed, watch } from 'vue'
import api from '@/api/axios'
import CustomSelect from '@/components/CustomSelect.vue'
import PasswordStrength from '@/components/PasswordStrength.vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const props = defineProps({
  domain: String,
  item: Object
})

const emit = defineEmits(['close', 'save'])

const isEdit = computed(() => !!props.item)
const localPart = ref('')
const targetDomain = ref('')
const domains = ref([])
const quotaMb = ref(0)
const showPassword = ref(false)
const loading = ref(false)

const form = reactive({
  username: '',
  password: '',
  name: '',
  quota: 0,
  active: true
})

onMounted(async () => {
  // Загружаем список доменов для выбора
  try {
    const { data } = await api.get('/domains')
    domains.value = data
  } catch (e) {}

  if (props.item) {
    Object.assign(form, props.item)
    form.password = ''
    const parts = props.item.username.split('@')
    localPart.value = parts[0]
    targetDomain.value = parts[1]
    quotaMb.value = props.item.quota / (1024 * 1024)
  } else {
    targetDomain.value = props.domain || (domains.value.length > 0 ? domains.value[0].domain : '')
  }
})

const generatePassword = () => {
  const charset = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghjkmnpqrstuvwxyz34679#$@!%&*"
  let res = ""
  const len = 12
  for (let i = 0; i < len; i++) {
    res += charset.charAt(Math.floor(Math.random() * charset.length))
  }
  form.password = res
  showPassword.value = true
}

const save = async () => {
  loading.value = true
  try {
    form.username = `${localPart.value}@${targetDomain.value}`
    form.quota = quotaMb.value * 1024 * 1024

    if (isEdit.value) {
      await api.put(`/mailboxes/${form.username}`, {
        password: form.password,
        name: form.name,
        quota: form.quota,
        active: form.active
      })
    } else {
      await api.post('/mailboxes', form)
    }

    emit('save')
  } catch (err) {
    alert(t('common.error') + ': ' + (err.response?.data?.error || err.message))
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
input[type="range"]::-webkit-slider-thumb {
  appearance: none;
  width: 16px;
  height: 16px;
  background: white;
  border: 4px solid #2563eb;
  border-radius: 50%;
  cursor: pointer;
  box-shadow: 0 0 10px rgba(37, 99, 235, 0.2);
}
</style>
