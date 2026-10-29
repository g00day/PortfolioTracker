<script setup>
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import LeftBar from '../components/LeftBar/LeftBar.vue'
import CreatePortfolioModal from '../components/CreatePortfolioModal/CreatePortfolioModal.vue'
import performance1 from '../assets/performance-1.svg'
import performance2 from '../assets/performance-2.svg'
import performance3 from '../assets/performance-3.svg'
import risksVar1 from '../assets/risks-var-1.svg'
import risksVar2 from '../assets/risks-var-2.svg'
import settingsPage from '../assets/settings-page.svg'

const route = useRoute()
const router = useRouter()
const isCreateModalOpen = ref(false)
const portfolioTitle = computed(() => {
  const id = String(route.params.id || '1')
  return id === '1' || id.length > 4 ? 'Портфель №1' : `Портфель №${id}`
})
const menuItems = [
  { id: 'overview', label: 'Обзор' },
  { id: 'structure', label: 'Структура' },
  { id: 'performance', label: 'Доходность' },
  { id: 'risks', label: 'Риски&VaR' },
  { id: 'settings', label: 'Настройки' },
]
const activeSection = computed(() => String(route.name || 'overview'))
const pageParts = computed(() => {
  if (route.name === 'performance') return [performance1, performance2, performance3]
  if (route.name === 'risks') return [risksVar1, risksVar2]
  return [settingsPage]
})

function navigateTo(sectionId) {
  const base = `/portfolio/${route.params.id}`
  const destinations = { overview: base, structure: `${base}/structure`, performance: `${base}/performance`, risks: `${base}/risks`, settings: `${base}/settings` }
  router.push(destinations[sectionId] || base)
}
</script>

<template>
  <div class="app-shell reference-shell">
    <LeftBar :portfolio-title="portfolioTitle" :menu-items="menuItems" :active-section="activeSection" @create-portfolio="isCreateModalOpen = true" @navigate="navigateTo" />
    <main class="reference-page" :aria-label="menuItems.find((item) => item.id === activeSection)?.label">
      <img v-for="(part, index) in pageParts" :key="part" :src="part" :alt="`Часть ${index + 1}`">
    </main>
    <CreatePortfolioModal v-if="isCreateModalOpen" @close="isCreateModalOpen = false" @created="router.push('/portfolio/' + $event)" />
  </div>
</template>

<style scoped>
.reference-shell{position:relative;align-items:flex-start;overflow:visible;background:#1e2638}.reference-page{width:calc(100vw - 260px);min-width:0;overflow:hidden;background:#1e2638}.reference-page img{position:relative;left:-260px;display:block;width:100vw;height:auto;margin:0;padding:0}@media(max-width:820px){.reference-shell{display:block}.reference-page{width:100%}.reference-page img{left:0;width:100%}}
</style>
