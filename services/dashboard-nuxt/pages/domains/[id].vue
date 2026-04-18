<template>
  <div class="p-8">
    <div class="mb-6">
      <button @click="navigateTo('/domains')" class="text-sm text-gray-600 hover:text-gray-900 flex items-center gap-2 mb-4">
        <ArrowLeftIcon class="w-4 h-4" />
        Back to Domains
      </button>
      
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold text-gray-900">{{ domain?.domain }}</h1>
          <p class="text-gray-600 mt-1">Domain Configuration & Verification</p>
        </div>
        <div class="flex items-center gap-3">
          <span :class="statusBadgeClass" class="px-3 py-1 rounded-full text-sm font-medium">
            {{ statusLabel }}
          </span>
        </div>
      </div>
    </div>

    <div v-if="loading" class="space-y-6">
      <div class="h-64 bg-gray-100 rounded-xl animate-pulse"></div>
      <div class="h-64 bg-gray-100 rounded-xl animate-pulse"></div>
    </div>

    <div v-else-if="error" class="card">
      <div class="text-center py-12">
        <ExclamationTriangleIcon class="w-12 h-12 text-red-500 mx-auto mb-3" />
        <p class="text-red-600">{{ error }}</p>
      </div>
    </div>

    <div v-else class="space-y-6">
      <!-- Domain Info -->
      <div class="card">
        <h3 class="text-lg font-semibold text-gray-900 mb-4">Domain Information</h3>
        <div class="grid grid-cols-2 gap-4">
          <div>
            <p class="text-sm text-gray-600">Domain</p>
            <p class="text-sm font-medium text-gray-900">{{ domain.domain }}</p>
          </div>
          <div>
            <p class="text-sm text-gray-600">Upstream URL</p>
            <p class="text-sm font-medium text-gray-900">{{ domain.upstreamUrl }}</p>
          </div>
          <div>
            <p class="text-sm text-gray-600">Status</p>
            <p class="text-sm font-medium text-gray-900">{{ domain.status }}</p>
          </div>
          <div>
            <p class="text-sm text-gray-600">Created</p>
            <p class="text-sm font-medium text-gray-900">{{ formatDate(domain.createdAt) }}</p>
          </div>
        </div>
      </div>

      <!-- Step 1: Verify Ownership -->
      <div class="card" :class="{ 'border-2 border-primary-500': !domain.verified }">
        <div class="flex items-start justify-between mb-4">
          <div class="flex items-center gap-3">
            <div :class="domain.verified ? 'bg-green-100' : 'bg-yellow-100'" class="w-10 h-10 rounded-full flex items-center justify-center">
              <CheckCircleIcon v-if="domain.verified" class="w-6 h-6 text-green-600" />
              <span v-else class="text-yellow-600 font-bold">1</span>
            </div>
            <div>
              <h3 class="text-lg font-semibold text-gray-900">Verify Domain Ownership</h3>
              <p class="text-sm text-gray-600">Add a DNS TXT record to verify you own this domain</p>
            </div>
          </div>
          <CheckCircleIcon v-if="domain.verified" class="w-6 h-6 text-green-600" />
        </div>

        <div v-if="!domain.verified" class="space-y-4">
          <div class="bg-blue-50 border border-blue-200 rounded-lg p-4">
            <p class="text-sm font-medium text-blue-900 mb-3">Add this DNS record to your domain:</p>
            
            <div class="space-y-3">
              <div class="bg-white rounded-lg p-3 border border-blue-200">
                <div class="flex items-center justify-between mb-1">
                  <span class="text-xs font-medium text-gray-600">Type</span>
                  <button @click="copyToClipboard('TXT')" class="text-xs text-primary-600 hover:text-primary-700">
                    Copy
                  </button>
                </div>
                <p class="text-sm font-mono text-gray-900">TXT</p>
              </div>

              <div class="bg-white rounded-lg p-3 border border-blue-200">
                <div class="flex items-center justify-between mb-1">
                  <span class="text-xs font-medium text-gray-600">Name</span>
                  <button @click="copyToClipboard(txtRecordHost)" class="text-xs text-primary-600 hover:text-primary-700">
                    Copy
                  </button>
                </div>
                <p class="text-sm font-mono text-gray-900 break-all">{{ txtRecordHost }}</p>
                <p class="text-xs text-gray-500 mt-1">FQDN: {{ txtRecordName }}</p>
              </div>

              <div class="bg-white rounded-lg p-3 border border-blue-200">
                <div class="flex items-center justify-between mb-1">
                  <span class="text-xs font-medium text-gray-600">Value</span>
                  <button @click="copyToClipboard(verificationTxtValue)" class="text-xs text-primary-600 hover:text-primary-700">
                    Copy
                  </button>
                </div>
                <p class="text-sm font-mono text-gray-900 break-all">{{ verificationTxtValue }}</p>
              </div>
            </div>
          </div>

          <div class="bg-gray-50 rounded-lg p-4">
            <p class="text-sm font-medium text-gray-900 mb-2">📚 DNS Provider Instructions</p>
            <ul class="text-sm text-gray-600 space-y-1">
              <li>• <strong>Cloudflare:</strong> DNS → Add Record → Type: TXT</li>
              <li>• <strong>GoDaddy:</strong> DNS Management → Add → Type: TXT</li>
              <li>• <strong>Namecheap:</strong> Advanced DNS → Add New Record → TXT Record</li>
              <li>• <strong>Route53:</strong> Hosted Zones → Create Record → Type: TXT</li>
            </ul>
            <p class="text-xs text-gray-500 mt-3">⏱️ DNS propagation can take 5-30 minutes</p>
          </div>

          <div class="flex items-center gap-3">
            <UiButton @click="handleVerifyDns" :loading="verifying" class="flex-1">
              <span v-if="verifying">Verifying...</span>
              <span v-else>Verify Now</span>
            </UiButton>
            <p class="text-xs text-gray-500">
              Attempts: {{ domain.verificationAttempts || 0 }}/10
            </p>
          </div>

          <div v-if="verificationMessage" :class="verificationMessageClass" class="p-3 rounded-lg text-sm">
            {{ verificationMessage }}
          </div>
        </div>

        <div v-else class="bg-green-50 border border-green-200 rounded-lg p-4">
          <div class="flex items-center gap-2">
            <CheckCircleIcon class="w-5 h-5 text-green-600" />
            <p class="text-sm font-medium text-green-900">Domain verified successfully!</p>
          </div>
          <p class="text-xs text-green-700 mt-1">Verified at {{ formatDate(domain.verifiedAt) }}</p>
        </div>
      </div>

      <!-- Step 2: Connect Domain -->
      <div class="card" :class="{ 'border-2 border-primary-500': domain.verified && !domain.proxyConnected }">
        <div class="flex items-start justify-between mb-4">
          <div class="flex items-center gap-3">
            <div :class="domain.proxyConnected ? 'bg-green-100' : domain.verified ? 'bg-yellow-100' : 'bg-gray-100'" class="w-10 h-10 rounded-full flex items-center justify-center">
              <CheckCircleIcon v-if="domain.proxyConnected" class="w-6 h-6 text-green-600" />
              <span v-else :class="domain.verified ? 'text-yellow-600' : 'text-gray-400'" class="font-bold">2</span>
            </div>
            <div>
              <h3 class="text-lg font-semibold text-gray-900">Connect to ShieldProxy</h3>
              <p class="text-sm text-gray-600">Point your domain to our proxy servers</p>
            </div>
          </div>
          <CheckCircleIcon v-if="domain.proxyConnected" class="w-6 h-6 text-green-600" />
        </div>

        <div v-if="!domain.verified" class="bg-gray-50 border border-gray-200 rounded-lg p-4">
          <p class="text-sm text-gray-600">Complete domain verification first</p>
        </div>

        <div v-else-if="!domain.proxyConnected" class="space-y-4">
          <div class="bg-blue-50 border border-blue-200 rounded-lg p-4">
            <p class="text-sm font-medium text-blue-900 mb-3">Choose one of the following methods:</p>
            
            <div class="space-y-4">
              <div class="bg-white rounded-lg p-4 border border-blue-200">
                <p class="text-sm font-semibold text-gray-900 mb-2">Option A: CNAME Record (Recommended)</p>
                <div class="space-y-2">
                  <div class="flex items-center justify-between">
                    <span class="text-xs text-gray-600">Name:</span>
                    <button @click="copyToClipboard(cnameRecordHost)" class="text-xs text-primary-600">Copy</button>
                  </div>
                  <p class="text-sm font-mono bg-gray-50 p-2 rounded">{{ cnameRecordHost }}</p>
                  
                  <div class="flex items-center justify-between">
                    <span class="text-xs text-gray-600">Value:</span>
                    <button @click="copyToClipboard(proxyCnameTarget)" class="text-xs text-primary-600">Copy</button>
                  </div>
                  <p class="text-sm font-mono bg-gray-50 p-2 rounded">{{ proxyCnameTarget }}</p>
                </div>
              </div>

              <div class="bg-white rounded-lg p-4 border border-blue-200">
                <p class="text-sm font-semibold text-gray-900 mb-2">Option B: A Record</p>
                <div class="space-y-2">
                  <div class="flex items-center justify-between">
                    <span class="text-xs text-gray-600">Name:</span>
                    <button @click="copyToClipboard(domain.domain)" class="text-xs text-primary-600">Copy</button>
                  </div>
                  <p class="text-sm font-mono bg-gray-50 p-2 rounded">{{ domain.domain }}</p>
                  
                  <div class="flex items-center justify-between">
                    <span class="text-xs text-gray-600">Value:</span>
                    <button @click="copyToClipboard(proxyServerIP)" class="text-xs text-primary-600">Copy</button>
                  </div>
                  <p class="text-sm font-mono bg-gray-50 p-2 rounded">{{ proxyServerIP }}</p>
                </div>
              </div>
            </div>
          </div>

          <UiButton @click="handleCheckConnection" :loading="checkingConnection" class="w-full">
            <span v-if="checkingConnection">Checking Connection...</span>
            <span v-else>Check Connection</span>
          </UiButton>

          <div v-if="connectionMessage" :class="connectionMessageClass" class="p-3 rounded-lg text-sm">
            {{ connectionMessage }}
          </div>
        </div>

        <div v-else class="bg-green-50 border border-green-200 rounded-lg p-4">
          <div class="flex items-center gap-2">
            <CheckCircleIcon class="w-5 h-5 text-green-600" />
            <p class="text-sm font-medium text-green-900">Domain connected successfully!</p>
          </div>
          <p class="text-xs text-green-700 mt-1">Connected at {{ formatDate(domain.connectedAt) }}</p>
          <p class="text-xs text-green-700">Method: {{ connectionMethod }}</p>
        </div>
      </div>

      <!-- Status Summary -->
      <div v-if="domain.verified && domain.proxyConnected" class="card bg-green-50 border-2 border-green-200">
        <div class="flex items-center gap-3">
          <div class="w-12 h-12 bg-green-100 rounded-full flex items-center justify-center">
            <CheckCircleIcon class="w-8 h-8 text-green-600" />
          </div>
          <div>
            <h3 class="text-lg font-semibold text-green-900">🎉 Domain is Active!</h3>
            <p class="text-sm text-green-700">Your domain is now protected by ShieldProxy</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {
  ArrowLeftIcon,
  CheckCircleIcon,
  ExclamationTriangleIcon,
} from '@heroicons/vue/24/outline';

