<script setup>
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import CreatePortfolioModal from '../components/CreatePortfolioModal/CreatePortfolioModal.vue'
import LeftBar from '../components/LeftBar/LeftBar.vue'
import riskFreeIcon from '../assets/risk-free.svg'

const route = useRoute()
const router = useRouter()

const isCreateModalOpen = ref(false)
const isAssetModalOpen = ref(false)
const assetType = ref('risky')
const symbol = ref('')
const amount = ref('')
const riskFreeValue = ref('')
const riskFreeRate = ref('US Treasuries (+5,10%)')

const assets = ref([
  {
    id: 1,
    symbol: 'BTC',
    currentPrice: '82543 USD',
    amount: '2.2',
    cost: '165087 USD',
    percentage: '6.1 %',
    pnl: '14000 USD',
    isRiskFree: false,
  },
  {
    id: 2,
    symbol: 'USD-Tres',
    currentPrice: 'US Treasuries (+5,10%)',
    amount: '—',
    cost: '25000 USD',
    percentage: '0.9 %',
    pnl: '14000 USD',
    isRiskFree: true,
  },
])

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

function goToAnalysis() {
  router.push('/portfolio/' + route.params.id)
}

function handleSidebarNavigation(sectionId) {
  if (sectionId === 'structure') return
  const base = `/portfolio/${route.params.id}`
  const destinations = { overview: base, performance: `${base}/performance`, risks: `${base}/risks`, settings: `${base}/settings` }
  router.push(destinations[sectionId] || base)
}

function closeAssetModal() {
  isAssetModalOpen.value = false
  symbol.value = ''
  amount.value = ''
  riskFreeValue.value = ''
  assetType.value = 'risky'
}

function addAsset() {
  const isRiskFree = assetType.value === 'riskFree'
  const assetSymbol = isRiskFree ? riskFreeRate.value.split(' (')[0] : symbol.value.trim().toUpperCase()
  if (!assetSymbol) return
  const units = Number(amount.value) || 0
  const value = Number(riskFreeValue.value) || 0

  if ((!isRiskFree && !units) || (isRiskFree && !value)) return

  assets.value.push({
    id: Date.now(),
    symbol: assetSymbol,
    currentPrice: isRiskFree ? riskFreeRate.value : '—',
    amount: isRiskFree ? '—' : String(units || 0),
    cost: isRiskFree ? value.toLocaleString('ru-RU') + ' USD' : '—',
    percentage: '0 %',
    pnl: '0 USD',
    isRiskFree,
  })

  closeAssetModal()
}
</script>

<template>
  <div class="app-shell structure-shell">
    <LeftBar
      :portfolio-title="portfolioTitle"
      :menu-items="menuItems"
      active-section="structure"
      @create-portfolio="isCreateModalOpen = true"
      @navigate="handleSidebarNavigation"
    />
    <main class="content structure-content">
      <div class="structure-header">
        <div>
          <h1>{{ portfolioTitle }}</h1>
          <div class="tabs structure-tabs">
            <button type="button" @click="goToAnalysis">Анализ</button>
            <button class="active" type="button">Структура</button>
          </div>
        </div>
        <button class="add-asset-button" type="button" @click="isAssetModalOpen = true">Добавить актив</button>
      </div>

      <section class="asset-table" aria-label="Структура портфеля">
        <div class="asset-row asset-header">
          <span>Symbol</span>
          <span>Current price</span>
          <span>Amount</span>
          <span>Cost</span>
          <span>Percentage (%)</span>
          <span>PnL</span>
        </div>
        <div v-for="asset in assets" :key="asset.id" class="asset-row">
          <span class="symbol-cell">
            {{ asset.symbol }}
            <img v-if="asset.isRiskFree" class="risk-free-icon" :src="riskFreeIcon" alt="Безрисковая ставка">
          </span>
          <span>{{ asset.currentPrice }}</span>
          <span>{{ asset.amount }}</span>
          <span>{{ asset.cost }}</span>
          <span>{{ asset.percentage }}</span>
          <span class="asset-pnl">{{ asset.pnl }}</span>
        </div>
      </section>
    </main>

    <CreatePortfolioModal v-if="isCreateModalOpen" @close="isCreateModalOpen = false" @created="router.push('/portfolio/' + $event)" />

    <div v-if="isAssetModalOpen" class="asset-modal-backdrop" @click.self="closeAssetModal">
      <section class="asset-modal" role="dialog" aria-modal="true" aria-labelledby="asset-modal-title">
        <h2 id="asset-modal-title" class="visually-hidden">Добавление актива</h2>
        <div class="asset-type-switch">
          <button :class="{ selected: assetType === 'risky' }" type="button" @click="assetType = 'risky'">Рисковый</button>
          <button :class="{ selected: assetType === 'riskFree' }" type="button" @click="assetType = 'riskFree'">Безрисковый</button>
        </div>

        <template v-if="assetType === 'risky'">
          <label for="asset-symbol">Symbol</label>
          <input id="asset-symbol" v-model="symbol" type="text" placeholder="BTC">
          <label for="asset-amount">Количество активов</label>
          <input id="asset-amount" v-model="amount" type="number" min="0" step="any" placeholder="2.2">
        </template>

        <template v-else>
          <label for="asset-risk-free-rate">Безрисковая ставка - r</label>
          <select id="asset-risk-free-rate" v-model="riskFreeRate">
            <option>US Treasuries (+5,10%)</option>
            <option>US Treasury Bills 3M (+4,80%)</option>
            <option>EUR Government Bonds (+3,20%)</option>
            <option>UK Gilts (+4,35%)</option>
          </select>
          <label for="asset-risk-free-value">Стоимость, USD</label>
          <input id="asset-risk-free-value" v-model="riskFreeValue" type="number" min="0" step="any" placeholder="10000">
        </template>

        <button class="asset-submit-button" type="button" @click="addAsset">Добавить</button>
      </section>
    </div>
  </div>
</template>

<style src="./StructureView.css" scoped></style>
