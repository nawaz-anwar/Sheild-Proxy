<template>
  <div class="p-8">
    <div class="mb-8">
      <h1 class="text-2xl font-bold text-gray-900">Analytics</h1>
      <p class="text-gray-600 mt-1">Detailed traffic analytics and insights</p>
    </div>

    <!-- Domain Selector -->
    <div class="mb-6 card">
      <div class="flex items-center justify-between">
        <div class="flex-1">
          <label for="domain-select" class="block text-sm font-medium text-gray-700 mb-2">
            Filter by Domain
          </label>
          <select
            id="domain-select"
            v-model="selectedDomainId"
            @change="handleDomainChange"
            class="input-field max-w-md"
          >
            <option value="">All Domains</option>
            <option v-for="domain in domains" :key="domain.id" :value="domain.id">
              {{ domain.domain }}
            </option>
          </select>
        </div>
        <div class="flex items-center gap-2">
          <UiButton variant="secondary" size="sm" @click="handleRefresh">
            <ArrowPathIcon class="w-4 h-4 mr-2" :class="{ 'animate-spin': analyticsLoading }" />
            Refresh
          </UiButton>
        </div>
      </div>
    </div>

    <!-- Stats Grid -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
      <UiStatCard
        title="Total Requests"
        :value="overview.totalRequests.toLocaleString()"
        :loading="analyticsLoading"
        :icon="ChartBarIcon"
        icon-color="text-primary-600"
        icon-bg-color="bg-primary-100"
      />
      
      <UiStatCard
        title="Blocked Requests"
        :value="overview.blockedRequests.toLocaleString()"
        :loading="analyticsLoading"
        :icon="ShieldExclamationIcon"
        icon-color="text-red-600"
        icon-bg-color="bg-red-100"
      />
      
      <UiStatCard
        title="Block Rate"
        :value="blockRateLabel"
        :loading="analyticsLoading"
        :icon="ChartPieIcon"
        icon-color="text-yellow-600"
        icon-bg-color="bg-yellow-100"
      />
    </div>

    <!-- Traffic Chart -->
    <div class="mb-8">
      <ChartsTrafficChart :data="timeSeries.points" :loading="analyticsLoading" />
    </div>

    <!-- Top IPs Table -->
    <div class="card">
      <h3 class="text-lg font-semibold text-gray-900 mb-6">Top IP Addresses</h3>
      
      <div v-if="analyticsLoading" class="space-y-3">
        <div v-for="i in 10" :key="i" class="h-12 bg-gray-100 rounded animate-pulse"></div>
      </div>
      
      <div v-else-if="topIPs.items.length === 0" class="text-center py-12">
        <ServerIcon class="w-12 h-12 text-gray-400 mx-auto mb-3" />
        <p class="text-gray-600">No traffic data available</p>
      </div>
      
      <div v-else class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200">
          <thead>
            <tr>
              <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Rank
              </th>
              <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                IP Address
              </th>
              <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Total Requests
              </th>
              <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Blocked
              </th>
              <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Block Rate
              </th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            <tr v-for="(item, index) in topIPs.items" :key="item.ip" class="hover:bg-gray-50">
              <td class="px-4 py-4 whitespace-nowrap text-sm text-gray-500">
                #{{ index + 1 }}
              </td>
              <td class="px-4 py-4 whitespace-nowrap text-sm font-mono text-gray-900">
                {{ item.ip }}
              </td>
              <td class="px-4 py-4 whitespace-nowrap text-sm text-gray-600">
                {{ item.requests.toLocaleString() }}
              </td>
              <td class="px-4 py-4 whitespace-nowrap text-sm text-gray-600">
                {{ item.blocked.toLocaleString() }}
              </td>
              <td class="px-4 py-4 whitespace-nowrap">
                <div class="flex items-center gap-2">
                  <div class="flex-1 bg-gray-200 rounded-full h-2 max-w-[100px]">
                    <div
                      class="h-2 rounded-full"
                      :class="getBlockRateBarClass(item.blocked / item.requests)"
                      :style="{ width: `${(item.blocked / item.requests) * 100}%` }"
                    ></div>
                  </div>
                  <span class="text-sm text-gray-600 min-w-[50px]">
                    {{ ((item.blocked / item.requests) * 100).toFixed(1) }}%
                  </span>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {
  ChartBarIcon,
  ShieldExclamationIcon,
  ChartPieIcon,
  ServerIcon,
  ArrowPathIcon,
} from '@heroicons/vue/24/outline';

definePageMeta({
  middleware: 'auth',
});

const { domains, selectedDomainId } = useDomains();
const domainsStore = useDomainsStore();

const analytics = useAnalytics();
const { overview, timeSeries, topIPs, loading: analyticsLoading } = analytics;

const blockRateLabel = computed(() => {
  return `${(overview.value.blockRate * 100).toFixed(1)}%`;
});

const handleDomainChange = async () => {
  await analytics.fetchAll(selectedDomainId.value || undefined);
};

const handleRefresh = async () => {
  await analytics.fetchAll(selectedDomainId.value || undefined);
};

const getBlockRateBarClass = (rate: number) => {
  if (rate > 0.5) return 'bg-red-500';
  if (rate > 0.2) return 'bg-yellow-500';
  return 'bg-green-500';
};

onMounted(async () => {
  await domainsStore.refresh();
  await analytics.fetchAll(selectedDomainId.value || undefined);
});
</script>
