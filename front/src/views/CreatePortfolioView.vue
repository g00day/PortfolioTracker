<template>
  <main class="create-portfolio-page">
    <h1>Создать портфель</h1>
    <button @click="createPortfolio">Создать</button>
  </main>
</template>

<script setup>
import { useRouter } from 'vue-router'

const router = useRouter()

function createPortfolio() {
  localStorage.setItem('portfolios', JSON.stringify(['1']))
  router.push('/portfolio/1')
}
</script>
