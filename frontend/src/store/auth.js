import { defineStore } from 'pinia'
import api from '@/api/axios'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: JSON.parse(localStorage.getItem('user')) || null,
    accessToken: localStorage.getItem('accessToken') || null
  }),

  getters: {
    isAuthenticated: (state) => !!state.accessToken,
    isSuperAdmin: (state) => state.user?.superadmin || false,
    mustChangePassword: (state) => {
      if (!state.accessToken) return false
      try {
        const payload = JSON.parse(atob(state.accessToken.split('.')[1]))
        return payload.must_change_password || false
      } catch (e) {
        return false
      }
    }
  },

  actions: {
    async login(username, password) {
      try {
        const { data } = await api.post('/auth/login', { username, password })
        this.accessToken = data.access_token
        this.user = data.user
        
        localStorage.setItem('accessToken', data.access_token)
        localStorage.setItem('user', JSON.stringify(data.user))
        
        return true
      } catch (error) {
        throw error
      }
    },

    async logout() {
      try {
        await api.post('/auth/logout')
      } catch (e) {
        console.error('Logout error', e)
      } finally {
        this.user = null
        this.accessToken = null
        localStorage.removeItem('accessToken')
        localStorage.removeItem('user')
      }
    }
  }
})
