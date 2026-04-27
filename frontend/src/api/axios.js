import axios from 'axios'

const instance = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8080/api',
  withCredentials: true // Для отправки куки с Refresh Token
})

// Интерцептор для добавления Access Token в заголовки
instance.interceptors.request.use(config => {
  const token = localStorage.getItem('accessToken')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Интерцептор для обработки 401 и обновления токена
instance.interceptors.response.use(
  response => response,
  async error => {
    const originalRequest = error.config
    const isLoginRequest = originalRequest.url.includes('/auth/login')
    const isRefreshRequest = originalRequest.url.includes('/auth/refresh')

    if (error.response?.status === 401 && !originalRequest._retry && !isLoginRequest && !isRefreshRequest) {
      originalRequest._retry = true

      try {
        const { data } = await axios.post(
          `${instance.defaults.baseURL}/auth/refresh`,
          {},
          { withCredentials: true }
        )

        localStorage.setItem('accessToken', data.access_token)
        originalRequest.headers.Authorization = `Bearer ${data.access_token}`

        return instance(originalRequest)
      } catch (refreshError) {
        // Если обновление не удалось - разлогиниваем
        localStorage.removeItem('accessToken')
        window.location.href = '/login'
        return Promise.reject(refreshError)
      }
    }

    return Promise.reject(error)
  }
)

export default instance
