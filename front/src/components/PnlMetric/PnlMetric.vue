<script setup>
import MetricCard from '../MetricCard/MetricCard.vue'

const items = [
  ['Суточный PnL', '-100 USD (-5%)', 'negative'],
  ['Месячный PnL', '+200 USD (+6%)', 'positive'],
  ['Исторический PnL', '+450 USD (+3%)', 'positive'],
]
</script>

<template>
  <div class="metric-wrapper">
    <MetricCard>
      <div class="gauge gauge-pnl"></div>
      <div class="metric-title">
        <b>P&amp;L</b>
        <strong class="negative">-400 USD (-25%)</strong>
      </div>
    </MetricCard>

    <div class="metric-details">
      <ul>
        <li v-for="item in items" :key="item[0]">
          {{ item[0] }}
          <strong :class="item[2]">{{ item[1] }}</strong>
        </li>
      </ul>
    </div>
  </div>
</template>

<style src="./PnlMetric.css" scoped></style>
