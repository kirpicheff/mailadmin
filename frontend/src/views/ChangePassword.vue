<template>
  <div class="min-h-screen flex items-center justify-center bg-slate-100 dark:bg-slate-900 px-4">
    <div class="max-w-md w-full glass-panel p-10 space-y-8">
      <div class="text-center">
        <h2 class="text-3xl font-extrabold text-slate-900 dark:text-white">{{ t('change_password.title') }}</h2>
        <p class="mt-2 text-sm text-slate-600 dark:text-slate-400">{{ t('change_password.subtitle') }}</p>
      </div>
      <form class="mt-8 space-y-6" @submit.prevent="handleChangePassword">
        <div v-if="error" class="bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 p-3 rounded-lg text-sm text-center">
          {{ error }}
        </div>
        <div class="space-y-4">
          <input v-model="newPassword" type="password" required class="appearance-none rounded-xl block w-full px-4 py-3 border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-900 dark:text-white sm:text-sm" :placeholder="t('change_password.new_password')" />
          <PasswordStrength :password="newPassword" />
          <input v-model="confirmPassword" type="password" required class="appearance-none rounded-xl block w-full px-4 py-3 border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-900 dark:text-white sm:text-sm" :placeholder="t('change_password.confirm_password')" />
        </div>

        <button :disabled="loading" type="submit" class="w-full py-3 px-4 border border-transparent text-sm font-bold rounded-xl text-white bg-mail-blue-600 hover:bg-mail-blue-700 focus:outline-none transition-all disabled:opacity-50">
          {{ loading ? t('change_password.saving') : t('change_password.button') }}
        </button>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import api from '@/api/axios'
import { useAuthStore } from '@/store/auth'
import PasswordStrength from '@/components/PasswordStrength.vue'
import { useI18n } from 'vue-i18n'

const router = useRouter()
const authStore = useAuthStore()
const { t } = useI18n()

const newPassword = ref('')
const confirmPassword = ref('')
const error = ref('')
const loading = ref(false)

const handleChangePassword = async () => {
  if (newPassword.value !== confirmPassword.value) {
    error.value = t('change_password.errors.mismatch')
    return
  }

  if (newPassword.value.length < 6) {
    error.value = t('change_password.errors.short')
    return
  }
  
  loading.value = true
  try {
    await api.post('/auth/change-password', { new_password: newPassword.value })
    alert(t('change_password.success'))
    
    // Очищаем всё и на логин
    await authStore.logout()
    router.push('/login')
  } catch (err) {
    error.value = t('change_password.errors.general') + ': ' + (err.response?.data?.error || err.message)
  } finally {
    loading.value = false
  }
}
</script>
