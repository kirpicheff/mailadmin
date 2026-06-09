<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/store/auth'
import { useI18n } from 'vue-i18n'

const authStore = useAuthStore()
const isDarkMode = ref(false)
const route = useRoute()
const router = useRouter()
const { t, locale } = useI18n()

// Список страниц без сайдбара
const fullScreenPages = ['/login', '/change-password']
const isFullScreen = computed(() => fullScreenPages.includes(route.path))

const toggleTheme = () => {
  isDarkMode.value = !isDarkMode.value
  const html = document.documentElement
  html.classList.toggle('dark', isDarkMode.value)
  html.style.colorScheme = isDarkMode.value ? 'dark' : 'light'
  localStorage.setItem('theme', isDarkMode.value ? 'dark' : 'light')
}

const toggleLanguage = () => {
  locale.value = locale.value === 'ru' ? 'en' : 'ru'
  localStorage.setItem('locale', locale.value)
}

const handleLogout = async () => {
  await authStore.logout()
  router.push('/login')
}

onMounted(() => {
  const savedTheme = localStorage.getItem('theme')
  const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
  
  if (savedTheme === 'dark' || (!savedTheme && prefersDark)) {
    isDarkMode.value = true
    document.documentElement.classList.add('dark')
    document.documentElement.style.colorScheme = 'dark'
  }
})

const menuItems = computed(() => {
  const items = [
    { name: t('menu.dashboard'), path: '/', icon: 'M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z' },
    { name: t('menu.domains'), path: '/domains', icon: 'M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9' },
    { name: t('menu.mailboxes'), path: '/mailboxes', icon: 'M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z' },
    { name: t('menu.tools'), path: '/tools/send-mail', icon: 'M11 5.882V19.24a1.76 1.76 0 01-3.417.592l-2.147-6.167M11 5.882c.443-.443 1.108-.592 1.69-.346l5.132 2.148c.582.246.963.81.963 1.441v5.66c0 .631-.381 1.195-.963 1.441l-5.132 2.148c-.582.246-1.247.097-1.69-.346M11 5.882V19.24' },
    { name: t('menu.logs'), path: '/logs', icon: 'M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z' },
  ]
  
  if (authStore.user?.superadmin) {
    items.push({ name: t('menu.queue'), path: '/system/queue', icon: 'M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10' })
    items.push({ name: t('menu.server_logs'), path: '/system/server-logs', icon: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z' })
  }

  items.push({ name: t('menu.settings'), path: '/settings/admins', icon: 'M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z' })
  
  return items
})
</script>

<template>
  <div v-if="isFullScreen">
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
        <a 
          href="https://github.com/kirpicheff/mailadmin" 
          target="_blank" 
          rel="noopener noreferrer"
          class="mt-3 flex items-center gap-2 text-sm text-slate-500 dark:text-slate-400 hover:text-mail-blue-600 dark:hover:text-mail-blue-400 transition-colors"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 24 24" fill="currentColor">
            <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
          </svg>
          <span>GitHub</span>
        </a>
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
          class="flex items-center gap-3 px-4 py-2 w-full rounded-xl hover:bg-slate-200/50 dark:hover:bg-slate-800/50 transition-colors text-slate-700 dark:text-slate-300"
        >
          <template v-if="isDarkMode">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-amber-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v1m0 16v1m9-9h1M4 9h1m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" />
            </svg>
            <span class="text-sm font-medium">{{ t('common.theme.light') }}</span>
          </template>
          <template v-else>
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-slate-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
            </svg>
            <span class="text-sm font-medium">{{ t('common.theme.dark') }}</span>
          </template>
        </button>

        <!-- Language Toggle -->
        <button 
          @click="toggleLanguage"
          class="flex items-center gap-3 px-4 py-2 w-full rounded-xl hover:bg-slate-200/50 dark:hover:bg-slate-800/50 transition-colors text-slate-700 dark:text-slate-300"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-mail-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 5h12M9 3v2m1.048 9.5A18.022 18.022 0 016.412 9m6.088 9h7M11 21l5-10 5 10M12.751 5C11.783 11.37 9.188 17.011 5 21M5 5a12.015 12.015 0 018.112 3.058" />
          </svg>
          <span class="text-sm font-medium">{{ locale === 'ru' ? t('common.language.en') : t('common.language.ru') }}</span>
        </button>

        <!-- Logout Button -->
        <button 
          @click="handleLogout"
          class="flex items-center gap-3 px-4 py-2 w-full rounded-xl hover:bg-red-50 dark:hover:bg-red-900/20 text-red-600 transition-colors"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
          </svg>
          <span class="text-sm font-medium">{{ t('common.logout') }}</span>
        </button>

        <!-- User Info Slot -->
        <div class="flex items-center gap-3 px-2 pt-2 border-t border-slate-100 dark:border-slate-800">
          <div class="w-10 h-10 rounded-full bg-mail-blue-100 dark:bg-mail-blue-900/30 text-mail-blue-600 flex items-center justify-center font-bold">
            {{ authStore.user?.username?.charAt(0).toUpperCase() || 'A' }}
          </div>
          <div class="flex-1 min-w-0">
            <p class="text-sm font-bold truncate text-slate-900 dark:text-white">{{ authStore.user?.username }}</p>
            <p class="text-xs text-slate-500 truncate">{{ authStore.user?.superadmin ? t('common.superadmin') : t('common.domain_admin') }}</p>
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
