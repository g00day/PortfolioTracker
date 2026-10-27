<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'

import logoUrl from '../../assets/logo.svg'
import downIcon from '../../assets/down.svg'
import upIcon from '../../assets/up.svg'

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
      { id: 'risk', label: 'Анализ рисков' },
      { id: 'volatility', label: 'Volatility Drag' },
      { id: 'pnl', label: 'Анализ P&L' },
      { id: 'structure', label: 'Структура' },
    ],
  },
})

const emit = defineEmits(['createPortfolio', 'navigate'])

const router = useRouter()
const isPortfolioMenuOpen = ref(false)
const username = localStorage.getItem('username') || 'admin'

function selectSection(sectionId) {
  emit('navigate', sectionId)
}

function logout() {
  localStorage.removeItem('isAuthenticated')
  localStorage.removeItem('username')
  router.replace('/login')
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
        :aria-expanded="isPortfolioMenuOpen"
        aria-controls="portfolio-section-menu"
        @click="isPortfolioMenuOpen = !isPortfolioMenuOpen"
      >
        <span>{{ props.portfolioTitle }}</span>
        <img :src="isPortfolioMenuOpen ? upIcon : downIcon" alt="">
      </button>

      <Transition name="menu">
        <div v-if="isPortfolioMenuOpen" id="portfolio-section-menu" class="portfolio-menu">
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
      </Transition>
    </nav>

    <div class="bottom">
      <button class="create-button" type="button" @click="emit('createPortfolio')">Создать портфель</button>
      <div class="user">{{ username }}</div>
      <button class="logout-button" type="button" @click="logout">Выйти из аккаунта</button>
    </div>
  </aside>
</template>

<style src="./LeftBar.css" scoped></style>
