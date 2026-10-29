import { createRouter, createWebHistory } from 'vue-router'

import LoginView from '../views/LoginView.vue'
import RegisterView from '../views/RegisterView.vue'
import ResetPasswordView from '../views/ResetPasswordView.vue'
import CreatePortfolioView from '../views/CreatePortfolioView.vue'
import EmptyPortfolioView from '../views/EmptyPortfolioView.vue'
import OverviewView from '../views/OverviewView.vue'
import StructureView from '../views/StructureView.vue'
import InteractiveAnalyticsView from '../views/InteractiveAnalyticsView.vue'

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
  { path: '/', component: OverviewView },
  { path: '/:pathMatch(.*)*', redirect: '/' },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to) => {
  const isAuthenticated = localStorage.getItem('isAuthenticated') === 'true'
  const portfolios = JSON.parse(localStorage.getItem('portfolios') || '[]')

  if (to.meta.requiresAuth && !isAuthenticated) {
    return '/login'
  }

  if (to.path === '/') {
    if (!portfolios.length) return '/empty'
    return '/portfolio/' + portfolios[0]
  }

  if (to.name === undefined && to.path.startsWith('/portfolio/')) {
    const portfolioId = to.params.id
    if (!portfolios.includes(portfolioId)) {
      return portfolios.length ? '/portfolio/' + portfolios[0] : '/empty'
    }
  }
})

export default router
