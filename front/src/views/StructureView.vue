<script setup>
import { ref } from 'vue'
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
const currentPrice = ref('')
const amount = ref('')
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
    currentPrice: '82543 USD',
    amount: '2.2',
    cost: '165087 USD',
    percentage: '6.1 %',
    pnl: '14000 USD',
    isRiskFree: true,
  },
])

const portfolioTitle = 'Портфель №1'

function goToAnalysis() {
  router.push('/portfolio/' + route.params.id)
}

function handleSidebarNavigation(sectionId) {
  if (sectionId === 'structure') return
  router.push(`/portfolio/${route.params.id}#${sectionId}`)
}

function closeAssetModal() {
  isAssetModalOpen.value = false
  symbol.value = ''
  currentPrice.value = ''
  amount.value = ''
  assetType.value = 'risky'
}

function addAsset() {
  const assetSymbol = symbol.value.trim()

  if (!assetSymbol) return

  const isRiskFree = assetType.value === 'riskFree'
  const price = Number(currentPrice.value) || 0
  const units = Number(amount.value) || 0

  assets.value.push({
    id: Date.now(),
    symbol: assetSymbol,
    currentPrice: isRiskFree ? riskFreeRate.value : (price || 0) + ' USD',
    amount: isRiskFree ? '—' : String(units || 0),
    cost: isRiskFree ? '—' : (price * units).toFixed(0) + ' USD',
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

        <label for="asset-symbol">Symbol</label>
        <input id="asset-symbol" v-model="symbol" type="text" placeholder="BTC">

        <template v-if="assetType === 'risky'">
          <label for="asset-price">Current price</label>
          <input id="asset-price" v-model="currentPrice" type="number" min="0" placeholder="1200">
          <label for="asset-amount">Amount</label>
          <input id="asset-amount" v-model="amount" type="number" min="0" step="any" placeholder="2.2">
        </template>

        <template v-else>
          <label for="asset-risk-free-rate">Безрисковая ставка - r</label>
          <select id="asset-risk-free-rate" v-model="riskFreeRate">
            <option>US Treasuries (+5,10%)</option>
            <option>US Treasuries (+4,80%)</option>
            <option>EUR Bonds (+3,20%)</option>
          </select>
        </template>

        <button class="asset-submit-button" type="button" @click="addAsset">Добавить</button>
      </section>
    </div>
  </div>
</template>

<style src="./StructureView.css" scoped></style>
