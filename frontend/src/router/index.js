import { createRouter, createWebHashHistory } from 'vue-router'

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: '/',
      name: 'Dashboard',
      component: () => import('../views/Dashboard.vue'),
      meta: { title: 'Дашборд' }
    },
    {
      path: '/domains',
      name: 'Domains',
      component: () => import('../views/Domains.vue'),
      meta: { title: 'Домены' }
    },
    {
      path: '/mailboxes',
      name: 'Mailboxes',
      component: () => import('../views/Mailboxes.vue'),
      meta: { title: 'Почтовые ящики' }
    },
    {
      path: '/login',
      name: 'Login',
      component: () => import('../views/Login.vue'),
      meta: { title: 'Вход' }
    },
    {
      path: '/settings/admins',
      name: 'AdminSettings',
      component: () => import('../views/AdminSettings.vue'),
      meta: { title: 'Администраторы' }
    },
    {
      path: '/change-password',
      name: 'ChangePassword',
      component: () => import('../views/ChangePassword.vue'),
      meta: { title: 'Смена пароля' }
    },
    {
      path: '/logs',
      name: 'Logs',
      component: () => import('../views/Logs.vue'),
      meta: { title: 'Логи системы' }
    },
    {
      path: '/tools/send-mail',
      name: 'SendMail',
      component: () => import('../views/SendMail.vue'),
      meta: { title: 'Отправка почты' }
    },
    {
      path: '/system/queue',
      name: 'MailQueue',
      component: () => import('../views/MailQueue.vue'),
      meta: { title: 'Почтовая очередь' }
    },
    {
      path: '/system/server-logs',
      name: 'ServerLogs',
      component: () => import('../views/ServerLogs.vue'),
      meta: { title: 'Логи сервера' }
    }
  ]
})

router.beforeEach((to) => {
  const publicPages = ['/login']
  const authRequired = !publicPages.includes(to.path)
  const loggedIn = localStorage.getItem('accessToken')

  if (authRequired && !loggedIn) {
    return '/login'
  }

  if (loggedIn) {
    // Проверка принудительной смены пароля
    try {
      const payload = JSON.parse(atob(loggedIn.split('.')[1]))
      if (payload.must_change_password && to.path !== '/change-password') {
        return '/change-password'
      }
    } catch (e) {}
  }

  if (to.path === '/login' && loggedIn) {
    return '/'
  }
})

router.afterEach((to) => {
  const title = to.meta.title
  document.title = title ? `${title} - MailAdmin` : 'MailAdmin'
})

export default router
