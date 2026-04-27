<template>
  <div class="space-y-8">
    <header>
      <h1 class="text-3xl font-bold tracking-tight">Дашборд</h1>
      <p class="text-slate-500 dark:text-slate-400">Общая статистика почтового сервера</p>
    </header>

    <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
      <!-- Домены -->
      <div class="glass-panel p-6">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm font-medium text-slate-500 dark:text-slate-400">Всего доменов</p>
            <p class="text-3xl font-bold mt-1">12</p>
          </div>
          <div class="w-12 h-12 rounded-full bg-mail-blue-500/10 flex items-center justify-center text-mail-blue-500">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="Language-Globe" />
            </svg>
          </div>
        </div>
      </div>

      <!-- Ящики -->
      <div class="glass-panel p-6">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm font-medium text-slate-500 dark:text-slate-400">Активных ящиков</p>
            <p class="text-3xl font-bold mt-1">124</p>
          </div>
          <div class="w-12 h-12 rounded-full bg-emerald-500/10 flex items-center justify-center text-emerald-500">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
            </svg>
          </div>
        </div>
      </div>

      <!-- Алиасы -->
      <div class="glass-panel p-6">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm font-medium text-slate-500 dark:text-slate-400">Алиасов</p>
            <p class="text-3xl font-bold mt-1">45</p>
          </div>
          <div class="w-12 h-12 rounded-full bg-amber-500/10 flex items-center justify-center text-amber-500">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
            </svg>
          </div>
        </div>
      </div>
    </div>

    <!-- График или последние действия -->
    <div class="glass-panel p-8 min-h-[300px] flex items-center justify-center border-dashed border-2 border-slate-200 dark:border-slate-800">
        <p class="text-slate-400">График активности (Будет реализован позже)</p>
    </div>
  </div>
</template>
