<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'

import CreatePortfolioModal from '../components/CreatePortfolioModal/CreatePortfolioModal.vue'
import LeftBar from '../components/LeftBar/LeftBar.vue'
import PerformanceChart from '../components/PerformanceChart/PerformanceChart.vue'
import RiskMetric from '../components/RiskMetric/RiskMetric.vue'
import PnlMetric from '../components/PnlMetric/PnlMetric.vue'
import ImpliedMetric from '../components/ImpliedMetric/ImpliedMetric.vue'

const router = useRouter()
const isCreateModalOpen = ref(false)

function handleCreated(portfolioId) {
  isCreateModalOpen.value = false
  router.push(`/portfolio/${portfolioId}`)
}
</script>

<template>
  <div class="app-shell">
    <LeftBar @create-portfolio="isCreateModalOpen = true" />

    <main class="content">
      <h1>Портфель №1</h1>

      <div class="tabs">
        <button>Анализ</button>
        <button class="active">Структура</button>
      </div>

      <PerformanceChart />

      <section class="metrics">
        <RiskMetric />
        <PnlMetric />
        <ImpliedMetric />
      </section>
    </main>

    <CreatePortfolioModal v-if="isCreateModalOpen" @close="isCreateModalOpen = false" @created="handleCreated" />
  </div>
</template>

