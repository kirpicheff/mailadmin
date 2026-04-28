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
              <input v-model="localPart" :disabled="isEdit" type="text" placeholder="sales" required
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

        <!-- Получатели -->
        <div v-if="!isDomainAlias" class="relative">
          <label class="block text-[10px] font-black uppercase tracking-widest text-slate-400 mb-1.5 ml-1">{{ t('alias_form.goto_label') }}</label>
          <textarea v-model="form.goto" :placeholder="t('alias_form.goto_placeholder')" required rows="4"
            class="w-full px-5 py-4 rounded-2xl border-2 border-slate-100 dark:border-slate-800 bg-white dark:bg-slate-900 text-slate-900 dark:text-white focus:border-indigo-500 transition-all outline-none font-bold resize-none leading-relaxed"></textarea>
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
  </div>
</template>

<script setup>
import { ref, onMounted, reactive, computed } from 'vue'
import api from '@/api/axios'
import CustomSelect from '@/components/CustomSelect.vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

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

const form = reactive({
  address: '',
  goto: '',
  active: true,
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
      form.goto = props.item.goto.split(',').map(s => s.trim()).join('\n')
    }
  } else {
    targetDomain.value = props.domain || (domains.value.length > 0 ? domains.value[0].domain : '')
  }
})

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
      // Очищаем и объединяем адреса обратно в запятые
      form.goto = form.goto.split(/[\n,;]+/).map(s => s.trim()).filter(s => s).join(',')
      
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
