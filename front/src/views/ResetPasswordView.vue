<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'

import logoUrl from '../assets/logo.svg'

const router = useRouter()
const step = ref(1)
const email = ref('')
const code = ref('')
const password = ref('')
const repeatPassword = ref('')
const error = ref('')

function sendCode() {
  error.value = ''
  step.value = 2
}

function verifyCode() {
  error.value = ''

  if (!code.value.trim()) {
    error.value = 'Введите код из письма'
    return
  }

  step.value = 3
}

function changePassword() {
  if (password.value !== repeatPassword.value) {
    error.value = 'Пароли не совпадают'
    return
  }

  router.push('/login')
}
</script>

<template>
  <main class="reset-page">
    <section class="reset-content">
      <img class="reset-logo" :src="logoUrl" alt="Portfolio Tracker">
      <h1>Portfolio Tracker</h1>
      <h2>Восстановление пароля</h2>
      <form v-if="step === 1" class="reset-form" @submit.prevent="sendCode">
        <label for="reset-email">Email</label>
        <input id="reset-email" v-model.trim="email" type="email" placeholder="example@mail.ru" required>
        <button class="reset-button" type="submit">Отправить код</button>
      </form>
      <form v-else-if="step === 2" class="reset-form" @submit.prevent="verifyCode">
        <p class="reset-hint">Код отправлен на {{ email }}</p>
        <label for="reset-code">Код из письма</label>
        <input id="reset-code" v-model="code" type="text" inputmode="numeric" required>
        <button class="reset-button" type="submit">Проверить код</button>
      </form>
      <form v-else class="reset-form" @submit.prevent="changePassword">
        <label for="new-password">Новый пароль</label>
        <input id="new-password" v-model="password" type="password" required>
        <label for="repeat-new-password">Повторите пароль</label>
        <input id="repeat-new-password" v-model="repeatPassword" type="password" required>
        <p v-if="error" class="reset-error">{{ error }}</p>
        <button class="reset-button" type="submit">Сохранить пароль</button>
      </form>
      <p v-if="step !== 3 && error" class="reset-error">{{ error }}</p>
      <RouterLink class="reset-link" to="/login">Вернуться ко входу</RouterLink>
    </section>
  </main>
</template>

<style src="./ResetPasswordView.css" scoped></style>
