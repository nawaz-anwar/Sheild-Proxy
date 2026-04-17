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

export const useAnalytics = () => {
  const overview = ref<AnalyticsOverview>({ totalRequests: 0, blockedRequests: 0, blockRate: 0 });
  const timeSeries = ref<TimeSeries>({ hours: 24, points: [] });
  const topIPs = ref<TopIPs>({ items: [] });
  const loading = ref(false);
  const error = ref<string | null>(null);
  
  const fetchOverview = async (domainId?: string) => {
    loading.value = true;
    error.value = null;
    try {
      const query = domainId ? { domainId } : undefined;
      overview.value = await $fetch<AnalyticsOverview>('/api/analytics/overview', { query });
    } catch (e: any) {
      error.value = e.message || 'Failed to fetch overview';
      throw e;
    } finally {
      loading.value = false;
    }
  };
  
  const fetchTimeSeries = async (domainId?: string, hours: number = 24) => {
    loading.value = true;
    error.value = null;
    try {
      const query = domainId ? { domainId, hours } : { hours };
      timeSeries.value = await $fetch<TimeSeries>('/api/analytics/time-series', { query });
    } catch (e: any) {
      error.value = e.message || 'Failed to fetch time series';
      throw e;
    } finally {
      loading.value = false;
    }
  };
  
  const fetchTopIPs = async (domainId?: string, limit: number = 10) => {
    loading.value = true;
    error.value = null;
    try {
      const query = domainId ? { domainId, limit } : { limit };
      topIPs.value = await $fetch<TopIPs>('/api/analytics/top-ips', { query });
    } catch (e: any) {
      error.value = e.message || 'Failed to fetch top IPs';
      throw e;
    } finally {
      loading.value = false;
    }
  };
  
  const fetchAll = async (domainId?: string) => {
    loading.value = true;
    error.value = null;
    try {
      await Promise.all([
        fetchOverview(domainId),
        fetchTimeSeries(domainId),
        fetchTopIPs(domainId),
      ]);
    } catch (e: any) {
      error.value = e.message || 'Failed to fetch analytics';
    } finally {
      loading.value = false;
    }
  };
  
  return {
    overview: computed(() => overview.value),
    timeSeries: computed(() => timeSeries.value),
    topIPs: computed(() => topIPs.value),
    loading: computed(() => loading.value),
    error: computed(() => error.value),
    fetchOverview,
    fetchTimeSeries,
    fetchTopIPs,
    fetchAll,
  };
};
