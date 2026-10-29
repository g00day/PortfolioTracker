<script setup>
import { ref } from 'vue'
import { api, extractCreatedId } from '../../services/api'

const emit = defineEmits(['close', 'created'])

const portfolioName = ref('')
const riskFreeRateId = ref('')
const rates = ref([])
const error = ref('')
const loading = ref(false)

api.getRiskFreeRates().then((data) => { rates.value = data; riskFreeRateId.value = String(data[0]?.id || '') }).catch((e) => { error.value = e.message })

async function submitPortfolio() {
  if (!portfolioName.value.trim() || !riskFreeRateId.value) return
  loading.value = true
  error.value = ''
  try {
    const result = await api.createPortfolio({ name: portfolioName.value.trim(), risk_free_rate_id: Number(riskFreeRateId.value) })
    emit('created', extractCreatedId(result.message))
  } catch (e) { error.value = e.message } finally { loading.value = false }
}
</script>

<template>
  <div class="modal-backdrop" @click.self="emit('close')">
    <section class="portfolio-modal" role="dialog" aria-modal="true" aria-labelledby="portfolio-modal-title">
      <h2 id="portfolio-modal-title" class="visually-hidden">Создание портфеля</h2>

      <label for="portfolio-name">Название портфеля</label>
      <input id="portfolio-name" v-model="portfolioName" type="text" placeholder="Например, Основной портфель" required>

      <label for="risk-free-rate">Безрисковая ставка - r</label>
      <select id="risk-free-rate" v-model="riskFreeRateId">
        <option v-for="rate in rates" :key="rate.id" :value="String(rate.id)">{{ rate.name }} ({{ (rate.value * 100).toLocaleString('ru-RU') }}%)</option>
      </select>

      <p v-if="error" role="alert">{{ error }}</p>
      <button class="create-button" type="button" :disabled="loading || !riskFreeRateId" @click="submitPortfolio">{{ loading ? 'Создание…' : 'Создать' }}</button>
    </section>
  </div>
</template>

<style src="./CreatePortfolioModal.css" scoped></style>
