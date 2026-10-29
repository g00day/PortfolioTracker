<script setup>
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import CreatePortfolioModal from '../components/CreatePortfolioModal/CreatePortfolioModal.vue'
import LeftBar from '../components/LeftBar/LeftBar.vue'
import lowerLeftChart from '../assets/lower-left-chart.svg'
import lowerRightChart from '../assets/lower-right-chart.svg'

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

function navigateTo(sectionId) {
  const base = `/portfolio/${route.params.id}`
  const destinations = { overview: base, structure: `${base}/structure`, performance: `${base}/performance`, risks: `${base}/risks`, settings: `${base}/settings` }
  router.push(destinations[sectionId] || base)
}
function handleCreated(id) {
  isCreateModalOpen.value = false
  router.push('/portfolio/' + id)
}
</script>

<template>
  <div class="app-shell overview-shell">
    <LeftBar :portfolio-title="portfolioTitle" :menu-items="menuItems" active-section="overview" @create-portfolio="isCreateModalOpen = true" @navigate="navigateTo" />
    <main class="overview-content">
      <h1>{{ portfolioTitle }}</h1>
      <div class="overview-tabs"><button class="active" type="button">Анализ</button><button type="button" @click="navigateTo('structure')">Структура</button></div>
      <section class="performance-layout" aria-label="Динамика портфеля">
        <article class="performance-card">
          <svg viewBox="0 0 1000 460" preserveAspectRatio="none" role="img" aria-label="График портфеля и безрисковой ставки">
            <g class="overview-grid"><line x1="0" y1="92" x2="1000" y2="92"/><line x1="0" y1="184" x2="1000" y2="184"/><line x1="0" y1="276" x2="1000" y2="276"/><line x1="0" y1="368" x2="1000" y2="368"/></g>
            <polyline class="rate-line" points="0,456 120,414 245,372 375,319 500,260 625,205 750,143 875,80 1000,18"/>
            <polyline class="asset-line" points="0,456 55,395 130,409 175,420 240,353 390,378 525,309 610,207 710,167 760,205 835,216 925,153 1000,119"/>
          </svg>
        </article>
        <div class="chart-legend"><span><i class="legend-portfolio"></i>Портфель</span><span><i class="legend-rate"></i>Безрисковая ставка</span></div>
      </section>
      <section class="overview-metrics" aria-label="Показатели портфеля">
        <img :src="lowerLeftChart" alt="Левый график показателей портфеля">
        <img :src="lowerRightChart" alt="Правый график показателей портфеля">
      </section>
    </main>
    <CreatePortfolioModal v-if="isCreateModalOpen" @close="isCreateModalOpen = false" @created="handleCreated" />
  </div>
</template>

<style scoped>
.overview-shell{min-height:100vh;background:#1e2638}.overview-content{flex:1;min-width:0;padding:72px clamp(36px,6vw,108px)}.overview-content h1{margin:0 0 44px;color:#fff;font-size:34px}.overview-tabs{display:flex;gap:13px;margin-bottom:30px}.overview-tabs button{width:200px;height:46px;border:0;border-radius:5px;background:#058b5c;color:#fff;font-size:20px;font-weight:800;cursor:pointer}.overview-tabs .active{background:#02603f}.performance-layout{display:flex;align-items:center;gap:42px;max-width:1120px}.performance-card{flex:1;min-width:0;height:458px;overflow:hidden;border:4px solid #058b5c;border-radius:5px;background:#c7c0cc}.performance-card svg{width:100%;height:100%}.overview-grid line{stroke:#9b96a0;opacity:.35}.rate-line,.asset-line{fill:none;stroke-width:4;vector-effect:non-scaling-stroke}.rate-line{stroke:#5fb874}.asset-line{stroke:#acaa5c}.chart-legend{min-width:180px;padding:16px 14px;border:2px solid #3a98a0;border-radius:5px;background:#303c57;color:#fff;font-size:14px;font-weight:800}.chart-legend span{display:flex;align-items:center;gap:9px;margin:10px 0}.chart-legend i{width:16px;height:3px}.legend-portfolio{background:#acaa5c}.legend-rate{background:#5fb874}.overview-metrics{display:grid;grid-template-columns:repeat(2,340px);gap:54px;max-width:1120px;margin-top:52px}.overview-metrics img{display:block;width:340px;max-width:100%;height:auto}@media(max-width:1050px){.performance-layout{display:block}.chart-legend{display:flex;gap:22px;margin-top:14px}}@media(max-width:820px){.overview-shell{display:block}.overview-content{padding:35px 20px}.performance-card{height:320px}.overview-metrics{grid-template-columns:1fr;gap:24px}.overview-tabs button{width:50%}}
</style>
