<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import LeftBar from '../components/LeftBar/LeftBar.vue'
import CreatePortfolioModal from '../components/CreatePortfolioModal/CreatePortfolioModal.vue'
import { api } from '../services/api'
import { buildLineChart } from '../utils/chart'

const route = useRoute()
const router = useRouter()
const saved = ref(false)
const defaultPortfolioName = `Портфель №${route.params.id || 1}`
const settingsPortfolioName = ref(defaultPortfolioName)
const settingsRiskFreeRate = ref('')
const isCreateModalOpen = ref(false)
const portfolio = ref(null)
const rates = ref([])
const yieldData = ref(null)
const risksData = ref(null)
const error = ref('')

const portfolioTitle = computed(() => settingsPortfolioName.value.trim() || defaultPortfolioName)
const menuItems = [
  { id: 'overview', label: 'Обзор' }, { id: 'structure', label: 'Структура' },
  { id: 'performance', label: 'Доходность' }, { id: 'risks', label: 'Риски&VaR' },
  { id: 'settings', label: 'Настройки' },
]
const page = computed(() => String(route.name))
const format = (value, suffix = '') => `${Number(value || 0).toLocaleString('ru-RU', { maximumFractionDigits: 2 })}${suffix}`
const returns = computed(() => Object.values(yieldData.value?.yield_top_3 || {}).map((row) => ({ symbol: row[0], value: format(row[1], ' USD'), percent: format(row[2], '%'), tone: Number(row[1]) >= 0 ? 'positive' : 'negative' })))
const performanceRows = computed(() => (yieldData.value?.portfolio_structure || []).map((row) => ({ symbol: row[0], price: format(row[1], ' USD'), amount: format(row[2]), cost: format(row[3], ' USD'), percentage: format(row[4], ' %'), pnl: format(row[5], ' USD') })))
const riskRows = computed(() => (risksData.value?.partial_risks || []).map((row) => ({ asset: row[0], expected: `${format(row[1]?.[1], ' USD')} (${format(row[1]?.[0], '%')})`, q90: `${format(row[2]?.[1], ' USD')} (${format(row[2]?.[0], '%')})`, q95: `${format(row[3]?.[1], ' USD')} (${format(row[3]?.[0], '%')})`, q99: `${format(row[4]?.[1], ' USD')} (${format(row[4]?.[0], '%')})`, mcri: format(row[5], '%'), cctr: format(row[6], '%') })))
const pnlChart = computed(() => buildLineChart({ portfolio: yieldData.value?.['p&l']?.portfolio, riskFree: yieldData.value?.['p&l']?.risk_free }))
const fomoChart = computed(() => buildLineChart({ portfolio: yieldData.value?.['p&l']?.portfolio, riskFree: yieldData.value?.['p&l']?.risk_free, fomo: yieldData.value?.['p&l']?.portfolio_FOMO }))
const quantiles = computed(() => {
  const values = risksData.value?.portfolioVaR_quantiles || {}
  const total = Number(risksData.value?.portfolio_value_usd || 0)
  return ['Mu', 'Q90', 'Q95', 'Q99'].map((key) => ({ key, usd: Number(values[key] || 0), percent: total > 0 ? Number(values[key] || 0) / total * 100 : 0 }))
})
function navigateTo(id) {
  const base = `/portfolio/${route.params.id}`
  router.push({ overview: base, structure: `${base}/structure`, performance: `${base}/performance`, risks: `${base}/risks`, settings: `${base}/settings` }[id] || base)
}
async function saveSettings() {
  try {
    await api.updatePortfolio(route.params.id, { name: settingsPortfolioName.value.trim(), risk_free_rate_id: Number(settingsRiskFreeRate.value) })
    portfolio.value.name = settingsPortfolioName.value.trim(); saved.value = true
    setTimeout(() => { saved.value = false }, 1800)
  } catch (e) { error.value = e.message }
}

