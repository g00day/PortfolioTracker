<template>
  <main class="create-portfolio-page">
    <h1>Создать портфель</h1>
    <button @click="createPortfolio">Создать</button>
  </main>
</template>

<script setup>
import { useRouter } from 'vue-router'
import { api, extractCreatedId } from '../services/api'

const router = useRouter()

async function createPortfolio() {
  const rates = await api.getRiskFreeRates()
  if (!rates.length) return
  const result = await api.createPortfolio({ name: 'Портфель №1', risk_free_rate_id: rates[0].id })
  router.push('/portfolio/' + extractCreatedId(result.message))
}
</script>
