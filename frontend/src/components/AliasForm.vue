<template>
  <div class="fixed inset-0 bg-slate-950/60 backdrop-blur-md flex items-center justify-center p-4 z-[100]">
    <div class="bg-white dark:bg-slate-900 rounded-[32px] w-full max-w-xl shadow-2xl border border-white/10 dark:border-slate-800/50 animate-in zoom-in duration-300 overflow-hidden">
      <!-- Header -->
      <div class="px-8 py-6 flex justify-between items-center bg-slate-50/50 dark:bg-slate-800/30 border-b border-slate-100 dark:border-slate-800/50">
        <div class="flex items-center gap-3">
          <div class="w-11 h-11 rounded-2xl bg-indigo-600 text-white flex items-center justify-center shadow-lg shadow-indigo-500/20">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" /></svg>
          </div>
          <div>
            <h3 class="text-xl font-extrabold text-slate-900 dark:text-white tracking-tight">
              {{ isEdit ? t('alias_form.title_edit') : (isDomainAlias ? t('alias_form.title_add_domain') : t('alias_form.title_add')) }}
            </h3>
            <p class="text-[10px] font-bold text-slate-400 uppercase tracking-widest">{{ isEdit ? form.address : t('alias_form.subtitle_add') }}</p>
          </div>
        </div>
        <button @click="$emit('close')" class="text-slate-400 hover:text-slate-600 dark:hover:text-white transition-all bg-white dark:bg-slate-800 p-2 rounded-xl shadow-sm border border-slate-100 dark:border-slate-700 hover:scale-110 active:scale-95">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M6 18L18 6M6 6l12 12" /></svg>
        </button>
      </div>

      <form @submit.prevent="save" class="p-8 space-y-6">
        <!-- Режим: Обычный / Доменный -->
        <div v-if="!isEdit" class="flex p-1 bg-slate-100 dark:bg-slate-800 rounded-2xl">
          <button type="button" @click="isDomainAlias = false" :class="!isDomainAlias ? 'bg-white dark:bg-slate-700 shadow-sm text-indigo-600' : 'text-slate-500'" class="flex-1 py-2 text-xs font-black uppercase tracking-widest rounded-xl transition-all">{{ t('alias_form.mode_regular') }}</button>
          <button type="button" @click="isDomainAlias = true" :class="isDomainAlias ? 'bg-white dark:bg-slate-700 shadow-sm text-amber-600' : 'text-slate-500'" class="flex-1 py-2 text-xs font-black uppercase tracking-widest rounded-xl transition-all">{{ t('alias_form.mode_domain') }}</button>
        </div>

        <!-- Источник -->
        <div v-if="!isDomainAlias" class="relative">
          <label class="block text-[10px] font-black uppercase tracking-widest text-slate-400 mb-1.5 ml-1">{{ t('alias_form.source_label') }}</label>
          <div class="flex items-center gap-3">
            <div class="relative flex-1 group">
              <input v-model="localPart" :disabled="isEdit" type="text" :placeholder="t('alias_form.source_placeholder')"
                class="w-full px-5 py-3.5 rounded-2xl border-2 border-slate-100 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/30 text-slate-900 dark:text-white focus:border-indigo-500 transition-all outline-none font-bold" />
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

        <div v-else class="space-y-4">
          <div class="relative">
            <label class="block text-[10px] font-black uppercase tracking-widest text-slate-400 mb-1.5 ml-1">{{ t('alias_form.source_domain_label') }}</label>
            <input v-model="form.alias_domain" type="text" placeholder="old-domain.com" required
              class="w-full px-5 py-3.5 rounded-2xl border-2 border-slate-100 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/30 text-slate-900 dark:text-white focus:border-amber-500 transition-all outline-none font-bold" />
          </div>

          <div class="relative">
            <label class="block text-[10px] font-black uppercase tracking-widest text-slate-400 mb-1.5 ml-1">{{ t('alias_form.target_domain_label') }}</label>
            <CustomSelect 
              v-model="form.target_domain" 
              :options="domains.map(d => ({ label: d.domain, value: d.domain }))"
            />
          </div>
          <p class="text-[9px] text-slate-400 ml-1 italic font-bold">
            {{ t('alias_form.domain_hint', { alias: form.alias_domain || '...', target: form.target_domain }) }}
          </p>
        </div>

        <!-- Глобальный алиас чекбокс -->
        <div v-if="!isDomainAlias && isSuperAdmin" class="flex flex-col bg-slate-50/30 dark:bg-slate-800/10 p-4 rounded-2xl border border-slate-100 dark:border-slate-800/50 gap-3">
          <div class="flex items-center justify-between">
            <div class="flex flex-col text-left">
              <span class="text-sm font-bold text-slate-700 dark:text-slate-300">{{ t('alias_form.all_mailboxes_label') }}</span>
              <span class="text-[10px] text-slate-400 uppercase font-black">{{ t('alias_form.all_mailboxes_hint') }}</span>
            </div>
            <button type="button" @click="toggleAllMailboxes" :class="allMailboxes ? 'bg-indigo-600' : 'bg-slate-300 dark:bg-slate-700'" class="relative inline-flex h-6 w-11 items-center rounded-full transition-colors active:scale-95">
              <span :class="allMailboxes ? 'translate-x-6' : 'translate-x-1'" class="inline-block h-4 w-4 transform rounded-full bg-white transition-transform" />
            </button>
          </div>
          <div v-if="allMailboxes" class="text-left border-t border-slate-100 dark:border-slate-800/50 pt-2">
            <button type="button" @click="showInstructions = true" class="text-[10px] font-bold text-indigo-500 hover:text-indigo-600 transition-colors uppercase tracking-widest flex items-center gap-1">
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
              {{ t('alias_form.all_mailboxes_instructions_btn') }}
            </button>
          </div>
        </div>

        <!-- Получатели -->
        <div v-if="!isDomainAlias" class="relative">
          <label class="block text-[10px] font-black uppercase tracking-widest text-slate-400 mb-1.5 ml-1">{{ t('alias_form.goto_label') }}</label>
          <textarea v-model="form.goto" :disabled="allMailboxes" :required="!allMailboxes" :placeholder="t('alias_form.goto_placeholder')" rows="4"
            class="w-full px-5 py-4 rounded-2xl border-2 border-slate-100 dark:border-slate-800 bg-white dark:bg-slate-900 text-slate-900 dark:text-white focus:border-indigo-500 transition-all outline-none font-bold resize-none leading-relaxed disabled:opacity-50 disabled:bg-slate-50 dark:disabled:bg-slate-900/50"></textarea>
          <p class="text-[9px] text-slate-400 mt-2 ml-1 font-bold tracking-tight uppercase opacity-70">{{ t('alias_form.goto_hint') }}</p>
        </div>

        <!-- Статус -->
        <div class="flex items-center justify-between bg-slate-50/30 dark:bg-slate-800/10 p-4 rounded-2xl border border-slate-100 dark:border-slate-800/50">
          <div class="flex flex-col">
            <span class="text-sm font-bold text-slate-700 dark:text-slate-300">{{ t('alias_form.status_label') }}</span>
            <span class="text-[10px] text-slate-400 uppercase font-black">{{ t('alias_form.status_hint') }}</span>
          </div>
          <button type="button" @click="form.active = !form.active" :class="form.active ? 'bg-green-500' : 'bg-slate-300 dark:bg-slate-700'" class="relative inline-flex h-6 w-11 items-center rounded-full transition-colors active:scale-95">
            <span :class="form.active ? 'translate-x-6' : 'translate-x-1'" class="inline-block h-4 w-4 transform rounded-full bg-white transition-transform" />
          </button>
        </div>

        <!-- Кнопки -->
        <div class="flex gap-4 pt-4">
          <button type="submit" :class="isDomainAlias ? 'bg-amber-600 hover:bg-amber-700 shadow-amber-500/30' : 'bg-indigo-600 hover:bg-indigo-700 shadow-indigo-500/30'" class="flex-1 py-4 text-white rounded-2xl font-black uppercase tracking-widest text-xs shadow-xl transition-all active:scale-95">
            {{ isEdit ? t('common.save') : t('common.create') }}
          </button>
          <button type="button" @click="$emit('close')" class="px-8 py-4 bg-slate-100 dark:bg-slate-800 text-slate-500 rounded-2xl font-black uppercase tracking-widest text-xs hover:bg-slate-200 dark:hover:bg-slate-700 transition-all">
            {{ t('common.cancel') }}
          </button>
        </div>
      </form>
    </div>

    <!-- Модальное окно инструкции по настройке Postfix -->
    <div v-if="showInstructions" class="fixed inset-0 bg-slate-950/80 backdrop-blur-md flex items-center justify-center p-4 z-[200] overflow-y-auto">
      <div class="bg-white dark:bg-slate-900 rounded-[32px] w-full max-w-2xl shadow-2xl border border-slate-200 dark:border-slate-800/50 max-h-[90vh] overflow-y-auto flex flex-col">
        <!-- Header -->
        <div class="px-8 py-6 flex justify-between items-center bg-slate-50/50 dark:bg-slate-800/30 border-b border-slate-100 dark:border-slate-800/50 sticky top-0 z-10 backdrop-blur-md">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-xl bg-indigo-600 text-white flex items-center justify-center shadow-lg shadow-indigo-500/20">
              <svg class="w-5.5 h-5.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" /></svg>
            </div>
            <div>
              <h4 class="text-lg font-extrabold text-slate-900 dark:text-white tracking-tight">
                {{ t('alias_form.all_mailboxes_instructions_title') }}
              </h4>
            </div>
          </div>
          <button type="button" @click="showInstructions = false" class="text-slate-400 hover:text-slate-600 dark:hover:text-white transition-all bg-white dark:bg-slate-800 p-2 rounded-xl shadow-sm border border-slate-100 dark:border-slate-700 hover:scale-110 active:scale-95">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M6 18L18 6M6 6l12 12" /></svg>
          </button>
        </div>

        <!-- Content -->
        <div class="p-8 space-y-6 overflow-y-auto text-left text-sm text-slate-600 dark:text-slate-400 font-medium">
          <!-- Step 1 -->
          <div class="space-y-2">
            <h5 class="font-bold text-slate-900 dark:text-white flex items-center gap-2">
              <span class="w-5 h-5 rounded-full bg-indigo-100 dark:bg-indigo-900/30 text-indigo-600 dark:text-indigo-400 text-xs flex items-center justify-center font-black">1</span>
              {{ t('alias_form.all_mailboxes_instructions_step1') }}
            </h5>
            <p>{{ t('alias_form.all_mailboxes_instructions_step1_desc') }}</p>
            <pre class="bg-slate-950 text-slate-200 p-4 rounded-xl text-xs overflow-x-auto font-mono select-all">query = SELECT goto FROM alias WHERE address='%s' AND active = '1' AND goto != '[ALL_MAILBOXES]'
        UNION 
        SELECT username FROM mailbox WHERE domain='%d' AND '%s'='all@%d' AND active='1'
        UNION
        SELECT username FROM mailbox WHERE active='1' AND EXISTS (
            SELECT 1 FROM alias WHERE address='%s' AND goto='[ALL_MAILBOXES]' AND active='1'
        )</pre>
          </div>

          <!-- Step 2 -->
          <div class="space-y-2">
            <h5 class="font-bold text-slate-900 dark:text-white flex items-center gap-2">
              <span class="w-5 h-5 rounded-full bg-indigo-100 dark:bg-indigo-900/30 text-indigo-600 dark:text-indigo-400 text-xs flex items-center justify-center font-black">2</span>
              {{ t('alias_form.all_mailboxes_instructions_step2') }}
            </h5>
            <p>{{ t('alias_form.all_mailboxes_instructions_step2_desc') }}</p>
            <pre class="bg-slate-950 text-slate-200 p-4 rounded-xl text-xs overflow-x-auto font-mono select-all">user = mailadmin