async function loadData() {
  error.value = ''
  try {
    const [portfolios, riskRates, result] = await Promise.all([api.getPortfolios(), api.getRiskFreeRates(), page.value === 'risks' ? api.getRisks(route.params.id) : api.getYield(route.params.id)])
    portfolio.value = portfolios.find((item) => String(item.id) === String(route.params.id))
    rates.value = riskRates; settingsPortfolioName.value = portfolio.value?.name || defaultPortfolioName
    settingsRiskFreeRate.value = String(portfolio.value?.risk_free_rate_id || riskRates[0]?.id || '')
    if (page.value === 'risks') risksData.value = result; else yieldData.value = result
  } catch (e) { error.value = e.message }
}
onMounted(loadData)
watch(() => route.name, loadData)
</script>

<template>
  <div class="app-shell analytics-shell">
    <LeftBar :portfolio-title="portfolioTitle" :menu-items="menuItems" :active-section="page" @create-portfolio="isCreateModalOpen = true" @navigate="navigateTo" />
    <main class="analytics-content">
      <p v-if="error" class="negative" role="alert">{{ error }}</p>
      <template v-if="page === 'performance'">
        <header class="income-header"><h1>{{ portfolioTitle }}</h1><h2>Анализ P&amp;L</h2></header>

        <section class="income-chart-row" aria-label="Анализ P&L">
          <div class="reference-chart">
            <svg viewBox="0 0 1000 460" preserveAspectRatio="none">
              <g class="dynamic-grid"><line v-for="tick in pnlChart.yTicks" :key="tick.y" x1="52" :y1="tick.y" x2="948" :y2="tick.y"/><text v-for="tick in pnlChart.yTicks" :key="`label-${tick.y}`" x="45" :y="tick.y - 7" text-anchor="end">{{ format(tick.value, ' USD') }}</text></g>
              <line class="zero-line" x1="52" :y1="pnlChart.zeroY" x2="948" :y2="pnlChart.zeroY"/>
              <line v-if="pnlChart.historyEndX" class="forecast-line" :x1="pnlChart.historyEndX" y1="52" :x2="pnlChart.historyEndX" y2="408"/>
              <polyline class="income-rate" :points="pnlChart.series.riskFree?.points" />
              <polyline class="income-pnl" :points="pnlChart.series.portfolio?.points" />
              <circle v-if="pnlChart.series.portfolio?.last" class="income-pnl-dot" :cx="pnlChart.series.portfolio.last.x" :cy="pnlChart.series.portfolio.last.y" r="6"/>
              <text class="x-label" x="52" y="448">{{ pnlChart.startLabel }}</text><text class="x-label" x="948" y="448" text-anchor="end">{{ pnlChart.endLabel }}</text>
            </svg>
          </div>
          <div class="reference-legend"><span><i class="rate-key"></i>безрисковая<br>ставка</span><span><i class="pnl-key"></i>P&amp;L портфеля</span></div>
        </section>

        <section class="income-section"><h2>Топ доходности</h2><div class="income-top"><article v-for="(item,index) in returns" :key="item.symbol"><span>{{ item.symbol }}</span><strong :class="item.tone">{{ item.value }}</strong><small :class="item.tone">({{ item.percent }})</small><b>{{ index + 1 }}</b></article></div></section>

        <section class="income-section income-table-wrap"><table class="income-table"><thead><tr><th>Symbol</th><th>Current price</th><th>Amount</th><th>Cost</th><th>Percentage (%)</th><th>PnL</th></tr></thead><tbody><tr v-for="row in performanceRows" :key="row.symbol"><th>{{ row.symbol }}</th><td>{{ row.price }}</td><td>{{ row.amount }}</td><td>{{ row.cost }}</td><td>{{ row.percentage }}</td><td class="positive">{{ row.pnl }}</td></tr></tbody></table></section>

        <section class="income-section indexes-section"><h2>Sharpes and Sortino ratios</h2><div class="index-values"><article><span>sharpes index</span><strong>{{ format(yieldData?.sharpes_ratio) }}</strong><small>По данным портфеля</small></article><article><span>sortino index</span><strong>{{ format(yieldData?.sortino_ratio) }}</strong><small>По данным портфеля</small></article></div></section>

        <section class="income-section fomo-section"><h2>Оценка упущенной выгоды (FOMO)</h2><div class="income-chart-row"><div class="reference-chart fomo-chart"><svg viewBox="0 0 1000 460" preserveAspectRatio="none"><g class="dynamic-grid"><line v-for="tick in fomoChart.yTicks" :key="tick.y" x1="52" :y1="tick.y" x2="948" :y2="tick.y"/><text v-for="tick in fomoChart.yTicks" :key="`fomo-label-${tick.y}`" x="45" :y="tick.y - 7" text-anchor="end">{{ format(tick.value, ' USD') }}</text></g><line class="zero-line" x1="52" :y1="fomoChart.zeroY" x2="948" :y2="fomoChart.zeroY"/><line v-if="fomoChart.historyEndX" class="forecast-line" :x1="fomoChart.historyEndX" y1="52" :x2="fomoChart.historyEndX" y2="408"/><polyline class="income-rate" :points="fomoChart.series.riskFree?.points"/><polyline class="income-pnl" :points="fomoChart.series.portfolio?.points"/><polyline class="fomo-line" :points="fomoChart.series.fomo?.points"/><circle v-if="fomoChart.series.fomo?.last" class="fomo-dot" :cx="fomoChart.series.fomo.last.x" :cy="fomoChart.series.fomo.last.y" r="6"/><text class="x-label" x="52" y="448">{{ fomoChart.startLabel }}</text><text class="x-label" x="948" y="448" text-anchor="end">{{ fomoChart.endLabel }}</text></svg></div><div class="reference-legend"><span><i class="rate-key"></i>безрисковая ставка</span><span><i class="pnl-key"></i>P&amp;L портфеля</span><span><i class="fomo-key"></i>FOMO P&amp;L портфеля</span></div></div></section>

        <section class="fomo-values"><h2>FOMO P&amp;L и P&amp;L портфеля</h2><div><article><span>Ковариация</span><strong>{{ format(yieldData?.cov_fomo_pnl) }}</strong></article><article><span>Корреляция</span><strong>{{ format(yieldData?.corr_fomo_pnl) }}</strong></article></div></section>
      </template>

      <template v-else-if="page === 'risks'">
        <header class="risk-header"><h1>Анализ рисков</h1><h2>Value at Risk</h2></header>
        <section class="lognormal-card" aria-label="Value at Risk с логнормальным распределением">
          <svg viewBox="0 0 1000 460" preserveAspectRatio="none">
            <rect class="quantile-band expected-band" x="0" y="0" width="250" height="460"/><rect class="quantile-band q90-band" x="250" y="0" width="250" height="460"/><rect class="quantile-band q95-band" x="500" y="0" width="250" height="460"/><rect class="quantile-band q99-band" x="750" y="0" width="250" height="460"/>
            <path class="lognormal-fill" d="M0 455 C24 345 55 190 105 105 C145 36 195 42 235 88 C285 145 302 326 340 410 C365 445 410 449 1000 455 L1000 460 L0 460Z"/>
            <path class="lognormal-line" d="M0 455 C24 345 55 190 105 105 C145 36 195 42 235 88 C285 145 302 326 340 410 C365 445 410 449 1000 455"/>
            <line class="equal-boundary" x1="250" y1="0" x2="250" y2="460"/><line class="equal-boundary" x1="500" y1="0" x2="500" y2="460"/><line class="equal-boundary" x1="750" y1="0" x2="750" y2="460"/>
            <text v-for="(item, index) in quantiles" :key="item.key" class="band-label" :x="125 + index * 250" y="38" text-anchor="middle">{{ item.key === 'Mu' ? 'Мат. ожидание' : item.key }}</text>
            <text v-for="(item, index) in quantiles" :key="`${item.key}-value`" class="quantile-value" :x="125 + index * 250" y="68" text-anchor="middle">{{ format(item.percent, '%') }} · {{ format(item.usd, ' USD') }}</text>
            <text class="axis-label" x="8" y="448">P</text><text class="axis-label" x="972" y="448">σ</text>
          </svg>
          <div class="quantile-legend"><span><i class="expected-key"></i>Мат. ожидание</span><span><i class="q90-key"></i>Q90</span><span><i class="q95-key"></i>Q95</span><span><i class="q99-key"></i>Q99</span></div>
        </section>

        <section class="risk-table-section"><table class="risk-detail-table"><thead><tr><th>Symbol</th><th>Мат. ожидание</th><th>Q90</th><th>Q95</th><th>Q99</th><th>MCRi</th><th>CCTR</th></tr></thead><tbody><tr v-for="row in riskRows" :key="row.asset"><th>{{ row.asset }}</th><td>{{ row.expected }}</td><td class="negative">{{ row.q90 }}</td><td class="negative">{{ row.q95 }}</td><td class="negative">{{ row.q99 }}</td><td>{{ row.mcri }}</td><td>{{ row.cctr }}</td></tr></tbody></table></section>

        <section class="drag-diversification"><div class="drag-block"><h2>Volatility Drag</h2><article><strong>{{ format(risksData?.volatility_drag, '%') }}</strong></article><h3>Средний дрейф</h3></div><div class="diversification-block"><h2>Оценка диверсификации</h2><div><article><strong>{{ format(risksData?.diversification_ratio) }}</strong><span>Diversification Ratio</span></article><article><strong>{{ format(risksData?.effective_number_of_assets) }}</strong><span>Effective Number of Assets</span></article></div></div></section>
      </template>

      <template v-else>
        <header class="settings-page-title"><h1>{{ portfolioTitle }}</h1></header>
        <section class="portfolio-settings-card">
          <label for="settings-portfolio-name">Название портфеля</label>
          <input id="settings-portfolio-name" v-model="settingsPortfolioName" type="text" placeholder="Название портфеля">
          <label for="settings-risk-free-rate">Безрисковая ставка - r</label>
          <select id="settings-risk-free-rate" v-model="settingsRiskFreeRate">
            <option v-for="rate in rates" :key="rate.id" :value="String(rate.id)">{{ rate.name }} ({{ (rate.value * 100).toLocaleString('ru-RU') }}%)</option>
          </select>
          <button type="button" @click="saveSettings">{{ saved ? 'Обновлено' : 'Обновить' }}</button>
        </section>
      </template>
    </main>
    <CreatePortfolioModal v-if="isCreateModalOpen" @close="isCreateModalOpen = false" @created="router.push('/portfolio/' + $event)" />
  </div>
