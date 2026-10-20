<script setup>
import { ref } from 'vue'

const emit = defineEmits(['close', 'created'])

const portfolioName = ref('Портфель №1')
const riskFreeRate = ref('US Treasuries (+5,10%)')

function submitPortfolio() {
  const portfolios = JSON.parse(localStorage.getItem('portfolios') || '[]')
  const portfolioId = String(Date.now())

  portfolios.push(portfolioId)
  localStorage.setItem('portfolios', JSON.stringify(portfolios))
  emit('created', portfolioId)
}
</script>

<template>
  <div class="modal-backdrop" @click.self="emit('close')">
    <section class="portfolio-modal" role="dialog" aria-modal="true" aria-labelledby="portfolio-modal-title">
      <h2 id="portfolio-modal-title" class="visually-hidden">Создание портфеля</h2>

      <label for="portfolio-name">Название портфеля</label>
      <input id="portfolio-name" v-model="portfolioName" type="text">

      <label for="risk-free-rate">Безрисковая ставка - r</label>
      <select id="risk-free-rate" v-model="riskFreeRate">
        <option>US Treasuries (+5,10%)</option>
        <option>US Treasuries (+4,80%)</option>
        <option>EUR Bonds (+3,20%)</option>
      </select>

      <button class="create-button" type="button" @click="submitPortfolio">Создать</button>
    </section>
  </div>
</template>

<style src="./CreatePortfolioModal.css" scoped></style>
