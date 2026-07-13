<template>
  <div class="p-6">
    <!-- Шапка страницы -->
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-2xl font-bold text-slate-900 dark:text-white">{{ t('admins.title') }}</h1>
    </div>

    <!-- Вкладки управления -->
    <div class="flex border-b border-slate-200 dark:border-slate-800 mb-6">
      <button @click="activeTab = 'admins'" :class="activeTab === 'admins' ? 'border-mail-blue-500 text-mail-blue-600 dark:text-mail-blue-400 font-bold border-b-2' : 'text-slate-500 dark:text-slate-400'" class="px-6 py-3 text-sm focus:outline-none transition-all">
        {{ t('admins.title') }}
      </button>
      <button @click="activeTab = 'notifications'; fetchNotificationRules()" :class="activeTab === 'notifications' ? 'border-mail-blue-500 text-mail-blue-600 dark:text-mail-blue-400 font-bold border-b-2' : 'text-slate-500 dark:text-slate-400'" class="px-6 py-3 text-sm focus:outline-none transition-all">
        {{ t('notifications.tab_title') }}
      </button>
    </div>

    <!-- Вкладка администраторов -->
    <div v-if="activeTab === 'admins'" class="space-y-6">
      <div class="flex justify-between items-center">
        <h2 class="text-lg font-bold text-slate-900 dark:text-white">{{ t('admins.title') }}</h2>
        <button @click="openCreateModal" class="px-4 py-2 bg-mail-blue-600 hover:bg-mail-blue-700 text-white rounded-lg transition-colors">
           {{ t('admins.add') }}
        </button>
      </div>

      <div class="glass-panel overflow-hidden">
        <table class="w-full text-left">
          <thead class="bg-slate-50 dark:bg-slate-800/50">
            <tr>
              <th class="px-6 py-3 text-xs font-medium text-slate-500 uppercase">{{ t('admins.table.user') }}</th>
              <th class="px-6 py-3 text-xs font-medium text-slate-500 uppercase">{{ t('admins.table.rights') }}</th>
              <th class="px-6 py-3 text-xs font-medium text-slate-500 uppercase">{{ t('admins.table.active') }}</th>
              <th class="px-6 py-3 text-xs font-medium text-slate-500 uppercase">{{ t('admins.table.contacts') }}</th>
              <th class="px-6 py-3 text-xs font-medium text-slate-500 uppercase">{{ t('common.actions') }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-200 dark:divide-slate-700/50">
            <tr v-for="admin in admins" :key="admin.username" class="group hover:bg-slate-50/80 dark:hover:bg-mail-blue-500/5 transition-all duration-200">
              <td class="px-6 py-4">
                <div class="flex items-center gap-3">
                  <div class="w-10 h-10 rounded-full bg-mail-blue-100 dark:bg-mail-blue-900/30 flex items-center justify-center text-mail-blue-600 dark:text-mail-blue-400 font-bold uppercase">
                    {{ admin.username.charAt(0) }}
                  </div>
                  <div>
                    <div class="text-sm font-bold text-slate-900 dark:text-white">{{ admin.username }}</div>
                    <div class="text-xs text-slate-500">{{ admin.email_other || t('admins.no_extra_email') }}</div>
                  </div>
                </div>
              </td>
              <td class="px-6 py-4">
                <span v-if="admin.superadmin" class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-purple-100 text-purple-700 dark:bg-purple-900/40 dark:text-purple-300 border border-purple-200 dark:border-purple-800">
                  <span class="w-1.5 h-1.5 rounded-full bg-purple-500 mr-1.5 shadow-[0_0_8px_rgba(168,85,247,0.5)]"></span>
                  {{ t('common.superadmin') }}
                </span>
                <span v-else class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-mail-blue-100 text-mail-blue-700 dark:bg-mail-blue-900/40 dark:text-mail-blue-300 border border-mail-blue-200 dark:border-mail-blue-800">
                  <span class="w-1.5 h-1.5 rounded-full bg-mail-blue-500 mr-1.5 shadow-[0_0_8px_rgba(59,130,246,0.5)]"></span>
                  {{ t('common.domain_admin') }}
                </span>
              </td>
              <td class="px-6 py-4">
                <button @click="toggleAdminStatus(admin)" :class="admin.active ? 'bg-green-500' : 'bg-slate-300 dark:bg-slate-700'" class="relative inline-flex h-6 w-11 items-center rounded-full transition-colors focus:outline-none">
                  <span :class="admin.active ? 'translate-x-6' : 'translate-x-1'" class="inline-block h-4 w-4 transform rounded-full bg-white transition-transform" />
                </button>
              </td>
              <td class="px-6 py-4 text-sm text-slate-600 dark:text-slate-400">
                <div v-if="admin.phone" class="flex items-center gap-1.5 font-medium">
                  <svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.5 19.5 0 0 1-6-6 19.79 19.79 0 0 1-3.07-8.67A2 2 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72 12.84 12.84 0 0 0 .7 2.81 2 2 0 0 1-.45 2.11L8.09 9.91a16 16 0 0 0 6 6l1.27-1.27a2 2 0 0 1 2.11-.45 12.84 12.84 0 0 0 2.81.7A2 2 0 0 1 22 16.92z" /></svg>
                  {{ admin.phone }}
                </div>
                <div class="text-xs opacity-40" v-else>{{ t('admins.not_specified') }}</div>
              </td>
              <td class="px-6 py-4">
                <div class="flex items-center gap-2">
                  <button @click="editAdmin(admin)" class="p-2 bg-slate-100 hover:bg-mail-blue-100 text-slate-600 hover:text-mail-blue-600 dark:bg-slate-800/50 dark:hover:bg-mail-blue-900/40 dark:text-slate-400 dark:hover:text-mail-blue-300 rounded-xl transition-all duration-200" :title="t('common.edit')">
                    <svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7M18.5 2.5a2.121 2.121 0 1 1 3 3L12 15l-4 1 1-4 9.5-9.5z" /></svg>
                  </button>
                  <button @click="deleteAdmin(admin.username)" class="p-2 bg-slate-100 hover:bg-red-100 text-slate-600 hover:text-red-600 dark:bg-slate-800/50 dark:hover:bg-red-900/40 dark:text-slate-400 dark:hover:text-red-300 rounded-xl transition-all duration-200" :title="t('common.delete')">
                    <svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 6h18M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2M10 11v6M14 11v6" /></svg>
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Вкладка настроек уведомлений -->
    <div v-else-if="activeTab === 'notifications'" class="space-y-6">
      <div class="flex justify-between items-center">
        <div>
          <h2 class="text-xl font-bold text-slate-900 dark:text-white">{{ t('notifications.title') }}</h2>
          <p class="text-xs text-slate-500 mt-1">{{ t('notifications.subtitle') }}</p>
        </div>
      </div>

      <!-- Добавить получателя -->
      <div class="bg-white dark:bg-slate-900 p-6 rounded-2xl border border-slate-200 dark:border-slate-800 shadow-sm flex flex-col md:flex-row gap-4 items-end">
        <div class="flex-1 space-y-2">
          <label class="block text-xs font-bold uppercase tracking-wider text-slate-500">{{ t('notifications.add_email') }}</label>
          <input v-model="newRecipientEmail" type="email" :placeholder="t('notifications.email_placeholder')" class="w-full px-4 py-2.5 rounded-xl border border-slate-200 dark:border-slate-700 bg-slate-50/50 dark:bg-slate-800/30 text-slate-900 dark:text-white focus:ring-2 focus:ring-mail-blue-500 outline-none transition-all font-bold text-sm" />
        </div>
        <button @click="addRecipient" class="px-6 py-2.5 bg-mail-blue-600 hover:bg-mail-blue-700 text-white rounded-xl font-bold transition-all shadow-md active:scale-95 text-sm whitespace-nowrap">
          {{ t('common.create') }}
        </button>
      </div>

      <!-- Список получателей -->
      <div v-if="groupedRules.length === 0" class="text-center py-12 text-slate-500 dark:text-slate-400">
        {{ t('notifications.no_rules') }}
      </div>

      <div v-else class="space-y-4">
        <div v-for="recipient in groupedRules" :key="recipient.email" class="bg-white dark:bg-slate-900 rounded-2xl border border-slate-200 dark:border-slate-800 shadow-sm overflow-hidden transition-all duration-300">
          <div class="px-6 py-4 bg-slate-50/50 dark:bg-slate-800/30 border-b border-slate-100 dark:border-slate-800 flex justify-between items-center">
            <div class="flex items-center gap-3">
              <span class="font-bold text-slate-800 dark:text-slate-200 text-sm">{{ recipient.email }}</span>
              <button @click="toggleRecipientActive(recipient)" :class="recipient.active ? 'bg-green-500' : 'bg-slate-300 dark:bg-slate-700'" class="relative inline-flex h-5 w-9 items-center rounded-full transition-colors focus:outline-none scale-75">
                <span :class="recipient.active ? 'translate-x-4' : 'translate-x-1'" class="inline-block h-3.5 w-3.5 transform rounded-full bg-white transition-transform" />
              </button>
              <span class="text-[10px] uppercase font-bold text-slate-400">{{ t('notifications.active') }}: {{ recipient.active ? t('common.yes') : t('common.no') }}</span>
            </div>
            <div class="flex items-center gap-2">
              <button @click="saveRecipientRules(recipient)" class="px-3 py-1 bg-green-600 hover:bg-green-700 text-white text-xs font-bold rounded-lg transition-all active:scale-95">
                {{ t('common.save') }}
              </button>
              <button @click="deleteRecipient(recipient.email)" class="p-1.5 hover:bg-red-50 dark:hover:bg-red-900/20 text-slate-400 hover:text-red-500 rounded-lg transition-all">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
              </button>
            </div>
          </div>

          <div class="p-6 space-y-4">
            <div class="flex justify-between items-center">
              <span class="text-xs font-bold text-slate-400 uppercase tracking-wider">{{ t('notifications.domains_list') }}</span>
              <div class="flex gap-2">
                <button @click="selectDomainsForRecipient(recipient, true)" class="text-[10px] text-mail-blue-600 dark:text-mail-blue-400 hover:underline font-bold uppercase tracking-wider">{{ t('notifications.select_all') }}</button>
                <span class="text-slate-300">|</span>
                <button @click="selectDomainsForRecipient(recipient, false)" class="text-[10px] text-slate-500 hover:underline font-bold uppercase tracking-wider">{{ t('notifications.deselect_all') }}</button>
              </div>
            </div>

            <div class="grid grid-cols-2 md:grid-cols-4 gap-3 max-h-48 overflow-y-auto p-4 bg-slate-50 dark:bg-slate-900/30 rounded-xl border border-slate-100 dark:border-slate-800 shadow-inner">
              <button v-for="d in allDomains" :key="typeof d === 'string' ? d : d.domain" 
                type="button"
                @click="toggleDomainForRecipient(recipient, typeof d === 'string' ? d : d.domain)"
                class="flex items-center justify-between gap-3 bg-white dark:bg-slate-900 p-2.5 px-4 rounded-xl border border-transparent hover:border-mail-blue-500/50 transition-all cursor-pointer shadow-sm hover:shadow text-left">
                <span class="text-xs font-bold text-slate-700 dark:text-slate-300 truncate select-none" 
                  :class="{ 'text-mail-blue-500': recipient.domains.includes(typeof d === 'string' ? d : d.domain) }">
                  {{ typeof d === 'string' ? d : d.domain }}
                </span>
                <div :class="recipient.domains.includes(typeof d === 'string' ? d : d.domain) ? 'bg-mail-blue-500' : 'bg-slate-300 dark:bg-slate-700'" 
                  class="relative inline-flex h-4 w-7 items-center rounded-full transition-colors flex-shrink-0 pointer-events-none scale-90">
                  <span :class="recipient.domains.includes(typeof d === 'string' ? d : d.domain) ? 'translate-x-3.5' : 'translate-x-0.5'" class="inline-block h-3.5 w-3.5 transform rounded-full bg-white transition-transform" />
                </div>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Модальное окно создания/редактирования администратора -->
    <div v-if="showModal" class="fixed inset-0 bg-slate-950/40 backdrop-blur-md flex items-center justify-center p-4 z-50 transition-all duration-300">
      <div class="bg-white dark:bg-slate-900 p-8 rounded-3xl w-full max-w-lg shadow-[0_20px_50px_rgba(0,0,0,0.3)] space-y-6 border border-white/20 dark:border-slate-800/50">
        <div class="flex justify-between items-center">
          <h3 class="text-2xl font-black text-slate-900 dark:text-white tracking-tight">
            {{ isEdit ? t('admins.modal.edit_title') : t('admins.modal.add_title') }}
          </h3>
          <button @click="showModal = false" class="text-slate-400 hover:text-slate-600 dark:hover:text-white transition-colors">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
          </button>
        </div>
        
        <form @submit.prevent="saveAdmin" class="space-y-5">
          <div class="space-y-4">
            <!-- Имя пользователя (Email) -->
            <div class="relative">
              <label class="block text-xs font-bold uppercase tracking-wider text-slate-500 mb-1.5 ml-1">{{ t('admins.modal.username_label') }}</label>
              <div class="relative">
                <span class="absolute left-4 top-1/2 -translate-y-1/2 text-slate-400">
                  <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2M12 7a4 4 0 1 1-8 0 4 4 0 0 1 8 0z" /></svg>
                </span>
                <input v-model="form.username" :disabled="isEdit" type="email" required 
                  class="w-full pl-12 pr-4 py-3 rounded-2xl border-0 bg-slate-100 dark:bg-slate-800/50 dark:text-white focus:ring-2 focus:ring-mail-blue-500 transition-all disabled:opacity-50" />
              </div>
            </div>

            <!-- Пароль и опции -->
            <div class="grid grid-cols-2 gap-5">
              <div class="relative">
                <label class="block text-xs font-bold uppercase tracking-wider text-slate-500 mb-1.5 ml-1">{{ t('admins.modal.password_label') }}</label>
                <div class="relative">
                  <span class="absolute left-4 top-1/2 -translate-y-1/2 text-slate-400">
                    <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path d="M12 15V17M12 15V13M12 15H10M12 15H14M12 15V17" /><rect x="3" y="11" width="18" height="11" rx="2" /><path d="M7 11V7a5 5 0 0 1 10 0v4" /></svg>
                  </span>
                  <input v-model="form.password" :required="!isEdit" type="password" 
                    placeholder="••••••••"
                    class="w-full pl-12 pr-4 py-3 rounded-2xl border-0 bg-slate-100 dark:bg-slate-800/50 dark:text-white focus:ring-2 focus:ring-mail-blue-500 transition-all" />
                </div>
              </div>
              <div class="flex flex-col justify-center gap-4">
                <!-- Переключатель Суперадмина -->
                <label class="flex items-center justify-between gap-4 cursor-pointer group bg-slate-100 dark:bg-slate-800/50 p-2.5 px-4 rounded-2xl transition-all hover:bg-slate-200 dark:hover:bg-slate-700/50">
                  <span class="text-xs font-bold uppercase tracking-wider text-slate-500 group-hover:text-mail-blue-500 transition-colors">{{ t('admins.modal.superadmin_label') }}</span>
                  <div @click="form.superadmin = !form.superadmin" 
                    :class="form.superadmin ? 'bg-purple-500' : 'bg-slate-300 dark:bg-slate-600'" 
                    class="relative inline-flex h-6 w-11 items-center rounded-full transition-colors focus:outline-none flex-shrink-0">
                    <span :class="form.superadmin ? 'translate-x-6' : 'translate-x-1'" class="inline-block h-4 w-4 transform rounded-full bg-white transition-transform shadow-sm" />
                  </div>
                </label>
                <!-- Переключатель Активности -->
                <label class="flex items-center justify-between gap-4 cursor-pointer group bg-slate-100 dark:bg-slate-800/50 p-2.5 px-4 rounded-2xl transition-all hover:bg-slate-200 dark:hover:bg-slate-700/50">
                  <span class="text-xs font-bold uppercase tracking-wider text-slate-500 group-hover:text-green-500 transition-colors">{{ t('admins.modal.status_active') }}</span>
                  <div @click="form.active = !form.active" 
                    :class="form.active ? 'bg-green-500' : 'bg-slate-300 dark:bg-slate-600'" 
                    class="relative inline-flex h-6 w-11 items-center rounded-full transition-colors focus:outline-none flex-shrink-0">
                    <span :class="form.active ? 'translate-x-6' : 'translate-x-1'" class="inline-block h-4 w-4 transform rounded-full bg-white transition-transform shadow-sm" />
                  </div>
                </label>
                <!-- Переключатель получения паролей для суперадмина -->
                <label v-if="form.superadmin" class="flex items-center justify-between gap-4 cursor-pointer group bg-slate-100 dark:bg-slate-800/50 p-2.5 px-4 rounded-2xl transition-all hover:bg-slate-200 dark:hover:bg-slate-700/50">
                  <span class="text-xs font-bold uppercase tracking-wider text-slate-500 group-hover:text-mail-blue-500 transition-colors text-left">{{ t('admins.modal.receive_passwords') }}</span>
                  <div @click="form.receive_passwords = !form.receive_passwords" 
                    :class="form.receive_passwords ? 'bg-mail-blue-500' : 'bg-slate-300 dark:bg-slate-600'" 
                    class="relative inline-flex h-6 w-11 items-center rounded-full transition-colors focus:outline-none flex-shrink-0">
                    <span :class="form.receive_passwords ? 'translate-x-6' : 'translate-x-1'" class="inline-block h-4 w-4 transform rounded-full bg-white transition-transform shadow-sm" />
                  </div>
                </label>
              </div>
            </div>

            <!-- Дополнительная почта и телефон -->
            <div class="grid grid-cols-2 gap-5">
              <div class="relative">
                <label class="block text-xs font-bold uppercase tracking-wider text-slate-500 mb-1.5 ml-1">{{ t('admins.modal.extra_email') }}</label>
                <div class="relative">
                  <span class="absolute left-4 top-1/2 -translate-y-1/2 text-slate-400">
                    <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path d="M3 8l7.89 5.26a2 2 0 0 0 2.22 0L21 8M5 19h14a2 2 0 0 0 2-2V7a2 2 0 0 0-2-2H5a2 2 0 0 0-2 2v10a2 2 0 0 0 2 2z" /></svg>
                  </span>
                  <input v-model="form.email_other" type="email" placeholder="backup@email.com"
                    class="w-full pl-12 pr-4 py-3 rounded-2xl border-0 bg-slate-100 dark:bg-slate-800/50 dark:text-white focus:ring-2 focus:ring-mail-blue-500 transition-all" />
                </div>
              </div>
              <div class="relative">
                <label class="block text-xs font-bold uppercase tracking-wider text-slate-500 mb-1.5 ml-1">{{ t('admins.modal.phone') }}</label>
                <div class="relative">
                  <span class="absolute left-4 top-1/2 -translate-y-1/2 text-slate-400">
                    <svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.5 19.5 0 0 1-6-6 19.79 19.79 0 0 1-3.07-8.67A2 2 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72 12.84 12.84 0 0 0 .7 2.81 2 2 0 0 1-.45 2.11L8.09 9.91a16 16 0 0 0 6 6l1.27-1.27a2 2 0 0 1 2.11-.45 12.84 12.84 0 0 0 2.81.7A2 2 0 0 1 22 16.92z" /></svg>
                  </span>
                  <input v-model="form.phone" type="text" placeholder="+7..."
                    class="w-full pl-12 pr-4 py-3 rounded-2xl border-0 bg-slate-100 dark:bg-slate-800/50 dark:text-white focus:ring-2 focus:ring-mail-blue-500 transition-all" />
                </div>
              </div>
            </div>

            <!-- Доступные домены (если не суперадмин) -->
            <div v-if="!form.superadmin" class="space-y-3 pt-2">
              <label class="block text-xs font-bold uppercase tracking-wider text-slate-500 ml-1">{{ t('admins.modal.domain_access') }}</label>
              <div class="grid grid-cols-2 gap-3 max-h-60 overflow-y-auto p-4 bg-slate-100 dark:bg-slate-800/50 rounded-3xl border border-slate-200 dark:border-slate-800/50 shadow-inner">
                <button v-for="d in allDomains" :key="typeof d === 'string' ? d : d.domain" 
                  type="button"
                  @click="toggleDomain(typeof d === 'string' ? d : d.domain)"
                  class="flex items-center justify-between gap-3 bg-white dark:bg-slate-900 p-3 px-4 rounded-2xl border border-transparent hover:border-mail-blue-500/50 transition-all cursor-pointer group shadow-sm hover:shadow-md text-left">
                  <span class="text-xs font-bold text-slate-700 dark:text-slate-300 truncate select-none" 
                    :class="{ 'text-mail-blue-500': form.domains.includes(typeof d === 'string' ? d : d.domain) }">
                    {{ typeof d === 'string' ? d : d.domain }}
                  </span>
                  <div :class="form.domains.includes(typeof d === 'string' ? d : d.domain) ? 'bg-mail-blue-500' : 'bg-slate-300 dark:bg-slate-700'" 
                    class="relative inline-flex h-5 w-9 items-center rounded-full transition-colors flex-shrink-0 pointer-events-none">
                    <span :class="form.domains.includes(typeof d === 'string' ? d : d.domain) ? 'translate-x-5' : 'translate-x-1'" class="inline-block h-3 w-3 transform rounded-full bg-white transition-transform" />
                  </div>
                </button>
              </div>
            </div>
          </div>

          <div class="flex gap-4 pt-4">
            <button type="submit" class="flex-1 py-4 bg-mail-blue-600 text-white rounded-2xl font-bold hover:bg-mail-blue-700 shadow-lg shadow-mail-blue-500/30 transition-all active:scale-95">
              {{ isEdit ? t('admins.modal.edit_title') : t('admins.modal.add_title') }}
            </button>
            <button type="button" @click="showModal = false" class="px-8 py-4 bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 rounded-2xl font-bold hover:bg-slate-200 dark:hover:bg-slate-700 transition-all">
              {{ t('common.cancel') }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, reactive } from 'vue'
import api from '@/api/axios'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const admins = ref([])
const allDomains = ref([])
const showModal = ref(false)
const isEdit = ref(false)
const activeTab = ref('admins')
const newRecipientEmail = ref('')
const groupedRules = ref([])

const form = reactive({
  username: '',
  password: '',
  superadmin: false,
  active: true,
  phone: '',
  email_other: '',
  domains: [],
  receive_passwords: true
})

const fetchAdmins = async () => {
  const { data } = await api.get('/admins')
  admins.value = data
}

const fetchDomains = async () => {
  const { data } = await api.get('/domains')
  allDomains.value = data
}

const fetchNotificationRules = async () => {
  try {
    const { data } = await api.get('/system/notifications')
    const grouped = {}
    data.forEach(rule => {
      if (!grouped[rule.email]) {
        grouped[rule.email] = {
          email: rule.email,
          domains: [],
          active: rule.active
        }
      }
      grouped[rule.email].domains.push(rule.domain)
    })
    groupedRules.value = Object.values(grouped)
  } catch (err) {
    console.error('Error fetching notification rules:', err)
  }
}

const addRecipient = () => {
  const email = newRecipientEmail.value.trim()
  if (!email) return
  if (groupedRules.value.some(r => r.email === email)) {
    alert('Этот получатель уже добавлен')
    return
  }
  groupedRules.value.push({
    email: email,
    domains: [],
    active: true
  })
  newRecipientEmail.value = ''
}

const toggleRecipientActive = (recipient) => {
  recipient.active = !recipient.active
}

const toggleDomainForRecipient = (recipient, domain) => {
  const idx = recipient.domains.indexOf(domain)
  if (idx === -1) {
    recipient.domains.push(domain)
  } else {
    recipient.domains.splice(idx, 1)
  }
}

const selectDomainsForRecipient = (recipient, selectAll) => {
  if (selectAll) {
    recipient.domains = allDomains.value.map(d => typeof d === 'string' ? d : d.domain)
  } else {
    recipient.domains = []
  }
}

const saveRecipientRules = async (recipient) => {
  try {
    await api.post('/system/notifications', {
      email: recipient.email,
      domains: recipient.domains,
      active: recipient.active
    })
    alert(t('notifications.save_success'))
    fetchNotificationRules()
  } catch (err) {
    alert(t('common.error') + ': ' + (err.response?.data?.error || err.message))
  }
}

const deleteRecipient = async (email) => {
  if (!confirm(t('notifications.delete_confirm', { email }))) return
  try {
    await api.delete('/system/notifications', { params: { email } })
    fetchNotificationRules()
  } catch (err) {
    alert(t('common.error') + ': ' + (err.response?.data?.error || err.message))
  }
}

const openCreateModal = () => {
  isEdit.value = false
  Object.assign(form, { username: '', password: '', superadmin: false, active: true, phone: '', email_other: '', domains: [], receive_passwords: true })
  showModal.value = true
}

const editAdmin = async (admin) => {
  isEdit.value = true
  const { data } = await api.get(`/admins/${admin.username}`)
  Object.assign(form, {
    username: data.admin.username,
    password: '',
    superadmin: data.admin.superadmin,
    active: data.admin.active,
    phone: data.admin.phone || '',
    email_other: data.admin.email_other || '',
    domains: data.domains || [],
    receive_passwords: data.receive_passwords !== undefined ? data.receive_passwords : true
  })
  showModal.value = true
}

const saveAdmin = async () => {
  try {
    if (isEdit.value) {
      await api.put(`/admins/${form.username}`, form)
    } else {
      await api.post('/admins', form)
    }
    showModal.value = false
    fetchAdmins()
  } catch (err) {
    alert(t('admins.errors.save') + ': ' + (err.response?.data?.error || err.message))
  }
}

const deleteAdmin = async (username) => {
  if (!confirm(t('admins.delete_confirm', { username }))) return
  try {
    await api.delete(`/admins/${username}`)
    fetchAdmins()
  } catch (err) {
    alert(t('admins.errors.delete'))
  }
}

const toggleDomain = (domain) => {
  if (!form.domains) form.domains = []
  const index = form.domains.indexOf(domain)
  if (index === -1) {
    form.domains.push(domain)
  } else {
    form.domains.splice(index, 1)
  }
}

const toggleAdminStatus = async (admin) => {
  try {
    const newStatus = !admin.active
    await api.put(`/admins/${admin.username}`, {
      ...admin,
      active: newStatus
    })
    admin.active = newStatus
  } catch (err) {
    alert(t('admins.errors.status_change'))
  }
}

onMounted(() => {
  fetchAdmins()
  fetchDomains()
})
</script>

<style scoped>
.glass-panel {
  background: rgba(255, 255, 255, 0.7);
  backdrop-filter: blur(20px);
}
.dark .glass-panel {
  background: rgba(15, 23, 42, 0.6);
}
</style>
