<template>
  <div class="min-h-screen flex items-center justify-center bg-slate-100 dark:bg-slate-900 px-4">
    <div class="max-w-md w-full glass-panel p-10 space-y-8">
      <div class="text-center">
        <h2 class="text-3xl font-extrabold text-slate-900 dark:text-white">Смена пароля</h2>
        <p class="mt-2 text-sm text-slate-600 dark:text-slate-400">Ваш пароль истек. Пожалуйста, установите новый.</p>
      </div>
      <form class="mt-8 space-y-6" @submit.prevent="handleChangePassword">
        <div v-if="error" class="bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 p-3 rounded-lg text-sm text-center">
          {{ error }}
        </div>
        <div class="space-y-4">
          <input v-model="newPassword" type="password" required class="appearance-none rounded-xl block w-full px-4 py-3 border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-900 dark:text-white sm:text-sm" placeholder="Новый пароль" />
          <input v-model="confirmPassword" type="password" required class="appearance-none rounded-xl block w-full px-4 py-3 border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-900 dark:text-white sm:text-sm" placeholder="Повторите пароль" />
        </div>

        <button :disabled="loading" type="submit" class="w-full py-3 px-4 border border-transparent text-sm font-bold rounded-xl text-white bg-mail-blue-600 hover:bg-mail-blue-700 focus:outline-none transition-all disabled:opacity-50">
          {{ loading ? 'Сохранение...' : 'Сменить пароль' }}
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

const router = useRouter()
const authStore = useAuthStore()

const newPassword = ref('')
const confirmPassword = ref('')
const error = ref('')
const loading = ref(false)

const handleChangePassword = async () => {
  if (newPassword.value !== confirmPassword.value) {
    error.value = 'Пароли не совпадают'
    return
  }

  if (newPassword.value.length < 6) {
    error.value = 'Пароль слишком короткий (мин. 6 символов)'
    return
  }
  
  loading.value = true
  try {
    await api.post('/auth/change-password', { new_password: newPassword.value })
    alert('Пароль успешно изменен! Пожалуйста, войдите снова.')
    
    // Очищаем всё и на логин
    await authStore.logout()
    router.push('/login')
  } catch (err) {
    error.value = 'Ошибка при смене пароля: ' + (err.response?.data?.error || err.message)
  } finally {
    loading.value = false
  }
}
</script>
