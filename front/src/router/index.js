import { createRouter, createWebHistory } from 'vue-router'

import LoginView from '../views/LoginView.vue'
import CreatePortfolioView from '../views/CreatePortfolioView.vue'
import EmptyPortfolioView from '../views/EmptyPortfolioView.vue'
import PortfolioView from '../views/PortfolioView.vue'

const routes = [
  { path: '/login', component: LoginView },
  { path: '/create-portfolio', component: CreatePortfolioView, meta: { requiresAuth: true } },
  { path: '/empty', component: EmptyPortfolioView, meta: { requiresAuth: true } },
  { path: '/portfolio/:id', component: PortfolioView, meta: { requiresAuth: true } },
  { path: '/', component: PortfolioView },
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
    return `/portfolio/${portfolios[0]}`
  }

  if (to.name === undefined && to.path.startsWith('/portfolio/')) {
    const portfolioId = to.params.id
    if (!portfolios.includes(portfolioId)) {
      return portfolios.length ? `/portfolio/${portfolios[0]}` : '/empty'
    }
  }
})

export default router

