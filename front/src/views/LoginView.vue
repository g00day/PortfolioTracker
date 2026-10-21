<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'

import logoUrl from '../assets/logo.svg'

const router = useRouter()
const login = ref('')
const password = ref('')

function submitLogin() {
  localStorage.setItem('isAuthenticated', 'true')
  router.push('/')
}
</script>

<template>
  <main class="login-page">
    <section class="login-content">
      <img class="login-logo" :src="logoUrl" alt="Portfolio Tracker">
      <h1>Portfolio Tracker</h1>
      <h2>Авторизация</h2>
      <form class="login-form" @submit.prevent="submitLogin">
        <label for="login">Логин</label>
        <input id="login" v-model="login" type="text" placeholder="Введите логин" required>
        <label for="password">Пароль</label>
        <input id="password" v-model="password" type="password" placeholder="Введите пароль" required>
        <button class="login-button" type="submit">Войти</button>
        <RouterLink class="login-link" to="/reset-password">Забыли пароль?</RouterLink>
        <RouterLink class="login-link" to="/register">Нет аккаунта?</RouterLink>
      </form>
    </section>
  </main>
</template>

<style src="./LoginView.css" scoped></style>
