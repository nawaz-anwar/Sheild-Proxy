export default defineNuxtConfig({
  compatibilityDate: '2025-01-01',
  devtools: { enabled: false },
  modules: [
    '@pinia/nuxt',
    '@nuxtjs/tailwindcss',
    '@pinia-plugin-persistedstate/nuxt'
  ],
  runtimeConfig: {
    backendApiBase: process.env.NUXT_BACKEND_API_BASE ?? 'http://api:3000',
  },
  css: ['~/assets/css/main.css'],
  app: {
    head: {
      title: 'ShieldProxy - Security Dashboard',
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
      ],
    },
  },
});
