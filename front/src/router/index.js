import { createRouter, createWebHistory } from 'vue-router'

import LoginView from '../views/LoginView.vue'
import RegisterView from '../views/RegisterView.vue'
import ResetPasswordView from '../views/ResetPasswordView.vue'
import CreatePortfolioView from '../views/CreatePortfolioView.vue'
import EmptyPortfolioView from '../views/EmptyPortfolioView.vue'
import OverviewView from '../views/OverviewView.vue'
import StructureView from '../views/StructureView.vue'
import InteractiveAnalyticsView from '../views/InteractiveAnalyticsView.vue'
import { api, clearSession, isAuthenticated } from '../services/api'

const routes = [
  { path: '/login', component: LoginView },
  { path: '/register', component: RegisterView },
  { path: '/reset-password', component: ResetPasswordView },
  { path: '/create-portfolio', component: CreatePortfolioView, meta: { requiresAuth: true } },
  { path: '/empty', component: EmptyPortfolioView, meta: { requiresAuth: true } },
  { path: '/portfolio/:id/structure', component: StructureView, meta: { requiresAuth: true } },
  { path: '/portfolio/:id/performance', name: 'performance', component: InteractiveAnalyticsView, meta: { requiresAuth: true } },
  { path: '/portfolio/:id/risks', name: 'risks', component: InteractiveAnalyticsView, meta: { requiresAuth: true } },
  { path: '/portfolio/:id/settings', name: 'settings', component: InteractiveAnalyticsView, meta: { requiresAuth: true } },
  { path: '/portfolio/:id', component: OverviewView, meta: { requiresAuth: true } },
  { path: '/', component: OverviewView, meta: { requiresAuth: true } },
  { path: '/:pathMatch(.*)*', redirect: '/' },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

window.addEventListener('auth:unauthorized', () => router.replace('/login'))

router.beforeEach(async (to) => {
  const authenticated = isAuthenticated()
  if (!authenticated) clearSession()
  if (to.meta.requiresAuth && !authenticated) {
    return '/login'
  }

  if ((to.path === '/login' || to.path === '/register') && authenticated) return '/'

  if (authenticated && (to.path === '/' || to.path === '/empty')) {
    try {
      const portfolios = await api.getPortfolios()
      if (to.path === '/' && portfolios.length) return `/portfolio/${portfolios[0].id}`
      if (to.path === '/' && !portfolios.length) return '/empty'
      if (to.path === '/empty' && portfolios.length) return `/portfolio/${portfolios[0].id}`
    } catch (error) {
      if (error.status === 401) return '/login'
    }
  }
})

export default router