password = password_from_config
hosts = 127.0.0.1
dbname = postfix
query = SELECT 'REJECT access denied' FROM alias WHERE address='%s' AND goto='[ALL_MAILBOXES]' AND active='1'</pre>
          </div>

          <!-- Step 3 -->
          <div class="space-y-2">
            <h5 class="font-bold text-slate-900 dark:text-white flex items-center gap-2">
              <span class="w-5 h-5 rounded-full bg-indigo-100 dark:bg-indigo-900/30 text-indigo-600 dark:text-indigo-400 text-xs flex items-center justify-center font-black">3</span>
              {{ t('alias_form.all_mailboxes_instructions_step3') }}
            </h5>
            <p>{{ t('alias_form.all_mailboxes_instructions_step3_desc') }}</p>
            <pre class="bg-slate-950 text-slate-200 p-4 rounded-xl text-xs overflow-x-auto font-mono select-all">smtpd_recipient_restrictions =
    ...
    permit_sasl_authenticated,
    permit_mynetworks,
    # Блокировка внешних писем на любые глобальные рассылки
    check_recipient_access mysql:/etc/postfix/mysql_restrict_global_aliases.cf,
    reject_unauth_destination,
    ...</pre>
          </div>
        </div>

        <!-- Footer -->
        <div class="px-8 py-5 bg-slate-50/50 dark:bg-slate-800/30 border-t border-slate-100 dark:border-slate-800/50 flex justify-end">
          <button type="button" @click="showInstructions = false" class="px-6 py-2.5 bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-500 rounded-xl font-bold transition-all text-sm">
            {{ t('alias_form.all_mailboxes_instructions_close') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, reactive, computed } from 'vue'
import api from '@/api/axios'
import CustomSelect from '@/components/CustomSelect.vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/store/auth'

const { t } = useI18n()
const authStore = useAuthStore()
const isSuperAdmin = computed(() => authStore.isSuperAdmin)

const props = defineProps({
  domain: String,
  item: Object
})

const emit = defineEmits(['close', 'save'])

const isEdit = computed(() => !!props.item)
const isDomainAlias = ref(false)
const localPart = ref('')
const targetDomain = ref('')
const domains = ref([])
const allMailboxes = ref(false)
const prevGoto = ref('')
const showInstructions = ref(false)

const form = reactive({
  address: '',
  goto: '',
  active: true,
  domain: '',
  alias_domain: '',
  target_domain: ''
})

onMounted(async () => {
  // Загружаем список доменов
  try {
    const { data } = await api.get('/domains')
    domains.value = data
  } catch (e) {}

  if (props.item) {
    if (props.item.type === 'domain_alias') {
      isDomainAlias.value = true
      form.alias_domain = props.item.address
      form.target_domain = props.item.target_domain
      form.active = props.item.active
    } else {
      Object.assign(form, props.item)
      const parts = props.item.address.split('@')
      localPart.value = parts[0]
      targetDomain.value = parts[1]
      if (props.item.goto === '[ALL_MAILBOXES]') {
        allMailboxes.value = true
        form.goto = '[ALL_MAILBOXES]'
      } else {
        form.goto = props.item.goto.split(',').map(s => s.trim()).join('\n')
      }
    }
  } else {
    targetDomain.value = props.domain || (domains.value.length > 0 ? domains.value[0].domain : '')
  }
})

const toggleAllMailboxes = () => {
  allMailboxes.value = !allMailboxes.value
  if (allMailboxes.value) {
    prevGoto.value = form.goto
    form.goto = '[ALL_MAILBOXES]'
  } else {
    form.goto = prevGoto.value === '[ALL_MAILBOXES]' ? '' : prevGoto.value
  }
}

const save = async () => {
  try {
    if (isDomainAlias.value) {
      form.target_domain = props.domain
      if (isEdit.value) {
        // Мы не разрешаем менять имя домена-алиаса обычно, но если нужно - API должен поддерживать
        alert(t('alias_form.errors.domain_edit_not_supported'))
        return
      } else {
        await api.post('/aliases/domain-aliases', form)
      }
    } else {
      form.address = `${localPart.value}@${targetDomain.value}`
      form.domain = targetDomain.value
      // Очищаем и объединяем адреса обратно в запятые
      if (allMailboxes.value) {
        form.goto = '[ALL_MAILBOXES]'
      } else {
        form.goto = form.goto.split(/[\n,;]+/).map(s => s.trim()).filter(s => s).join(',')
      }
      
      if (isEdit.value) {
        await api.put(`/aliases/${form.address}`, form)
      } else {
        await api.post('/aliases', form)
      }
    }
    emit('save')
  } catch (err) {
    alert(t('common.error') + ': ' + (err.response?.data?.error || err.message))
  }
}
</script>
