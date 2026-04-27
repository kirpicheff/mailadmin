<script setup>
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'

const isDarkMode = ref(false)
const route = useRoute()
const router = useRouter()

const toggleTheme = () => {
  isDarkMode.value = !isDarkMode.value
  document.documentElement.classList.toggle('dark', isDarkMode.value)
  localStorage.setItem('theme', isDarkMode.value ? 'dark' : 'light')
}

onMounted(() => {
  const savedTheme = localStorage.getItem('theme')
  if (savedTheme === 'dark' || (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
    isDarkMode.value = true
    document.documentElement.classList.add('dark')
  }
})

const menuItems = [
  { name: 'Дашборд', path: '/', icon: 'M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z' },
  { name: 'Домены', path: '/domains', icon: 'M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9' },
  { name: 'Почтовые ящики', path: '/mailboxes', icon: 'M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z' },
]
</script>

<template>
  <div v-if="route.path === '/login'">
    <router-view />
  </div>
  
  <div v-else class="min-h-screen flex bg-slate-50 dark:bg-slate-950 transition-colors duration-500">
    <!-- Сайдбар -->
    <aside class="w-72 border-r border-slate-200 dark:border-slate-800 bg-white/50 dark:bg-slate-900/50 backdrop-blur-xl flex flex-col sticky top-0 h-screen">
      <div class="p-8">
        <h2 class="text-2xl font-black text-mail-blue-600 dark:text-mail-blue-400 flex items-center gap-2">
          <span class="w-8 h-8 bg-mail-blue-600 rounded-lg flex items-center justify-center text-white text-sm">M</span>
          MailAdmin
        </h2>
      </div>

      <nav class="flex-1 px-4 space-y-2 mt-4">
        <router-link 
          v-for="item in menuItems" 
          :key="item.path"
          :to="item.path"
          class="sidebar-link"
          :class="{ 'active': route.path === item.path }"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" :d="item.icon" />
          </svg>
          {{ item.name }}
        </router-link>
      </nav>

      <div class="p-6 border-t border-slate-200 dark:border-slate-800 space-y-4">
        <!-- Theme Toggle -->
        <button 
          @click="toggleTheme"
          class="flex items-center gap-3 px-4 py-2 w-full rounded-xl hover:bg-slate-200/50 dark:hover:bg-slate-800/50 transition-colors"
        >
          <template v-if="isDarkMode">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-amber-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v1m0 16v1m9-9h1M4 9h1m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" />
            </svg>
            <span class="text-sm font-medium">Светлая тема</span>
          </template>
          <template v-else>
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-slate-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
            </svg>
            <span class="text-sm font-medium">Темная тема</span>
          </template>
        </button>

        <!-- User Info Slot -->
        <div class="flex items-center gap-3 px-2">
          <div class="w-10 h-10 rounded-full bg-slate-200 dark:bg-slate-700 flex items-center justify-center font-bold">A</div>
          <div class="flex-1 min-w-0">
            <p class="text-sm font-bold truncate">Админ</p>
            <p class="text-xs text-slate-500 truncate">superadmin</p>
          </div>
        </div>
      </div>
    </aside>

    <!-- Основной контент -->
    <main class="flex-1 p-10 overflow-y-auto">
      <router-view />
    </main>
  </div>
</template>
