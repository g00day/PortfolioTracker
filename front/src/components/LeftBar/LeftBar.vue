<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'

import logoUrl from '../../assets/logo.svg'
import { api, clearSession, getCurrentUser } from '../../services/api'

const props = defineProps({
  hasPortfolio: {
    type: Boolean,
    default: true,
  },
  portfolioTitle: {
    type: String,
    default: 'Портфель №1',
  },
  activeSection: {
    type: String,
    default: '',
  },
  menuItems: {
    type: Array,
    default: () => [
      { id: 'overview', label: 'Обзор' },
      { id: 'structure', label: 'Структура' },
      { id: 'performance', label: 'Доходность' },
      { id: 'risks', label: 'Риски&VaR' },
      { id: 'settings', label: 'Настройки' },
    ],
  },
})

const emit = defineEmits(['createPortfolio', 'navigate'])

const router = useRouter()
const username = getCurrentUser()?.login || localStorage.getItem('username') || 'Пользователь'
const isRateModalOpen = ref(false)
const rateName = ref('')
const rateValue = ref('')
const rateError = ref('')
const rateSuccess = ref('')
const isRateSaving = ref(false)

function selectSection(sectionId) {
  emit('navigate', sectionId)
}

function logout() {
  clearSession()
  router.replace('/login')
}

function closeRateModal() {
  if (isRateSaving.value) return
  isRateModalOpen.value = false
  rateError.value = ''
  rateSuccess.value = ''
}

async function createRiskFreeRate() {
  const name = rateName.value.trim()
  const value = Number(rateValue.value)
  rateError.value = ''
  rateSuccess.value = ''

  if (!name) {
    rateError.value = 'Введите название ставки'
    return
  }
  if (rateValue.value === '' || !Number.isFinite(value)) {
    rateError.value = 'Введите корректное числовое значение'
    return
  }

  isRateSaving.value = true
  try {
    await api.createRiskFreeRate({ name, value })
    rateSuccess.value = 'Безрисковая ставка сохранена'
    rateName.value = ''
    rateValue.value = ''
    setTimeout(() => {
      isRateModalOpen.value = false
      rateSuccess.value = ''
    }, 700)
  } catch (error) {
    rateError.value = error.message
  } finally {
    isRateSaving.value = false
  }
}
</script>

<template>
  <aside class="leftbar">
    <div class="brand">
      <img class="icon" :src="logoUrl" alt="">
      <div class="name">Portfolio Tracker</div>
    </div>

    <nav v-if="props.hasPortfolio" class="portfolio-nav" aria-label="Навигация по портфелю">
      <button
        class="portfolio-selector"
        type="button"
        aria-controls="portfolio-section-menu"
        @click="selectSection('overview')"
      >
        <span>{{ props.portfolioTitle }}</span>
      </button>

      <div id="portfolio-section-menu" class="portfolio-menu">
          <button
            v-for="item in props.menuItems"
            :key="item.id"
            class="portfolio-menu-item"
            :class="{ active: props.activeSection === item.id }"
            type="button"
            @click="selectSection(item.id)"
          >
            {{ item.label }}
          </button>
      </div>
    </nav>

    <div class="bottom">
      <button class="rate-button" type="button" @click="isRateModalOpen = true">Добавить безрисковую ставку</button>
      <button class="create-button" type="button" @click="emit('createPortfolio')">Создать портфель</button>
      <div class="user">{{ username }}</div>
      <button class="logout-button" type="button" @click="logout">Выйти из аккаунта</button>
    </div>

    <div v-if="isRateModalOpen" class="rate-modal-backdrop" @click.self="closeRateModal">
      <form class="rate-modal" role="dialog" aria-modal="true" aria-labelledby="rate-modal-title" @submit.prevent="createRiskFreeRate">
        <div class="rate-modal-header">
          <h2 id="rate-modal-title">Новая безрисковая ставка</h2>
          <button type="button" aria-label="Закрыть" :disabled="isRateSaving" @click="closeRateModal">×</button>
        </div>

        <label for="rate-name">Название</label>
        <input id="rate-name" v-model.trim="rateName" type="text" placeholder="USD Treasury 3M" required autofocus>

        <label for="rate-value">Значение</label>
        <input id="rate-value" v-model="rateValue" type="number" step="any" placeholder="0.051" required>
        <small>Укажите ставку десятичной дробью: 0.051 = 5,1%</small>

        <p v-if="rateError" class="rate-message rate-error" role="alert">{{ rateError }}</p>
        <p v-if="rateSuccess" class="rate-message rate-success" role="status">{{ rateSuccess }}</p>
        <button class="rate-submit" type="submit" :disabled="isRateSaving">{{ isRateSaving ? 'Сохранение…' : 'Сохранить' }}</button>
      </form>
    </div>
  </aside>
</template>

<style src="./LeftBar.css" scoped></style>
