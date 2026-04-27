<template>
  <div class="min-h-screen flex items-center justify-center bg-slate-100 dark:bg-slate-900 px-4">
    <div class="max-w-md w-full glass-panel p-10 space-y-8">
      <div class="text-center">
        <h2 class="text-3xl font-extrabold text-slate-900 dark:text-white">MailAdmin 2026</h2>
        <p class="mt-2 text-sm text-slate-600 dark:text-slate-400">Войдите в панель управления</p>
      </div>
      <form class="mt-8 space-y-6" @submit.prevent="handleLogin">
        <div class="space-y-4">
          <div>
            <label for="username" class="sr-only">Username</label>
            <input id="username" name="username" type="text" required class="appearance-none rounded-xl relative block w-full px-4 py-3 border border-slate-300 dark:border-slate-700 bg-white/50 dark:bg-slate-800/50 text-slate-900 dark:text-white focus:outline-none focus:ring-mail-blue-500 focus:border-mail-blue-500 sm:text-sm" placeholder="Имя пользователя" />
          </div>
          <div>
            <label for="password" class="sr-only">Password</label>
            <input id="password" name="password" type="password" required class="appearance-none rounded-xl relative block w-full px-4 py-3 border border-slate-300 dark:border-slate-700 bg-white/50 dark:bg-slate-800/50 text-slate-900 dark:text-white focus:outline-none focus:ring-mail-blue-500 focus:border-mail-blue-500 sm:text-sm" placeholder="Пароль" />
          </div>
        </div>

        <div>
          <button type="submit" class="group relative w-full flex justify-center py-3 px-4 border border-transparent text-sm font-bold rounded-xl text-white bg-mail-blue-600 hover:bg-mail-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-mail-blue-500 transition-all">
            Войти
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'
const router = useRouter()
const handleLogin = () => {
  // Заглушка
  router.push('/')
}
</script>
