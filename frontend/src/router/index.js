import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'Dashboard',
      component: () => import('../views/Dashboard.vue')
    },
    {
      path: '/domains',
      name: 'Domains',
      component: () => import('../views/Domains.vue')
    },
    {
      path: '/mailboxes',
      name: 'Mailboxes',
      component: () => import('../views/Mailboxes.vue')
    },
    {
      path: '/login',
      name: 'Login',
      component: () => import('../views/Login.vue')
    },
    {
      path: '/settings/admins',
      name: 'AdminSettings',
      component: () => import('../views/AdminSettings.vue')
    },
    {
      path: '/change-password',
      name: 'ChangePassword',
      component: () => import('../views/ChangePassword.vue')
    }
  ]
})

router.beforeEach((to, from, next) => {
  const publicPages = ['/login']
  const authRequired = !publicPages.includes(to.path)
  const loggedIn = localStorage.getItem('accessToken')

  if (authRequired && !loggedIn) {
    return next('/login')
  }

  if (loggedIn) {
    // Проверка принудительной смены пароля
    try {
      const payload = JSON.parse(atob(loggedIn.split('.')[1]))
      if (payload.must_change_password && to.path !== '/change-password') {
        return next('/change-password')
      }
    } catch (e) {}
  }

  if (to.path === '/login' && loggedIn) {
    return next('/')
  }

  next()
})

export default router
