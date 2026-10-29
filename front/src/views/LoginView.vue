<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'

import logoUrl from '../assets/logo.svg'
import { api } from '../services/api'

const router = useRouter()
const login = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

async function submitLogin() {
  error.value = ''
  loading.value = true
  try {
    await api.login(login.value.trim(), password.value)
    await router.push('/')
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
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
        <input id="login" v-model="login" type="text" placeholder="Введите логин, например demo" autocomplete="username" required>
        <label for="password">Пароль</label>
        <input id="password" v-model="password" type="password" placeholder="Введите пароль" autocomplete="current-password" required>
        <p v-if="error" class="login-error" role="alert">{{ error }}</p>
        <button class="login-button" type="submit" :disabled="loading">{{ loading ? 'Вход…' : 'Войти' }}</button>
        <RouterLink class="login-link" to="/reset-password">Забыли пароль?</RouterLink>
        <RouterLink class="login-link" to="/register">Нет аккаунта?</RouterLink>
      </form>
    </section>
  </main>
</template>

<style src="./LoginView.css" scoped></style>
