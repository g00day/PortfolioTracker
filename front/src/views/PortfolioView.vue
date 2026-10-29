<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import CreatePortfolioModal from '../components/CreatePortfolioModal/CreatePortfolioModal.vue'
import LeftBar from '../components/LeftBar/LeftBar.vue'

const route = useRoute()
const router = useRouter()
const isCreateModalOpen = ref(false)
const activeSection = ref('risk')
const selectedPeriod = ref('D')
let sectionObserver

const portfolioTitle = computed(() => {
  const id = String(route.params.id || '1')
  return id === '1' || id.length > 4 ? 'Портфель №1' : `Портфель №${id}`
})

const menuItems = [
  { id: 'risk', label: 'Анализ рисков' },
  { id: 'volatility', label: 'Volatility Drag' },
  { id: 'pnl', label: 'Анализ P&L' },
  { id: 'structure', label: 'Структура' },
]

const periods = ['D', 'M', 'Y']
const pnlChartData = {
  D: {
    portfolio: '0,340 80,286 165,302 245,254 330,272 415,238 500,252 590,205 680,169 760,108 840,76 925,96 1000,88',
    benchmark: '0,340 110,316 220,278 335,226 450,174 565,124 680,73 800,18 1000,-70',
  },
  M: {
    portfolio: '0,340 90,310 180,265 270,284 360,221 450,235 540,182 630,204 720,138 810,105 900,56 1000,72',
    benchmark: '0,340 110,307 220,270 335,231 450,192 565,151 680,111 800,70 1000,8',
  },
  Y: {
    portfolio: '0,340 95,295 190,310 285,238 380,205 475,224 570,170 665,132 760,88 855,115 940,39 1000,22',
    benchmark: '0,340 110,321 220,288 335,250 450,203 565,159 680,113 800,69 1000,-12',
  },
}
const activePnlData = computed(() => pnlChartData[selectedPeriod.value])

const riskRows = [
  { symbol: 'BTC', values: ['−26k USD (−6%)', '−34k USD (−9,3%)', '−48k USD (−15,6%)'], tone: 'negative' },
  { symbol: 'USD-Tres', values: ['−2,1k USD (−0,8%)', '−3,8k USD (−1,4%)', '−5,2k USD (−2,1%)'], tone: 'negative' },
  { symbol: 'XAUT', values: ['−6,4k USD (−2,7%)', '−9,1k USD (−4,2%)', '−13k USD (−6,8%)'], tone: 'negative' },
]

const volatilityMetrics = [
  {
    label: 'Volatility Drag',
    value: '11,5%',
    caption: 'Средний дрейф',
    tone: 'neutral',
    notes: ['Требуется ребалансировка', 'Недостаточная диверсификация'],
  },
  {
    label: 'Оценка риска',
    value: '42%',
    caption: 'Выше среднего',
    tone: 'danger',
    notes: ['Высокая доля волатильных активов', 'VaR превышает целевой уровень'],
  },
  {
    label: 'Диверсификация',
    value: '63%',
    caption: 'Умеренная',
    tone: 'positive',
    notes: ['Три класса активов', 'BTC формирует основной риск'],
  },
]

const topReturns = [
  { symbol: 'BTC', value: '+102k USD', percent: '(+48%)', place: 1, tone: 'positive' },
  { symbol: 'USD-Tres', value: '+80k USD', percent: '(+38%)', place: 2, tone: 'positive' },
  { symbol: 'XAUT', value: '−10k USD', percent: '(−8%)', place: 3, tone: 'negative' },
]

function handleCreated(portfolioId) {
  isCreateModalOpen.value = false
  router.push('/portfolio/' + portfolioId)
}

function navigateTo(sectionId) {
  if (sectionId === 'structure') {
    router.push(`/portfolio/${route.params.id}/structure`)
    return
  }

  const section = document.getElementById(sectionId)
  if (!section) return

  activeSection.value = sectionId
  section.scrollIntoView({ behavior: 'smooth', block: 'start' })
  window.history.replaceState({ ...window.history.state }, '', `${route.path}#${sectionId}`)
}

