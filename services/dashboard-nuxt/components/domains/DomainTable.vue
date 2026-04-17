<template>
  <div class="card">
    <div class="flex items-center justify-between mb-6">
      <h3 class="text-lg font-semibold text-gray-900">Your Domains</h3>
      <UiButton @click="emit('add')">
        <PlusIcon class="w-4 h-4 mr-2" />
        Add Domain
      </UiButton>
    </div>

    <div v-if="loading" class="space-y-3">
      <div v-for="i in 3" :key="i" class="h-16 bg-gray-100 rounded-lg animate-pulse"></div>
    </div>

    <div v-else-if="!domains || domains.length === 0" class="text-center py-12">
      <GlobeAltIcon class="w-12 h-12 text-gray-400 mx-auto mb-3" />
      <p class="text-gray-600 mb-4">No domains yet</p>
      <UiButton @click="emit('add')">Add your first domain</UiButton>
    </div>

    <div v-else class="overflow-x-auto">
      <table class="min-w-full divide-y divide-gray-200">
        <thead>
          <tr>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
              Domain
            </th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
              Upstream
            </th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
              Status
            </th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
              Created
            </th>
            <th class="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
              Actions
            </th>
          </tr>
        </thead>
        <tbody class="bg-white divide-y divide-gray-200">
          <tr v-for="domain in domains" :key="domain.id" class="hover:bg-gray-50">
            <td class="px-4 py-4 whitespace-nowrap">
              <div class="flex items-center">
                <GlobeAltIcon class="w-5 h-5 text-gray-400 mr-2" />
                <span class="text-sm font-medium text-gray-900">{{ domain.domain }}</span>
              </div>
            </td>
            <td class="px-4 py-4 whitespace-nowrap text-sm text-gray-600">
              {{ domain.upstreamUrl }}
            </td>
            <td class="px-4 py-4 whitespace-nowrap">
              <span
                class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium"
                :class="statusClass(domain.status)"
              >
                {{ domain.status }}
              </span>
            </td>
            <td class="px-4 py-4 whitespace-nowrap text-sm text-gray-600">
              {{ formatDate(domain.createdAt) }}
            </td>
            <td class="px-4 py-4 whitespace-nowrap text-right text-sm font-medium">
              <NuxtLink
                :to="`/domains/${domain.id}`"
                class="text-primary-600 hover:text-primary-900"
              >
                View Details
              </NuxtLink>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { PlusIcon, GlobeAltIcon } from '@heroicons/vue/24/outline';

defineProps<{
  domains: Array<{
    id: string;
    domain: string;
    upstreamUrl: string;
    status: string;
    createdAt: string;
  }>;
  loading?: boolean;
}>();

const emit = defineEmits<{
  add: [];
  select: [id: string];
}>();

const statusClass = (status: string) => {
  const classes = {
    active: 'bg-green-100 text-green-800',
    pending: 'bg-yellow-100 text-yellow-800',
    inactive: 'bg-gray-100 text-gray-800',
  };
  return classes[status as keyof typeof classes] || classes.inactive;
};

const formatDate = (date: string) => {
  return new Date(date).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });
};
</script>
