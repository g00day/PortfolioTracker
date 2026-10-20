<template>
  <div class="metric-wrapper">
    <MetricCard>
      <div class="decline-line"></div>
      <div class="metric-title"><b>Implied P&amp;L</b></div>
    </MetricCard>

    <div class="metric-details">
      <ul>
        <li>Implied Volatility <strong class="negative">-2%</strong></li>
        <li>Implied PnL <strong class="negative">-100 USD (-10%)</strong></li>
      </ul>
    </div>
  </div>
</template>

<script setup>
import MetricCard from '../MetricCard/MetricCard.vue'
</script>

<style src="./ImpliedMetric.css" scoped></style>