onMounted(async () => {
  await nextTick()

  sectionObserver = new IntersectionObserver(
    (entries) => {
      const visibleEntry = entries
        .filter((entry) => entry.isIntersecting)
        .sort((a, b) => b.intersectionRatio - a.intersectionRatio)[0]

      if (visibleEntry) activeSection.value = visibleEntry.target.id
    },
    { rootMargin: '-15% 0px -60% 0px', threshold: [0, 0.1, 0.25] },
  )

  ;['risk', 'volatility', 'pnl'].forEach((id) => {
    const section = document.getElementById(id)
    if (section) sectionObserver.observe(section)
  })

  const requestedSection = route.hash.replace('#', '')
  if (['risk', 'volatility', 'pnl'].includes(requestedSection)) {
    requestAnimationFrame(() => navigateTo(requestedSection))
  }
})

onBeforeUnmount(() => sectionObserver?.disconnect())
</script>

<template>
  <div class="app-shell portfolio-shell">
    <LeftBar
      :portfolio-title="portfolioTitle"
      :menu-items="menuItems"
      :active-section="activeSection"
      @create-portfolio="isCreateModalOpen = true"
      @navigate="navigateTo"
    />

    <main class="portfolio-content">
      <header class="portfolio-header">
        <div>
          <p class="eyebrow">Обзор портфеля</p>
          <h1>{{ portfolioTitle }}</h1>
        </div>
        <div class="portfolio-status">
          <span></span>
          Данные актуальны
        </div>
      </header>

      <section id="risk" class="portfolio-section risk-section" aria-labelledby="risk-title">
        <div class="section-heading">
          <div>
            <p class="section-kicker">Риски</p>
            <h2 id="risk-title">Анализ рисков</h2>
          </div>
          <p>Потенциальные потери портфеля при разных доверительных уровнях.</p>
        </div>

        <article class="analysis-card var-card">
          <div class="card-header">
            <div>
              <span class="card-label">Value at Risk</span>
              <strong>−34 000 USD</strong>
            </div>
            <div class="confidence-pill">95% confidence</div>
          </div>

          <div class="var-chart-wrap">
            <svg class="var-chart" viewBox="0 0 1120 410" role="img" aria-label="Распределение вероятных доходностей и зона Value at Risk">
              <defs>
                <linearGradient id="areaFill" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stop-color="#079963" stop-opacity=".22" />
                  <stop offset="100%" stop-color="#079963" stop-opacity=".02" />
                </linearGradient>
                <linearGradient id="riskFill" x1="0" y1="0" x2="1" y2="0">
                  <stop offset="0%" stop-color="#df3636" stop-opacity=".34" />
                  <stop offset="100%" stop-color="#df3636" stop-opacity=".06" />
                </linearGradient>
              </defs>
              <g class="chart-grid-lines">
                <line x1="36" y1="92" x2="1085" y2="92" />
                <line x1="36" y1="180" x2="1085" y2="180" />
                <line x1="36" y1="268" x2="1085" y2="268" />
                <line x1="36" y1="356" x2="1085" y2="356" />
              </g>
              <path class="var-area" d="M36 356 C72 252 112 116 225 82 C334 49 420 103 475 235 C510 318 540 342 603 346 C750 351 906 354 1085 356 L1085 356 L36 356 Z" />
              <path class="var-risk-area" d="M36 356 C72 252 112 116 225 82 C255 73 283 70 310 73 L310 356 Z" />
              <path class="var-line" d="M36 356 C72 252 112 116 225 82 C334 49 420 103 475 235 C510 318 540 342 603 346 C750 351 906 354 1085 356" />
              <line class="var-marker" x1="310" y1="73" x2="310" y2="356" />
              <circle class="var-dot" cx="310" cy="73" r="7" />
              <text x="324" y="104" class="chart-callout">VaR 95%</text>
              <text x="36" y="392" class="axis-text">Потери</text>
              <text x="1007" y="392" class="axis-text">Доходность</text>
            </svg>
          </div>
        </article>

        <div class="risk-table-wrap">
          <table class="risk-table">
            <thead>
              <tr>
                <th>Актив</th>
                <th>68,5%</th>
                <th>95%</th>
                <th>99,7%</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="row in riskRows" :key="row.symbol">
                <th>{{ row.symbol }}</th>
                <td v-for="value in row.values" :key="value" :class="row.tone">{{ value }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <section id="volatility" class="portfolio-section volatility-section" aria-labelledby="volatility-title">
        <div class="section-heading">
          <div>
            <p class="section-kicker">Волатильность</p>
            <h2 id="volatility-title">Volatility Drag</h2>
          </div>
          <p>Влияние колебаний рынка на геометрическую доходность портфеля.</p>
        </div>

        <div class="volatility-grid">
          <article v-for="metric in volatilityMetrics" :key="metric.label" class="volatility-card">
            <span class="card-label">{{ metric.label }}</span>
            <div class="volatility-value" :class="metric.tone">{{ metric.value }}</div>
            <strong class="volatility-caption">{{ metric.caption }}</strong>
            <ul>
              <li v-for="note in metric.notes" :key="note">{{ note }}</li>
            </ul>
          </article>
        </div>
      </section>

      <section id="pnl" class="portfolio-section pnl-section" aria-labelledby="pnl-title">
        <div class="section-heading">
          <div>
            <p class="section-kicker">Доходность</p>
            <h2 id="pnl-title">Анализ P&amp;L</h2>
          </div>
          <p>Динамика результата портфеля относительно безрисковой ставки.</p>
        </div>

        <article class="analysis-card pnl-chart-card">
          <div class="card-header pnl-card-header">
            <div>
              <span class="card-label">Общий P&amp;L</span>
              <strong class="positive">+172 000 USD</strong>
            </div>
            <div class="chart-legend" aria-label="Легенда графика">
              <span><i class="benchmark-dot"></i>Безрисковая ставка</span>
              <span><i class="portfolio-dot"></i>P&amp;L портфеля</span>
            </div>
          </div>

          <div class="pnl-chart-wrap">
            <svg class="pnl-chart" viewBox="0 0 1000 380" preserveAspectRatio="none" role="img" aria-label="Динамика P&L портфеля">
              <g class="chart-grid-lines">
                <line x1="0" y1="60" x2="1000" y2="60" />
                <line x1="0" y1="130" x2="1000" y2="130" />
                <line x1="0" y1="200" x2="1000" y2="200" />
                <line x1="0" y1="270" x2="1000" y2="270" />
                <line x1="0" y1="340" x2="1000" y2="340" />
              </g>
              <polyline :points="activePnlData.benchmark" class="benchmark-line" />
              <polyline :points="activePnlData.portfolio" class="portfolio-line" />
            </svg>
          </div>

          <div class="period-switch" aria-label="Период графика">
            <button
              v-for="period in periods"
              :key="period"
              type="button"
              :class="{ selected: selectedPeriod === period }"
              @click="selectedPeriod = period"
            >
              {{ period }}
            </button>
          </div>
        </article>

        <div class="top-heading">
          <h3>Топ доходности</h3>
          <span>За выбранный период</span>
        </div>
        <div class="return-grid">
          <article v-for="item in topReturns" :key="item.symbol" class="return-card">
            <div>
              <span class="return-symbol">{{ item.symbol }}</span>
              <strong :class="item.tone">{{ item.value }}</strong>
              <small :class="item.tone">{{ item.percent }}</small>
            </div>
            <span class="return-place">{{ item.place }}</span>
          </article>
        </div>
      </section>
    </main>

    <CreatePortfolioModal
      v-if="isCreateModalOpen"
      @close="isCreateModalOpen = false"
      @created="handleCreated"
    />
  </div>
</template>

<style src="./PortfolioView.css" scoped></style>
