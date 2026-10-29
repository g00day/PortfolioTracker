const API_URL = (import.meta.env.VITE_API_URL || 'http://localhost:8080').replace(/\/$/, '')
const TOKEN_KEY = 'portfolio_tracker_token'

export class ApiError extends Error {
  constructor(message, status) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

export function getToken() {
  return localStorage.getItem(TOKEN_KEY)
}

function readClaims(token) {
  try {
    return JSON.parse(atob(token.split('.')[1].replace(/-/g, '+').replace(/_/g, '/')))
  } catch {
    return null
  }
}

export function getCurrentUser() {
  const claims = readClaims(getToken() || '')
  return claims ? { id: claims.sub, login: claims.login, expiresAt: claims.exp } : null
}

export function isAuthenticated() {
  const user = getCurrentUser()
  return Boolean(user && (!user.expiresAt || user.expiresAt * 1000 > Date.now()))
}

export function clearSession() {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem('username')
}

async function request(path, options = {}) {
  const headers = new Headers(options.headers)
  const token = getToken()
  if (token) headers.set('Authorization', `Bearer ${token}`)
  if (options.body && !headers.has('Content-Type')) headers.set('Content-Type', 'application/json')

  let response
  try {
    response = await fetch(`${API_URL}${path}`, { ...options, headers })
  } catch {
    throw new ApiError('Не удалось подключиться к серверу. Проверьте, что backend запущен.', 0)
  }

  const text = await response.text()
  let data = null
  try { data = text ? JSON.parse(text) : null } catch { data = text }

  if (!response.ok) {
    if (response.status === 401 && token) {
      clearSession()
      window.dispatchEvent(new Event('auth:unauthorized'))
    }
    throw new ApiError(data?.error || data?.message || `Ошибка сервера (${response.status})`, response.status)
  }
  return data
}

export const api = {
  async login(login, password) {
    const result = await request('/auth/login', { method: 'POST', body: JSON.stringify({ login, password }) })
    localStorage.setItem(TOKEN_KEY, result.token)
    localStorage.setItem('username', login)
    return result
  },
  register: (login, email, password) => request('/auth/register', { method: 'POST', body: JSON.stringify({ login, email, password }) }),
  requestPasswordReset: (email) => request('/auth/request-password-reset', { method: 'POST', body: JSON.stringify({ email }) }),
  verifyOtc: (email, code) => request('/auth/verify-otc', { method: 'POST', body: JSON.stringify({ email, code }) }),
  resetPassword: (email, code, newPassword) => request('/auth/reset-password', { method: 'POST', body: JSON.stringify({ email, code, new_password: newPassword }) }),
  getPortfolios: () => request('/portfolios/'),
  createPortfolio: (payload) => request('/portfolios/', { method: 'POST', body: JSON.stringify(payload) }),
  updatePortfolio: (id, payload) => request(`/portfolios/${id}`, { method: 'PUT', body: JSON.stringify(payload) }),
  getRiskFreeRates: () => request('/risk-free-rates/'),
  createRiskFreeRate: (payload) => request('/risk-free-rates/', { method: 'POST', body: JSON.stringify(payload) }),
  getOverview: (id) => request(`/portfolios/${id}/overview`),
  getStructure: (id) => request(`/portfolios/${id}/structure`),
  getYield: (id) => request(`/portfolios/${id}/yield`),
  getRisks: (id) => request(`/portfolios/${id}/risks`),
  addPortfolioItem: (id, payload) => request(`/portfolios/${id}/items`, { method: 'POST', body: JSON.stringify(payload) }),
  addTransaction: (id, payload) => request(`/portfolios/${id}/transactions`, { method: 'POST', body: JSON.stringify(payload) }),
}

export function extractCreatedId(message) {
  return String(message?.match(/\b(\d+)\b/)?.[1] || '')
}
