<script setup>
import { ref } from 'vue'

import logoUrl from '../../assets/logo.svg'
import downIcon from '../../assets/down.svg'
import upIcon from '../../assets/up.svg'

const props = defineProps({
  hasPortfolio: {
    type: Boolean,
    default: true,
  },
})

const emit = defineEmits(['createPortfolio'])

const isPortfolioMenuOpen = ref(false)

const portfolioMenuItems = [
  'Доходность',
  'Риски&VaR',
  'Настройки',
]
</script>

<template>
  <aside class="leftbar">
    <img class="icon" :src="logoUrl" alt="Portfolio Tracker" />
    <div class="name">Portfolio Tracker</div>

    <nav v-if="props.hasPortfolio" class="portfolio-nav">
      <div class="portfolio-selector">
        <button class="portfolio-button" @click="isPortfolioMenuOpen = false">
          Портфель №1
        </button>

        <button class="portfolio-toggle" aria-label="Открыть меню портфеля" @click="isPortfolioMenuOpen = !isPortfolioMenuOpen">
          <img :src="isPortfolioMenuOpen ? upIcon : downIcon" alt="">
        </button>
      </div>

      <div v-if="isPortfolioMenuOpen" class="portfolio-menu">
        <button v-for="item in portfolioMenuItems" :key="item" class="portfolio-menu-item">
          {{ item }}
        </button>
      </div>
    </nav>

    <div class="bottom">
      <button class="create-button" @click="emit('createPortfolio')">Создать портфель</button>
      <div class="user">admin</div>
      <button class="logout-button">Выйти из аккаунта</button>
    </div>
  </aside>
</template>

<style src="./LeftBar.css" scoped></style>