definePageMeta({
  middleware: 'auth',
});

const route = useRoute();
const domainId = route.params.id as string;
const config = useRuntimeConfig();

const domain = ref<any>(null);
const loading = ref(true);
const error = ref<string | null>(null);
const verifying = ref(false);
const checkingConnection = ref(false);
const verificationMessage = ref<string | null>(null);
const connectionMessage = ref<string | null>(null);

const proxyCnameTarget = computed(() => config.public.proxyCnameTarget as string);
const proxyServerIP = computed(() => config.public.proxyServerIp as string);
const txtRecordHost = '_shieldproxy';

const txtRecordName = computed(() => {
  if (!domain.value) return '';
  return `_shieldproxy.${domain.value.domain}`;
});

const cnameRecordHost = computed(() => {
  if (!domain.value?.domain) return 'www';
  return domain.value.domain.startsWith('www.') ? domain.value.domain : `www.${domain.value.domain}`;
});

const verificationTxtValue = computed(() => {
  const token = (domain.value?.verificationToken ?? '').trim();
  if (!token) return '';
  return token.startsWith('sp-verify-') ? token : `sp-verify-${token}`;
});

const statusLabel = computed(() => {
  if (!domain.value) return '';
  if (domain.value.proxyConnected) return '🟢 Active';
  if (domain.value.verified) return '🟡 Verified';
  return '🟡 Pending Verification';
});

