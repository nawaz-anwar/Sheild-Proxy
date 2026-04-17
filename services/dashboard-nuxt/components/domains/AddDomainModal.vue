<template>
  <UiModal
    :show="show"
    title="Add New Domain"
    confirm-text="Add Domain"
    :show-cancel="true"
    :loading="loading"
    :icon="GlobeAltIcon"
    icon-color="text-primary-600"
    icon-bg-color="bg-primary-100"
    @close="emit('close')"
    @confirm="handleSubmit"
  >
    <form @submit.prevent="handleSubmit" class="space-y-4 text-left">
      <div v-if="error" class="p-3 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm">
        {{ error }}
      </div>

      <div>
        <label for="clientName" class="block text-sm font-medium text-gray-700 mb-1">
          Client Name
        </label>
        <input
          id="clientName"
          v-model="form.clientName"
          type="text"
          required
          class="input-field"
          placeholder="My Company"
        />
      </div>

      <div>
        <label for="domain" class="block text-sm font-medium text-gray-700 mb-1">
          Domain
        </label>
        <input
          id="domain"
          v-model="form.domain"
          type="text"
          required
          class="input-field"
          placeholder="example.com"
        />
        <p class="mt-1 text-xs text-gray-500">
          The domain that will be protected by ShieldProxy
        </p>
      </div>

      <div>
        <label for="upstreamUrl" class="block text-sm font-medium text-gray-700 mb-1">
          Upstream URL
        </label>
        <input
          id="upstreamUrl"
          v-model="form.upstreamUrl"
          type="url"
          required
          class="input-field"
          placeholder="https://origin.example.com"
        />
        <p class="mt-1 text-xs text-gray-500">
          The origin server where traffic will be proxied
        </p>
      </div>
    </form>
  </UiModal>
</template>

<script setup lang="ts">
import { GlobeAltIcon } from '@heroicons/vue/24/outline';

const props = defineProps<{
  show: boolean;
  loading?: boolean;
  error?: string;
}>();

const emit = defineEmits<{
  close: [];
  submit: [data: { clientName: string; domain: string; upstreamUrl: string }];
}>();

const form = reactive({
  clientName: '',
  domain: '',
  upstreamUrl: '',
});

const handleSubmit = () => {
  emit('submit', { ...form });
};

watch(() => props.show, (newVal) => {
  if (!newVal) {
    form.clientName = '';
    form.domain = '';
    form.upstreamUrl = '';
  }
});
</script>
