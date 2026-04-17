<template>
  <main class="page">
    <header class="header">
      <h1>Shield Proxy Dashboard</h1>
      <p>Domain management, traffic analytics, and request insights.</p>
    </header>

    <section class="panel auth">
      <h2>Auth</h2>
      <form class="row" @submit.prevent="onLogin">
        <input v-model="authForm.email" type="email" placeholder="Email" required />
        <input v-model="authForm.password" type="password" placeholder="Password" required />
        <button :disabled="auth.loading" type="submit">Login</button>
      </form>
      <p v-if="auth.email" class="hint">Signed in as {{ auth.email }}</p>
    </section>

    <section class="panel domains">
      <h2>Domain management</h2>
      <form class="row" @submit.prevent="onAddDomain">
        <input v-model="domainForm.clientName" placeholder="Client name" required />
        <input v-model="domainForm.domain" placeholder="example.com" required />
        <input v-model="domainForm.upstreamUrl" placeholder="https://origin.example.com" required />
        <button :disabled="domains.loading" type="submit">Add domain</button>
      </form>

      <div class="row">
        <label for="domain-select">Active domain</label>
        <select id="domain-select" v-model="domains.selectedDomainId" @change="refreshAnalytics">
          <option v-for="item in domains.items" :key="item.id" :value="item.id">
            {{ item.domain }} ({{ item.status }})
          </option>
        </select>
      </div>
    </section>

    <section class="cards">
      <article class="card">
        <h3>Total requests</h3>
        <p>{{ overview.totalRequests }}</p>
      </article>
      <article class="card">
        <h3>Blocked requests</h3>
        <p>{{ overview.blockedRequests }}</p>
      </article>
      <article class="card">
        <h3>Block rate</h3>
        <p>{{ blockRateLabel }}</p>
      </article>
      <article class="card">
        <h3>Tracked Top IPs</h3>
        <p>{{ topIPs.items.length }}</p>
      </article>
    </section>

    <section class="panel chart">
      <h2>Traffic chart (last {{ timeSeries.hours }}h)</h2>
      <ul class="bars">
        <li v-for="point in timeSeries.points" :key="point.bucket">
          <span class="bucket">{{ point.bucket.slice(11, 16) }}</span>
          <div class="bar-wrap">
            <div class="bar total" :style="{ width: barWidth(point.totalRequests) }"></div>
            <div class="bar blocked" :style="{ width: barWidth(point.blockedRequests) }"></div>
          </div>
          <span class="value">{{ point.totalRequests }}</span>
        </li>
      </ul>
    </section>

    <section class="panel insights">
      <h2>Traffic insights: top IPs</h2>
      <table>
        <thead>
          <tr>
            <th>IP</th>
            <th>Requests</th>
            <th>Blocked</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in topIPs.items" :key="item.ip">
            <td>{{ item.ip }}</td>
            <td>{{ item.requests }}</td>
            <td>{{ item.blocked }}</td>
          </tr>
        </tbody>
      </table>
    </section>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive } from 'vue';
import { storeToRefs } from 'pinia';
import { useAuthStore } from '../stores/auth';
import { useDomainsStore } from '../stores/domains';

type AnalyticsOverview = {
  totalRequests: number;
  blockedRequests: number;
  blockRate: number;
};

type TimeSeries = {
  hours: number;
  points: Array<{ bucket: string; totalRequests: number; blockedRequests: number }>;
};

type TopIPs = {
  items: Array<{ ip: string; requests: number; blocked: number }>;
};

const auth = useAuthStore();
const domains = useDomainsStore();
const { selectedDomainId } = storeToRefs(domains);

const authForm = reactive({ email: '', password: '' });
const domainForm = reactive({ clientName: 'default-client', domain: '', upstreamUrl: '' });

const overview = reactive<AnalyticsOverview>({ totalRequests: 0, blockedRequests: 0, blockRate: 0 });
const timeSeries = reactive<TimeSeries>({ hours: 24, points: [] });
const topIPs = reactive<TopIPs>({ items: [] });

const blockRateLabel = computed(() => `${(overview.blockRate * 100).toFixed(1)}%`);

function barWidth(value: number): string {
  const max = Math.max(1, ...timeSeries.points.map((point) => point.totalRequests));
  return `${Math.round((value / max) * 100)}%`;
}

async function refreshAnalytics() {
  const query = selectedDomainId.value ? { domainId: selectedDomainId.value } : undefined;
  const [overviewResult, seriesResult, topResult] = await Promise.all([
    $fetch<AnalyticsOverview>('/api/analytics/overview', { query }),
    $fetch<TimeSeries>('/api/analytics/time-series', { query: { ...query, hours: 24 } }),
    $fetch<TopIPs>('/api/analytics/top-ips', { query: { ...query, limit: 10 } }),
  ]);

  overview.totalRequests = overviewResult.totalRequests;
  overview.blockedRequests = overviewResult.blockedRequests;
  overview.blockRate = overviewResult.blockRate;

  timeSeries.hours = seriesResult.hours;
  timeSeries.points = seriesResult.points;

  topIPs.items = topResult.items;
}

async function onLogin() {
  await auth.login(authForm);
}

async function onAddDomain() {
  await domains.addDomain(domainForm);
  domainForm.domain = '';
  domainForm.upstreamUrl = '';
  await refreshAnalytics();
}

onMounted(async () => {
  await domains.refresh();
  await refreshAnalytics();
});
</script>

<style scoped>
.page { max-width: 1080px; margin: 0 auto; padding: 1rem; font-family: Arial, sans-serif; }
.header { margin-bottom: 1rem; }
.panel { border: 1px solid #ddd; border-radius: 8px; padding: 1rem; margin-bottom: 1rem; }
.row { display: flex; gap: 0.5rem; flex-wrap: wrap; align-items: center; margin-bottom: 0.5rem; }
.cards { display: grid; grid-template-columns: repeat(auto-fit, minmax(180px, 1fr)); gap: 0.75rem; margin-bottom: 1rem; }
.card { border: 1px solid #ddd; border-radius: 8px; padding: 0.75rem; }
.card p { font-size: 1.25rem; margin: 0.25rem 0 0; }
.hint { color: #0a7c2f; margin: 0; }
.bars { list-style: none; padding: 0; margin: 0; display: grid; gap: 0.4rem; }
.bars li { display: grid; grid-template-columns: 44px 1fr 52px; gap: 0.5rem; align-items: center; }
.bar-wrap { position: relative; background: #f2f4f7; border-radius: 6px; min-height: 10px; overflow: hidden; }
.bar { height: 10px; }
.bar.total { background: #2563eb; opacity: 0.8; }
.bar.blocked { background: #dc2626; opacity: 0.75; margin-top: -10px; }
.bucket, .value { font-size: 0.8rem; color: #444; }
table { width: 100%; border-collapse: collapse; }
th, td { border-bottom: 1px solid #eee; text-align: left; padding: 0.4rem; }
</style>
