<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import CreatePortfolioModal from '../components/CreatePortfolioModal/CreatePortfolioModal.vue'
import LeftBar from '../components/LeftBar/LeftBar.vue'
import riskFreeIcon from '../assets/risk-free.svg'
import { api } from '../services/api'

const route = useRoute()
const router = useRouter()

const isCreateModalOpen = ref(false)
const isAssetModalOpen = ref(false)
const assetType = ref('risky')
const symbol = ref('')
const amount = ref('')
const riskFreeValue = ref('')
const riskFreeRate = ref('')
const rates = ref([])
const portfolio = ref(null)
const error = ref('')
const isTransactionModalOpen = ref(false)
const selectedAsset = ref(null)
const transactionType = ref('deposit')
const transactionAmount = ref('')
const transactionPrice = ref('')
const transactionLoading = ref(false)

const assets = ref([])

const portfolioTitle = computed(() => {
  return portfolio.value?.name || `Портфель №${route.params.id}`
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

async function loadStructure() {
  const data = await api.getStructure(route.params.id)
  assets.value = (data.portfolio_structure || []).map((row, index) => ({ id: Number(row[6] ?? index), symbol: row[0], currentPriceValue: Number(row[1]), currentPrice: `${Number(row[1]).toLocaleString('ru-RU')} USD`, amount: Number(row[2]).toLocaleString('ru-RU'), cost: `${Number(row[3]).toLocaleString('ru-RU')} USD`, percentage: `${Number(row[4]).toLocaleString('ru-RU', { maximumFractionDigits: 2 })} %`, pnl: `${Number(row[5]).toLocaleString('ru-RU')} USD`, isRiskFree: rates.value.some((r) => r.name === row[0]) }))
}

function openTransaction(asset) {
  selectedAsset.value = asset
  transactionType.value = 'deposit'
  transactionAmount.value = ''
  transactionPrice.value = asset.isRiskFree ? '1' : String(asset.currentPriceValue || '')
  error.value = ''
  isTransactionModalOpen.value = true
}

function closeTransaction() {
  if (transactionLoading.value) return
  isTransactionModalOpen.value = false
  selectedAsset.value = null
}

async function submitTransaction() {
  const transactionAmountValue = Number(transactionAmount.value)
  const price = selectedAsset.value?.isRiskFree ? 1 : Number(transactionPrice.value)
  if (!transactionAmountValue || transactionAmountValue <= 0 || !price || price <= 0) {
    error.value = 'Введите положительные количество и цену'
    return
  }
  transactionLoading.value = true
  error.value = ''
  try {
    await api.addTransaction(route.params.id, { portfolio_item_id: selectedAsset.value.id, type: transactionType.value, amount: transactionAmountValue, price_usd: price })
    closeTransaction()
    await loadStructure()
  } catch (e) {
    error.value = e.message
  } finally {
    transactionLoading.value = false
    if (!error.value) closeTransaction()
  }
}

async function addAsset() {
  const isRiskFree = assetType.value === 'riskFree'
  const assetSymbol = isRiskFree ? riskFreeRate.value.split(' (')[0] : symbol.value.trim().toUpperCase()
  if (!assetSymbol) return
  const units = Number(amount.value) || 0
  const value = Number(riskFreeValue.value) || 0

  if ((!isRiskFree && !units) || (isRiskFree && !value)) return

  try {
    await api.addPortfolioItem(route.params.id, { ticker: isRiskFree ? null : assetSymbol, risk_free_rate_id: isRiskFree ? Number(riskFreeRate.value) : null, total_amount: isRiskFree ? value : units })
    closeAssetModal(); await loadStructure()
  } catch (e) { error.value = e.message }
}

onMounted(async () => {
  try {
    const [portfolios, riskRates] = await Promise.all([api.getPortfolios(), api.getRiskFreeRates()])
    portfolio.value = portfolios.find((item) => String(item.id) === String(route.params.id))
    rates.value = riskRates; riskFreeRate.value = String(riskRates[0]?.id || '')
    await loadStructure()
  } catch (e) { error.value = e.message }
})
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
          <p v-if="error" class="negative" role="alert">{{ error }}</p>
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
          <span>Действие</span>
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
          <span><button class="transaction-button" type="button" @click="openTransaction(asset)">Транзакция</button></span>
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
          <input id="asset-symbol" v-model="symbol" type="text" placeholder="Например, BTCUSDT">
          <label for="asset-amount">Количество активов</label>
          <input id="asset-amount" v-model="amount" type="number" min="0" step="any" placeholder="Например, 0.25">
        </template>

        <template v-else>
          <label for="asset-risk-free-rate">Безрисковая ставка - r</label>
          <select id="asset-risk-free-rate" v-model="riskFreeRate">
            <option v-for="rate in rates" :key="rate.id" :value="String(rate.id)">{{ rate.name }} ({{ (rate.value * 100).toLocaleString('ru-RU') }}%)</option>
          </select>
          <label for="asset-risk-free-value">Стоимость, USD</label>
          <input id="asset-risk-free-value" v-model="riskFreeValue" type="number" min="0" step="any" placeholder="Например, 10 000">
        </template>

        <button class="asset-submit-button" type="button" @click="addAsset">Добавить</button>
      </section>
    </div>

    <div v-if="isTransactionModalOpen" class="asset-modal-backdrop" @click.self="closeTransaction">
      <form class="asset-modal transaction-modal" role="dialog" aria-modal="true" aria-labelledby="transaction-title" @submit.prevent="submitTransaction">
        <h2 id="transaction-title">Транзакция: {{ selectedAsset?.symbol }}</h2>
        <div class="asset-type-switch">
          <button :class="{ selected: transactionType === 'deposit' }" type="button" @click="transactionType = 'deposit'">Пополнение</button>
          <button :class="{ selected: transactionType === 'withdraw' }" type="button" @click="transactionType = 'withdraw'">Вывод</button>
        </div>
        <label for="transaction-amount">Количество</label>
        <input id="transaction-amount" v-model="transactionAmount" type="number" min="0" step="any" placeholder="Например, 0.1" required>
        <label for="transaction-price">Цена за единицу, USD</label>
        <input id="transaction-price" v-model="transactionPrice" type="number" min="0" step="any" :disabled="selectedAsset?.isRiskFree" placeholder="Например, 65000" required>
        <p v-if="error" class="transaction-error" role="alert">{{ error }}</p>
        <button class="asset-submit-button" type="submit" :disabled="transactionLoading">{{ transactionLoading ? 'Сохранение…' : 'Сохранить' }}</button>
      </form>
    </div>
  </div>
</template>

<style src="./StructureView.css" scoped></style>
