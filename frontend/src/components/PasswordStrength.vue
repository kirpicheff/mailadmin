<template>
  <div class="mt-3 space-y-2">
    <!-- Метка и название уровня сложности -->
    <div class="flex justify-between items-center text-[10px] font-black uppercase tracking-widest">
      <span class="text-slate-400">{{ t('password_strength.title') }}</span>
      <span :class="strengthColorText" class="transition-colors duration-300 font-extrabold">
        {{ strengthLabel }}
      </span>
    </div>

    <!-- Индикаторные полосы (шкала сложности) -->
    <div class="grid grid-cols-4 gap-1.5 h-1.5 w-full">
      <div 
        v-for="index in 4" 
        :key="index"
        :class="getSegmentClass(index)"
        class="h-full rounded-full transition-all duration-500 ease-out"
      />
    </div>

    <!-- Чек-лист критериев надежности пароля -->
    <div class="pt-1.5 grid grid-cols-1 sm:grid-cols-2 gap-x-4 gap-y-1.5 text-[10px] font-bold text-slate-400 uppercase tracking-tight">
      <div 
        v-for="(met, key) in criteria" 
        :key="key"
        class="flex items-center gap-1.5 transition-all duration-300"
        :class="met ? 'text-emerald-500 dark:text-emerald-400' : 'text-slate-400 dark:text-slate-600'"
      >
        <span class="flex-shrink-0 w-3.5 h-3.5 rounded-full flex items-center justify-center border transition-all"
          :class="met ? 'border-emerald-500 bg-emerald-500/10' : 'border-slate-300 dark:border-slate-700'"
        >
          <svg v-if="met" class="w-2 h-2 text-emerald-500 dark:text-emerald-400" fill="none" stroke="currentColor" stroke-width="3" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" d="M4.5 12.75l6 6 9-13.5" />
          </svg>
          <span v-else class="w-1 h-1 rounded-full bg-slate-300 dark:bg-slate-700" />
        </span>
        <span>{{ t(`password_strength.criteria.${key}`) }}</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const props = defineProps({
  password: {
    type: String,
    default: ''
  }
})

// Проверка критериев пароля
const criteria = computed(() => {
  const p = props.password || ''
  return {
    length: p.length >= 8,
    lowercase: /[a-z]/.test(p),
    uppercase: /[A-Z]/.test(p),
    number: /[0-9]/.test(p),
    special: /[^A-Za-z0-9]/.test(p)
  }
})

// Подсчет пройденных проверок
const score = computed(() => {
  if (!props.password) return 0
  
  let count = 0
  if (criteria.value.length) count++
  if (criteria.value.lowercase) count++
  if (criteria.value.uppercase) count++
  if (criteria.value.number) count++
  if (criteria.value.special) count++
  
  // Возвращаем оценку от 1 до 4
  if (count <= 1) return 1 // Очень слабый
  if (count === 2 || count === 3) return 2 // Слабый
  if (count === 4) return 3 // Средний
  return 4 // Надежный
})

// Название уровня сложности пароля
const strengthLabel = computed(() => {
  if (!props.password) return ''
  const val = score.value
  if (val === 1) return t('password_strength.very_weak')
  if (val === 2) return t('password_strength.weak')
  if (val === 3) return t('password_strength.medium')
  return t('password_strength.strong')
})

// Цвет текста для названия уровня сложности
const strengthColorText = computed(() => {
  const val = score.value
  if (val === 1) return 'text-rose-500 dark:text-rose-400'
  if (val === 2) return 'text-amber-500 dark:text-amber-400'
  if (val === 3) return 'text-blue-500 dark:text-blue-400'
  if (val === 4) return 'text-emerald-500 dark:text-emerald-400'
  return 'text-slate-400'
})

// Классы стилизации полос сложности
const getSegmentClass = (index) => {
  const val = score.value
  if (!props.password || index > val) {
    return 'bg-slate-200 dark:bg-slate-800'
  }
  
  if (val === 1) return 'bg-rose-500 dark:bg-rose-600'
  if (val === 2) return 'bg-amber-500 dark:bg-amber-600'
  if (val === 3) return 'bg-blue-500 dark:bg-blue-600'
  return 'bg-emerald-500 dark:bg-emerald-600'
}
</script>
