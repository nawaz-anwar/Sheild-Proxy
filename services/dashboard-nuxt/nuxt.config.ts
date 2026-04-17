export default defineNuxtConfig({
  compatibilityDate: '2025-01-01',
  devtools: { enabled: false },
  modules: ['@pinia/nuxt'],
  runtimeConfig: {
    backendApiBase: process.env.BACKEND_API_BASE ?? 'http://api-nest:3000',
  },
});
