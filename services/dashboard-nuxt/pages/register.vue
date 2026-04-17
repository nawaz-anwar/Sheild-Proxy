<template>
  <div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-primary-50 to-primary-100 px-4">
    <div class="max-w-md w-full">
      <div class="text-center mb-8">
        <div class="inline-flex items-center justify-center w-16 h-16 bg-primary-600 rounded-2xl mb-4">
          <ShieldCheckIcon class="w-10 h-10 text-white" />
        </div>
        <h1 class="text-3xl font-bold text-gray-900">ShieldProxy</h1>
        <p class="text-gray-600 mt-2">Create your account</p>
      </div>

      <div class="card">
        <form @submit.prevent="handleRegister" class="space-y-4">
          <div v-if="auth.error" class="p-3 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm">
            {{ auth.error }}
          </div>

          <div>
            <label for="name" class="block text-sm font-medium text-gray-700 mb-1">Name</label>
            <input
              id="name"
              v-model="form.name"
              type="text"
              required
              class="input-field"
              placeholder="John Doe"
            />
          </div>

          <div>
            <label for="email" class="block text-sm font-medium text-gray-700 mb-1">Email</label>
            <input
              id="email"
              v-model="form.email"
              type="email"
              required
              class="input-field"
              placeholder="you@example.com"
            />
          </div>

          <div>
            <label for="password" class="block text-sm font-medium text-gray-700 mb-1">Password</label>
            <input
              id="password"
              v-model="form.password"
              type="password"
              required
              minlength="8"
              class="input-field"
              placeholder="••••••••"
            />
          </div>

          <button
            type="submit"
            :disabled="auth.loading"
            class="btn-primary w-full"
          >
            <span v-if="auth.loading">Creating account...</span>
            <span v-else>Create account</span>
          </button>
        </form>

        <div class="mt-6 text-center">
          <p class="text-sm text-gray-600">
            Already have an account?
            <NuxtLink to="/login" class="text-primary-600 hover:text-primary-700 font-medium">
              Sign in
            </NuxtLink>
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ShieldCheckIcon } from '@heroicons/vue/24/solid';

definePageMeta({
  layout: false,
});

const auth = useAuthStore();
const form = reactive({
  name: '',
  email: '',
  password: '',
});

const handleRegister = async () => {
  try {
    await auth.register(form);
    navigateTo('/');
  } catch (error) {
    // Error is handled in store
  }
};

onMounted(() => {
  auth.clearError();
});
</script>