const statusBadgeClass = computed(() => {
  if (!domain.value) return '';
  if (domain.value.proxyConnected) return 'bg-green-100 text-green-800';
  if (domain.value.verified) return 'bg-yellow-100 text-yellow-800';
  return 'bg-gray-100 text-gray-800';
});

const verificationMessageClass = computed(() => {
  if (!verificationMessage.value) return '';
  if (verificationMessage.value.includes('success')) return 'bg-green-50 border border-green-200 text-green-800';
  return 'bg-red-50 border border-red-200 text-red-800';
});

const connectionMessageClass = computed(() => {
  if (!connectionMessage.value) return '';
  if (connectionMessage.value.includes('success') || connectionMessage.value.includes('connected')) {
    return 'bg-green-50 border border-green-200 text-green-800';
  }
  return 'bg-red-50 border border-red-200 text-red-800';
});

const connectionMethod = computed(() => {
  if (!domain.value?.dnsRecords?.connection) return 'Unknown';
  const method = domain.value.dnsRecords.connection.method;
  return method === 'cname' ? 'CNAME' : 'A Record';
});

const fetchDomain = async () => {
  loading.value = true;
  error.value = null;
  try {
    domain.value = await $fetch(`/api/domains/${domainId}`);
  } catch (e: any) {
    error.value = e.data?.message || 'Failed to load domain';
  } finally {
    loading.value = false;
  }
};

