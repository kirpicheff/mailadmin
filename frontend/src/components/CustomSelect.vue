<template>
  <div class="relative" ref="selectRef">
    <div 
      @click="toggle"
      :class="[
        'flex items-center justify-between px-5 py-3.5 rounded-2xl border-2 transition-all cursor-pointer font-bold select-none',
        isOpen ? 'border-mail-blue-500 ring-4 ring-mail-blue-500/5 bg-white dark:bg-slate-900' : 'border-slate-100 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/30',
        disabled ? 'opacity-50 cursor-not-allowed' : ''
      ]"
    >
      <span :class="modelValue ? 'text-slate-900 dark:text-white' : 'text-slate-400'">
        {{ selectedLabel || placeholder || t('mailboxes.select_prompt') }}
      </span>
      <svg 
        :class="['w-5 h-5 text-slate-400 transition-transform duration-300', isOpen ? 'rotate-180 text-mail-blue-500' : '']" 
        fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M19 9l-7 7-7-7" />
      </svg>
    </div>

    <!-- Dropdown -->
    <transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="transform scale-95 opacity-0 -translate-y-2"
      enter-to-class="transform scale-100 opacity-100 translate-y-0"
      leave-active-class="transition duration-150 ease-in"
      leave-from-class="transform scale-100 opacity-100 translate-y-0"
      leave-to-class="transform scale-95 opacity-0 -translate-y-2"
    >
      <div 
        v-if="isOpen && !disabled" 
        class="absolute z-[110] left-0 right-0 mt-2 bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-[20px] shadow-2xl overflow-hidden max-h-60 overflow-y-auto backdrop-blur-xl"
      >
        <div 
          v-for="option in options" 
          :key="option.value"
          @click="select(option)"
          :class="[
            'px-5 py-3 cursor-pointer transition-colors font-bold text-sm',
            modelValue === option.value ? 'bg-mail-blue-600 text-white' : 'text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-800 hover:text-slate-950 dark:hover:text-white'
          ]"
        >
          {{ option.label }}
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const props = defineProps({
  modelValue: [String, Number],
  options: {
    type: Array, // [{ label: 'Name', value: 'val' }]
    default: () => []
  },
  placeholder: {
    type: String,
    default: ''
  },
  disabled: Boolean
})

const emit = defineEmits(['update:modelValue', 'change'])

const isOpen = ref(false)
const selectRef = ref(null)

const selectedLabel = computed(() => {
  const option = props.options.find(o => o.value === props.modelValue)
  return option ? option.label : ''
})

const toggle = () => {
  if (!props.disabled) isOpen.value = !isOpen.value
}

const select = (option) => {
  emit('update:modelValue', option.value)
  emit('change', option.value)
  isOpen.value = false
}

const handleClickOutside = (event) => {
  if (selectRef.value && !selectRef.value.contains(event.target)) {
    isOpen.value = false
  }
}

onMounted(() => document.addEventListener('click', handleClickOutside))
onUnmounted(() => document.removeEventListener('click', handleClickOutside))
</script>
