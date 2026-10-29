<script setup>
import { computed, ref } from 'vue'

const periods = ['D', 'M', 'Y']
const selectedPeriod = ref('D')

const chartData = {
  D: {
    riskFree: '0,350 140,330 280,300 420,270 560,235 700,190 840,135 1000,80',
    pnl: '0,350 140,315 280,330 420,280 560,300 700,240 840,260 1000,205',
  },
  M: {
    riskFree: '0,350 140,340 280,315 420,285 560,250 700,205 840,155 1000,100',
    pnl: '0,350 140,290 280,315 420,250 560,275 700,185 840,220 1000,130',
  },
  Y: {
    riskFree: '0,350 140,325 280,285 420,235 560,175 700,115 840,65 1000,25',
    pnl: '0,350 140,305 280,260 420,290 560,205 700,150 840,175 1000,90',
  },
}

const activeData = computed(() => chartData[selectedPeriod.value])
</script>

<template>
  <section class="performance-chart">
    <div class="chart-row">
      <div class="chart-panel">
        <svg class="chart" viewBox="0 0 1000 420" preserveAspectRatio="none" role="img" aria-label="График доходности портфеля">
          <g class="chart-grid">
            <line x1="0" y1="70" x2="1000" y2="70"></line>
            <line x1="0" y1="140" x2="1000" y2="140"></line>
            <line x1="0" y1="210" x2="1000" y2="210"></line>
            <line x1="0" y1="280" x2="1000" y2="280"></line>
            <line x1="0" y1="350" x2="1000" y2="350"></line>
          </g>
          <polyline :points="activeData.riskFree" class="risk-free-line"></polyline>
          <polyline :points="activeData.pnl" class="pnl-line"></polyline>
        </svg>
      </div>

      <div class="legend">
        <div><i class="green"></i>Безрисковая ставка</div>
        <div><i class="yellow"></i>P&amp;L портфеля</div>
      </div>
    </div>

    <div class="periods">
      <button v-for="period in periods" :key="period" :class="{ selected: selectedPeriod === period }" @click="selectedPeriod = period">
        {{ period }}
      </button>
    </div>
  </section>
</template>

<style src="./PerformanceChart.css" scoped></style>
