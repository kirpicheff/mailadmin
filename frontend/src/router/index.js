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
    }
  ]
})

export default router
