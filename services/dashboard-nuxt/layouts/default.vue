<template>
  <div class="min-h-screen bg-gray-50">
    <div class="flex h-screen">
      <!-- Sidebar -->
      <aside class="w-64 bg-white border-r border-gray-200 flex flex-col">
        <div class="p-6 border-b border-gray-200">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 bg-primary-600 rounded-lg flex items-center justify-center">
              <ShieldCheckIcon class="w-6 h-6 text-white" />
            </div>
            <div>
              <h1 class="text-lg font-bold text-gray-900">ShieldProxy</h1>
              <p class="text-xs text-gray-500">Security Dashboard</p>
            </div>
          </div>
        </div>

        <nav class="flex-1 p-4 space-y-1">
          <NuxtLink
            to="/"
            class="nav-item"
            :class="{ 'nav-item-active': $route.path === '/' }"
          >
            <ChartBarIcon class="w-5 h-5" />
            <span>Dashboard</span>
          </NuxtLink>
          
          <NuxtLink
            to="/domains"
            class="nav-item"
            :class="{ 'nav-item-active': $route.path === '/domains' }"
          >
            <GlobeAltIcon class="w-5 h-5" />
            <span>Domains</span>
          </NuxtLink>
          
          <NuxtLink
            to="/analytics"
            class="nav-item"
            :class="{ 'nav-item-active': $route.path === '/analytics' }"
          >
            <ChartPieIcon class="w-5 h-5" />
            <span>Analytics</span>
          </NuxtLink>
        </nav>

        <div class="p-4 border-t border-gray-200">
          <div class="flex items-center gap-3 mb-3">
            <div class="w-8 h-8 bg-primary-100 rounded-full flex items-center justify-center">
              <UserIcon class="w-4 h-4 text-primary-600" />
            </div>
            <div class="flex-1 min-w-0">
              <p class="text-sm font-medium text-gray-900 truncate">{{ auth.email }}</p>
              <p class="text-xs text-gray-500">Client ID: {{ shortClientId }}</p>
            </div>
          </div>
          <button
            @click="handleLogout"
            class="w-full flex items-center gap-2 px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
          >
            <ArrowRightOnRectangleIcon class="w-4 h-4" />
            <span>Sign out</span>
          </button>
        </div>
      </aside>

      <!-- Main content -->
      <main class="flex-1 overflow-auto">
        <slot />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import {
  ShieldCheckIcon,
  ChartBarIcon,
  GlobeAltIcon,
  ChartPieIcon,
  UserIcon,
  ArrowRightOnRectangleIcon,
} from '@heroicons/vue/24/outline';

const auth = useAuthStore();

const shortClientId = computed(() => {
  if (!auth.clientId) return '';
  return auth.clientId.slice(0, 8);
});

const handleLogout = () => {
  auth.logout();
};
</script>

<style scoped>
.nav-item {
  @apply flex items-center gap-3 px-3 py-2 text-sm font-medium text-gray-700 rounded-lg hover:bg-gray-100 transition-colors;
}

.nav-item-active {
  @apply bg-primary-50 text-primary-700 hover:bg-primary-100;
}
</style>