</template>

<style scoped>
.analytics-shell{min-height:100vh;background:#1e2638}.analytics-content{flex:1;min-width:0;max-width:1500px;padding:52px clamp(30px,5vw,80px) 100px}.page-header{display:flex;justify-content:space-between;align-items:flex-end;margin-bottom:42px}.page-header p{margin:0 0 7px;color:#62d3aa;font-size:12px;font-weight:800;letter-spacing:.14em;text-transform:uppercase}.page-header h1{margin:0;color:#fff;font-size:42px}.status{padding:10px 14px;border:1px solid #344159;border-radius:20px;color:#b8c1d0;font-size:13px}.status i{display:inline-block;width:8px;height:8px;margin-right:8px;border-radius:50%;background:#16c784}.summary-grid{display:grid;grid-template-columns:repeat(3,1fr);gap:22px;margin-bottom:30px}.summary-grid article,.panel{border:1px solid #344159;border-radius:10px;background:#111d30}.summary-grid article{padding:24px}.summary-grid span,.panel-head span{display:block;margin-bottom:10px;color:#9da9bc;font-size:13px}.summary-grid strong{display:block;color:#fff;font-size:27px}.summary-grid small{display:block;margin-top:9px;color:#8e9bad}.positive{color:#41ca75!important}.negative{color:#ef6666!important}.panel{padding:28px}.chart-panel-live,.var-panel{margin-bottom:46px}.panel-head{display:flex;align-items:center;justify-content:space-between;margin-bottom:20px}.panel-head strong{color:#fff;font-size:25px}.periods,.confidence{display:flex;gap:7px}.periods button,.confidence button{border:1px solid #344159;border-radius:6px;padding:8px 12px;background:#17253b;color:#aab5c6;cursor:pointer}.periods .active,.confidence .active{border-color:#079963;background:#079963;color:#fff}.chart-panel-live svg,.var-panel svg{width:100%;height:350px}.grid line{stroke:#344159;stroke-width:1}.portfolio,.benchmark{fill:none;stroke-width:4;vector-effect:non-scaling-stroke}.portfolio{stroke:#58c456}.benchmark{stroke:#acaa5c}.legend{display:flex;justify-content:flex-end;gap:22px;color:#aab5c6;font-size:13px}.legend i{display:inline-block;width:18px;height:3px;margin-right:7px}.portfolio-dot{background:#58c456}.benchmark-dot{background:#acaa5c}.section-title{display:flex;align-items:end;justify-content:space-between;margin:0 0 18px}.section-title h2,.settings-card h2{margin:0;color:#fff;font-size:25px}.section-title span{color:#8f9bad;font-size:13px}.return-list{display:grid;gap:10px}.return-list article{display:grid;grid-template-columns:40px 1fr 190px 100px;align-items:center;padding:19px 22px;border:1px solid #344159;border-radius:8px;background:#111d30;color:#fff}.return-list b{display:grid;width:28px;height:28px;place-items:center;border-radius:50%;background:#22314a}.return-list em{text-align:right;font-style:normal;font-weight:800}.confidence{margin-bottom:4px}.risk-summary{margin-bottom:30px}.var-area{fill:rgb(7 153 99 / 18%)}.var-danger{fill:rgb(223 54 54 / 28%)}.var-line{fill:none;stroke:#079963;stroke-width:4}.marker{stroke:#ef6666;stroke-width:2;stroke-dasharray:8 7}.var-panel text{fill:#fff;font-size:16px;font-weight:800}.table-wrap{overflow:hidden;border:1px solid #344159;border-radius:9px}.table-wrap table{width:100%;border-collapse:collapse;background:#111d30;color:#fff}.table-wrap th,.table-wrap td{padding:20px;border-bottom:1px solid #27364d;text-align:left}.table-wrap td{color:#ef7777}.settings-layout{display:grid;grid-template-columns:1fr 1fr;gap:24px}.settings-card{padding:30px}.settings-card h2{margin-bottom:26px}.settings-card label{display:block;margin-bottom:20px;color:#aab5c6;font-size:14px}.settings-card input:not([type=checkbox]),.settings-card select{display:block;width:100%;height:46px;margin-top:8px;padding:0 13px;border:1px solid #344159;border-radius:7px;background:#0d1728;color:#fff}.primary,.security-row button{border:0;border-radius:7px;padding:12px 18px;background:#079963;color:#fff;font-weight:800;cursor:pointer}.switch-row,.security-row{display:flex!important;align-items:center;justify-content:space-between;gap:20px}.switch-row small,.security-row small{display:block;margin-top:5px;color:#7f8ca1}.switch-row input{width:22px;height:22px;accent-color:#079963}.wide{grid-column:1/-1}.security-row{padding:18px 0;border-top:1px solid #27364d}.security-row strong{color:#fff}.security-row button{background:#22314a}@media(max-width:900px){.analytics-shell{display:block}.summary-grid,.settings-layout{grid-template-columns:1fr}.wide{grid-column:auto}.analytics-content{padding:35px 20px}.return-list article{grid-template-columns:35px 1fr}.return-list strong,.return-list em{margin-top:8px;text-align:left}.page-header{align-items:flex-start;gap:20px}.chart-panel-live svg,.var-panel svg{height:260px}}
.income-header{margin-bottom:58px}.income-header h1{margin:0 0 42px;color:#fff;font-size:36px}.income-header h2,.income-section h2,.fomo-values h2{margin:0 0 30px;color:#fff;font-size:28px}.income-chart-row{display:flex;align-items:flex-end;gap:44px;max-width:1140px}.reference-chart{width:min(100%,915px);height:458px;overflow:hidden;border:4px solid #079963;border-radius:4px;background:#c7c0cc}.reference-chart svg{width:100%;height:100%}.income-rate,.income-pnl,.fomo-line{fill:none;stroke-width:4;vector-effect:non-scaling-stroke}.income-rate{stroke:#58c875}.income-pnl{stroke:#adaa4b}.fomo-line{stroke:#080808}.reference-legend{display:grid;gap:20px;min-width:160px;padding:15px 12px;border:2px solid #3699aa;border-radius:5px;background:#303c57;color:#fff;font-size:14px;font-weight:800}.reference-legend span{display:block}.reference-legend i{display:inline-block;width:15px;height:3px;margin-right:6px;vertical-align:middle}.rate-key{background:#58c875}.pnl-key{background:#adaa4b}.fomo-key{background:#080808}.income-section{max-width:1200px;margin-top:54px}.income-top{display:grid;grid-template-columns:repeat(3,235px);gap:54px}.income-top article{position:relative;min-height:185px;padding:18px;border:2px solid #079963;background:#c7c0cc;color:#0b0b0b}.income-top span{display:block;font-size:18px;font-weight:800}.income-top strong,.income-top small{position:relative;z-index:1;display:block;margin-top:38px;font-size:27px}.income-top small{margin-top:3px;font-size:25px;font-weight:800}.income-top b{position:absolute;right:18px;bottom:16px;color:#777;font-size:92px;line-height:1}.income-table-wrap{overflow:auto;padding:0;border-radius:4px;background:#0d1728}.income-table{width:100%;min-width:880px;border-collapse:collapse;color:#fff}.income-table th,.income-table td{padding:20px 24px;border-right:3px solid #1e2a40;border-bottom:3px solid #1e2a40;text-align:center;font-size:16px}.income-table th:last-child,.income-table td:last-child{border-right:0}.income-table tbody tr:last-child th,.income-table tbody tr:last-child td{border-bottom:0}.indexes-section{margin-top:48px}.index-values{display:grid;grid-template-columns:1fr 1fr;gap:72px}.index-values article{min-height:180px;padding:28px;border:2px solid #079963;background:#c7c0cc;color:#111}.index-values span{display:block;font-size:20px;font-weight:800}.index-values strong{display:block;margin-top:22px;color:#a68d00;font-size:42px}.index-values small{display:block;margin-top:6px;color:#8d7900;font-size:16px;font-weight:800}.index-values ul{margin:18px 0 0;padding-left:20px;color:#202020;line-height:1.7}.fomo-section{margin-top:80px}.fomo-chart{height:458px}.fomo-values{max-width:1140px;margin-top:68px;text-align:center}.fomo-values h2{font-size:38px}.fomo-values>div{display:grid;grid-template-columns:1fr 1fr;gap:72px}.fomo-values article{padding:34px;border:2px solid #079963;background:#c7c0cc;color:#111}.fomo-values span{display:block;font-weight:800}.fomo-values strong{display:block;margin-top:20px;color:#e32b2b;font-size:34px}@media(max-width:900px){.income-chart-row{display:block}.reference-chart{height:320px}.reference-legend{margin-top:14px}.income-top,.index-values,.fomo-values>div{grid-template-columns:1fr}.income-top{gap:18px}.income-top article{width:100%}.fomo-values h2{font-size:28px}}
.risk-header{margin-bottom:34px}.risk-header h1{margin:0 0 42px;color:#fff;font-size:36px}.risk-header h2,.drag-diversification h2{margin:0;color:#fff;font-size:27px}.lognormal-card{position:relative;width:min(100%,915px);height:460px;margin:36px 0 74px;border:4px solid #079963;border-radius:4px;background:#c7c0cc}.lognormal-card svg{width:100%;height:100%}.quantile-band{opacity:.16}.expected-band{fill:#58c875}.q90-band{fill:#d8bd35}.q95-band{fill:#db7548}.q99-band{fill:#cf3338}.equal-boundary{stroke:#243146;stroke-width:3;stroke-dasharray:10 8}.lognormal-fill{fill:rgb(15 16 18/4%)}.lognormal-line{fill:none;stroke:#090909;stroke-width:4;vector-effect:non-scaling-stroke}.lognormal-card text{fill:#111;font-size:17px;font-weight:800}.lognormal-card .band-label{font-size:18px}.lognormal-card .axis-label{font-size:24px}.quantile-legend{position:absolute;right:24px;bottom:22px;display:flex;gap:15px;padding:10px 12px;border-radius:5px;background:rgb(255 255 255/68%);color:#111;font-size:13px;font-weight:800}.quantile-legend i{display:inline-block;width:12px;height:3px;margin-right:5px}.expected-key{background:#58c875}.q90-key{background:#d8bd35}.q95-key{background:#db7548}.q99-key{background:#cf3338}.risk-table-section{max-width:1200px;overflow:auto;border-radius:4px;background:#0d1728}.risk-detail-table{width:100%;min-width:1080px;border-collapse:collapse;color:#fff}.risk-detail-table th,.risk-detail-table td{padding:16px 14px;border-right:3px solid #1e2a40;border-bottom:3px solid #1e2a40;text-align:center;font-size:15px;white-space:nowrap}.risk-detail-table th:last-child,.risk-detail-table td:last-child{border-right:0}.risk-detail-table tbody tr:last-child>*{border-bottom:0}.drag-diversification{display:grid;grid-template-columns:.8fr 1.7fr;gap:85px;max-width:1200px;margin-top:85px}.drag-block>article{display:grid;min-height:185px;margin-top:28px;place-items:center;border:2px solid #079963;background:#c7c0cc}.drag-block>article strong{color:#050505;font-size:92px}.drag-block h3{margin:8px 0;text-align:center}.drag-block ul{margin:0;padding-left:26px;color:#fff;font-weight:700;line-height:1.8}.diversification-block>div{display:grid;grid-template-columns:1fr 1fr;gap:54px;margin-top:28px}.diversification-block article{display:flex;min-height:185px;align-items:center;justify-content:center;flex-direction:column;border:2px solid #079963;background:#c7c0cc;color:#111;text-align:center}.diversification-block strong{font-size:58px}.diversification-block span{margin-top:18px;font-size:16px;font-weight:800}@media(max-width:900px){.lognormal-card{height:320px}.quantile-legend{right:10px;bottom:10px;flex-wrap:wrap}.drag-diversification{grid-template-columns:1fr;gap:42px}.diversification-block>div{grid-template-columns:1fr;gap:18px}}
.settings-page-title h1{margin:24px 0 0;color:#fff;font-size:36px}.portfolio-settings-card{width:550px;max-width:100%;margin:230px auto 0;padding:60px 58px 18px;border:4px solid #079963;border-radius:4px;background:#465673}.portfolio-settings-card label{display:block;margin-bottom:8px;color:#aeb9cf;font-size:25px}.portfolio-settings-card input,.portfolio-settings-card select{width:100%;height:50px;margin-bottom:38px;padding:0 16px;border:0;border-radius:8px;outline:none;background:#101c2e;color:#f4f7fb;font-size:22px}.portfolio-settings-card input:focus,.portfolio-settings-card select:focus{outline:2px solid #30c58e}.portfolio-settings-card button{display:block;width:250px;height:65px;margin:-16px auto 0;border:0;border-radius:6px;background:#079963;color:#fff;font-size:25px;cursor:pointer}.portfolio-settings-card button:hover{background:#0aaa70}@media(max-width:900px){.portfolio-settings-card{margin-top:80px;padding:36px 26px 20px}.portfolio-settings-card label{font-size:19px}.portfolio-settings-card input,.portfolio-settings-card select{font-size:17px}}
</style>
<style scoped>
.reference-chart{border:1px solid #344159;background:#111d30}.dynamic-grid line{stroke:#2a3950;stroke-width:1}.dynamic-grid text,.x-label{fill:#8795aa;font-size:13px}.zero-line{stroke:#718096;stroke-width:1.5;stroke-dasharray:7 7}.forecast-line{stroke:#61a8ff;stroke-width:1.5;stroke-dasharray:5 6}.income-pnl-dot{fill:#adaa4b;stroke:#fff;stroke-width:2}.fomo-dot{fill:#61a8ff;stroke:#fff;stroke-width:2}.fomo-line{stroke:#61a8ff}.fomo-key{background:#61a8ff}.lognormal-card .quantile-value{fill:#172033;font-size:14px;font-weight:700}
</style>