const handleVerifyDns = async () => {
  verifying.value = true;
  verificationMessage.value = null;
  try {
    const result = await $fetch(`/api/domains/${domainId}/verify-dns`, { method: 'POST' });
    if (result.verified) {
      verificationMessage.value = '✅ Domain verified successfully!';
      await fetchDomain();
    } else {
      verificationMessage.value = result.message || 'Verification failed. Please check your DNS records.';
    }
  } catch (e: any) {
    verificationMessage.value = e.data?.message || 'Verification failed';
  } finally {
    verifying.value = false;
  }
};

const handleCheckConnection = async () => {
  checkingConnection.value = true;
  connectionMessage.value = null;
  try {
    const result = await $fetch(`/api/domains/${domainId}/check-connection`, { method: 'POST' });
    if (result.connected) {
      connectionMessage.value = `✅ ${result.message}`;
      await fetchDomain();
    } else {
      connectionMessage.value = result.message || 'Connection check failed';
    }
  } catch (e: any) {
    connectionMessage.value = e.data?.message || 'Connection check failed';
  } finally {
    checkingConnection.value = false;
  }
};

const copyToClipboard = async (text: string) => {
  try {
    await navigator.clipboard.writeText(text);
    // Could add a toast notification here
  } catch (e) {
    console.error('Failed to copy:', e);
  }
};

const formatDate = (date: string | null) => {
  if (!date) return 'N/A';
  return new Date(date).toLocaleString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
};

onMounted(() => {
  fetchDomain();
});
</script>
