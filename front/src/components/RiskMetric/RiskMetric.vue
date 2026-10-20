<script setup>
import MetricCard from '../MetricCard/MetricCard.vue'
</script>

<template>
  <div class="metric-wrapper">
    <MetricCard>
      <div class="gauge-value">37%</div>
      <div class="gauge gauge-risk"></div>
      <h3>Высокий риск</h3>
    </MetricCard>

    <div class="metric-details risk-details">
      <ul>
        <li>Volatility Drag</li>
        <li>High VaR</li>
        <li>Low Diversification</li>
      </ul>
    </div>
  </div>
</template>

<style src="./RiskMetric.css" scoped></style>
