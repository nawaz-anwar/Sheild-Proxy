<template>
  <div class="stat-card">
    <div class="flex items-start justify-between">
      <div class="flex-1">
        <p class="text-sm font-medium text-gray-600">{{ title }}</p>
        <p class="mt-2 text-3xl font-bold text-gray-900">
          <span v-if="loading" class="inline-block w-20 h-8 bg-gray-200 rounded animate-pulse"></span>
          <span v-else>{{ value }}</span>
        </p>
        <p v-if="change" class="mt-2 text-sm" :class="changeColor">
          <span class="font-medium">{{ change }}</span>
          <span class="text-gray-600 ml-1">{{ changeLabel }}</span>
        </p>
      </div>
      <div v-if="icon" class="flex-shrink-0">
        <div :class="iconBgColor" class="w-12 h-12 rounded-lg flex items-center justify-center">
          <component :is="icon" :class="iconColor" class="w-6 h-6" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{
  title: string;
  value: string | number;
  change?: string;
  changeLabel?: string;
  changeType?: 'positive' | 'negative' | 'neutral';
  icon?: any;
  iconColor?: string;
  iconBgColor?: string;
  loading?: boolean;
}>();

const changeColor = computed(() => {
  if (!props.changeType) return 'text-gray-600';
  return {
    positive: 'text-green-600',
    negative: 'text-red-600',
    neutral: 'text-gray-600',
  }[props.changeType];
});
</script>
