<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import CreatePortfolioModal from '../components/CreatePortfolioModal/CreatePortfolioModal.vue'
import LeftBar from '../components/LeftBar/LeftBar.vue'
import { api } from '../services/api'
import { buildLineChart } from '../utils/chart'

const route = useRoute()
const router = useRouter()
const isCreateModalOpen = ref(false)
const portfolio = ref(null)
const overview = ref(null)
const error = ref('')
const portfolioTitle = computed(() => portfolio.value?.name || `Портфель №${route.params.id}`)
const chart = computed(() => buildLineChart({ portfolio: overview.value?.['p&l_history']?.portfolio, riskFree: overview.value?.['p&l_history']?.risk_free }))
const formatMoney = (value) => Number(value || 0).toLocaleString('ru-RU', { maximumFractionDigits: 2 })
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
onMounted(async () => {
  try {
    const [portfolios, data] = await Promise.all([api.getPortfolios(), api.getOverview(route.params.id)])
    portfolio.value = portfolios.find((item) => String(item.id) === String(route.params.id))
    overview.value = data
  } catch (e) { error.value = e.message }
})
</script>

<template>
  <div class="app-shell overview-shell">
    <LeftBar :portfolio-title="portfolioTitle" :menu-items="menuItems" active-section="overview" @create-portfolio="isCreateModalOpen = true" @navigate="navigateTo" />
    <main class="overview-content">
      <h1>{{ portfolioTitle }}</h1>
      <p v-if="error" class="negative" role="alert">{{ error }}</p>
      <div class="overview-tabs"><button class="active" type="button">Анализ</button><button type="button" @click="navigateTo('structure')">Структура</button></div>
      <section class="performance-layout" aria-label="Динамика портфеля">
        <article class="performance-card">
          <svg viewBox="0 0 1000 460" preserveAspectRatio="none" role="img" aria-label="График портфеля и безрисковой ставки">
            <g class="overview-grid"><line v-for="tick in chart.yTicks" :key="tick.y" x1="52" :y1="tick.y" x2="948" :y2="tick.y"/><text v-for="tick in chart.yTicks" :key="`label-${tick.y}`" x="45" :y="tick.y - 7" text-anchor="end">{{ formatMoney(tick.value) }} USD</text></g>
            <line class="overview-zero" x1="52" :y1="chart.zeroY" x2="948" :y2="chart.zeroY"/>
            <line v-if="chart.historyEndX" class="overview-forecast" :x1="chart.historyEndX" y1="52" :x2="chart.historyEndX" y2="408"/>
            <polyline class="rate-line" :points="chart.series.riskFree?.points"/>
            <polyline class="asset-line" :points="chart.series.portfolio?.points"/>
            <circle v-if="chart.series.portfolio?.last" class="overview-dot" :cx="chart.series.portfolio.last.x" :cy="chart.series.portfolio.last.y" r="6"/>
            <text class="overview-x-label" x="52" y="448">{{ chart.startLabel }}</text><text class="overview-x-label" x="948" y="448" text-anchor="end">{{ chart.endLabel }}</text>
          </svg>
        </article>
        <div class="chart-legend"><span><i class="legend-portfolio"></i>Портфель</span><span><i class="legend-rate"></i>Безрисковая ставка</span></div>
      </section>
      <section class="overview-metrics" aria-label="Показатели портфеля">
        <article><span>Общий P&amp;L</span><strong>{{ Number(overview?.['p&l']?.total_pnl?.[0] || 0).toLocaleString('ru-RU', { maximumFractionDigits: 2 }) }} USD</strong><small>{{ Number(overview?.['p&l']?.total_pnl?.[1] || 0).toLocaleString('ru-RU', { maximumFractionDigits: 2 }) }}%</small></article>
        <article><span>P&amp;L за день</span><strong>{{ Number(overview?.['p&l']?.daily_pnl?.[0] || 0).toLocaleString('ru-RU', { maximumFractionDigits: 2 }) }} USD</strong><small>{{ Number(overview?.['p&l']?.daily_pnl?.[1] || 0).toLocaleString('ru-RU', { maximumFractionDigits: 2 }) }}%</small></article>
        <article><span>Уровень риска</span><strong>{{ overview?.risks?.risks_metrick || 0 }} / 100</strong><small>{{ (overview?.risks?.reasons || []).join(' · ') || 'Нет сигналов' }}</small></article>
      </section>
    </main>
    <CreatePortfolioModal v-if="isCreateModalOpen" @close="isCreateModalOpen = false" @created="handleCreated" />
  </div>
</template>

<style scoped>
.overview-shell{min-height:100vh;background:#1e2638}.overview-content{flex:1;min-width:0;padding:72px clamp(36px,6vw,108px)}.overview-content h1{margin:0 0 44px;color:#fff;font-size:34px}.overview-tabs{display:flex;gap:13px;margin-bottom:30px}.overview-tabs button{width:200px;height:46px;border:0;border-radius:5px;background:#058b5c;color:#fff;font-size:20px;font-weight:800;cursor:pointer}.overview-tabs .active{background:#02603f}.performance-layout{display:flex;align-items:center;gap:42px;max-width:1120px}.performance-card{flex:1;min-width:0;height:458px;overflow:hidden;border:1px solid #344159;border-radius:8px;background:#111d30}.performance-card svg{width:100%;height:100%}.overview-grid line{stroke:#2a3950}.overview-grid text,.overview-x-label{fill:#8795aa;font-size:13px}.overview-zero{stroke:#718096;stroke-width:1.5;stroke-dasharray:7 7}.overview-forecast{stroke:#61a8ff;stroke-width:1.5;stroke-dasharray:5 6}.overview-dot{fill:#acaa5c;stroke:#fff;stroke-width:2}.rate-line,.asset-line{fill:none;stroke-width:4;vector-effect:non-scaling-stroke}.rate-line{stroke:#5fb874}.asset-line{stroke:#acaa5c}.chart-legend{min-width:180px;padding:16px 14px;border:2px solid #3a98a0;border-radius:5px;background:#303c57;color:#fff;font-size:14px;font-weight:800}.chart-legend span{display:flex;align-items:center;gap:9px;margin:10px 0}.chart-legend i{width:16px;height:3px}.legend-portfolio{background:#acaa5c}.legend-rate{background:#5fb874}.overview-metrics{display:grid;grid-template-columns:repeat(2,340px);gap:54px;max-width:1120px;margin-top:52px}.overview-metrics img{display:block;width:340px;max-width:100%;height:auto}@media(max-width:1050px){.performance-layout{display:block}.chart-legend{display:flex;gap:22px;margin-top:14px}}@media(max-width:820px){.overview-shell{display:block}.overview-content{padding:35px 20px}.performance-card{height:320px}.overview-metrics{grid-template-columns:1fr;gap:24px}.overview-tabs button{width:50%}}
</style>
<style scoped>
.overview-metrics{grid-template-columns:repeat(3,1fr);gap:24px}.overview-metrics article{padding:24px;border:1px solid #344159;border-radius:8px;background:#111d30}.overview-metrics span,.overview-metrics small{display:block;color:#9da9bc}.overview-metrics strong{display:block;margin:12px 0;color:#fff;font-size:24px}@media(max-width:820px){.overview-metrics{grid-template-columns:1fr}}
</style>
