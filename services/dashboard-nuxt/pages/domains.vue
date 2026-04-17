<template>
  <div class="p-8">
    <div class="mb-8">
      <h1 class="text-2xl font-bold text-gray-900">Domain Management</h1>
      <p class="text-gray-600 mt-1">Manage your protected domains and configurations</p>
    </div>

    <DomainsDomainTable
      :domains="domains"
      :loading="loading"
      @add="showAddModal = true"
      @select="handleSelectDomain"
    />

    <DomainsAddDomainModal
      :show="showAddModal"
      :loading="loading"
      :error="error"
      @close="showAddModal = false"
      @submit="handleAddDomain"
    />
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  middleware: 'auth',
});

const domainsStore = useDomainsStore();
const { domains, loading } = useDomains();
const error = computed(() => domainsStore.error);

const showAddModal = ref(false);

const handleAddDomain = async (data: { clientName: string; domain: string; upstreamUrl: string }) => {
  try {
    await domainsStore.addDomain(data);
    showAddModal.value = false;
  } catch (error) {
    // Error is handled in store
  }
};

const handleSelectDomain = (id: string) => {
  domainsStore.selectedDomainId = id;
  navigateTo('/');
};

onMounted(async () => {
  await domainsStore.refresh();
});
</script>
