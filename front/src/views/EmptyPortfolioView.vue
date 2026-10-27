<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'

import CreatePortfolioModal from '../components/CreatePortfolioModal/CreatePortfolioModal.vue'
import LeftBar from '../components/LeftBar/LeftBar.vue'

const router = useRouter()
const isModalOpen = ref(false)

function openCreatePortfolio() {
  isModalOpen.value = true
}

function handleCreated(portfolioId) {
  isModalOpen.value = false
  router.push(`/portfolio/${portfolioId}`)
}
</script>

<template>
  <div class="app-shell empty-portfolio-page">
    <LeftBar :has-portfolio="false" @create-portfolio="openCreatePortfolio" />

    <main class="empty-portfolio-content">
      <button class="create-portfolio-button" @click="openCreatePortfolio">
        Создать Портфель
      </button>
    </main>

    <CreatePortfolioModal v-if="isModalOpen" @close="isModalOpen = false" @created="handleCreated" />
  </div>
</template>
