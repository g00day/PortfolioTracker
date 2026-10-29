<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'

import logoUrl from '../assets/logo.svg'
import { api } from '../services/api'

const router = useRouter()
const login = ref('')
const email = ref('')
const password = ref('')
const repeatPassword = ref('')
const error = ref('')
const loading = ref(false)

async function submitRegistration() {
  error.value = password.value === repeatPassword.value ? '' : 'Пароли не совпадают'

  if (error.value) return

  loading.value = true
  try {
    await api.register(login.value, email.value, password.value)
    await api.login(login.value, password.value)
    await router.push('/')
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <main class="register-page">
    <section class="register-content">
      <img class="register-logo" :src="logoUrl" alt="Portfolio Tracker">
      <h1>Portfolio Tracker</h1>
      <h2>Регистрация</h2>
      <form class="register-form" @submit.prevent="submitRegistration">
        <label for="register-login">Логин</label>
        <input id="register-login" v-model.trim="login" type="text" placeholder="Например, investor" autocomplete="username" required>
        <label for="register-email">Email</label>
        <input id="register-email" v-model.trim="email" type="email" placeholder="example@mail.ru" required>
        <label for="register-password">Пароль</label>
        <input id="register-password" v-model="password" type="password" placeholder="Придумайте пароль" autocomplete="new-password" required>
        <label for="register-repeat-password">Повторите пароль</label>
        <input id="register-repeat-password" v-model="repeatPassword" type="password" placeholder="Повторите пароль" autocomplete="new-password" required>
        <p v-if="error" class="register-error">{{ error }}</p>
        <button class="register-button" type="submit" :disabled="loading">{{ loading ? 'Регистрация…' : 'Зарегистрироваться' }}</button>
        <RouterLink class="register-link" to="/login">Есть аккаунт?</RouterLink>
      </form>
    </section>
  </main>
</template>

<style src="./RegisterView.css" scoped></style>
