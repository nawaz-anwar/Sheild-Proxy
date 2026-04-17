import { defineStore } from 'pinia';

type DomainRow = {
  id: string;
  domain: string;
  upstreamUrl: string;
  status: string;
  createdAt: string;
};

type Rule = {
  pathPrefix: string;
  action: 'proxy' | 'block' | 'challenge';
};

type DomainsState = {
  items: DomainRow[];
  selectedDomainId: string | null;
  loading: boolean;
  error: string | null;
};

export const useDomainsStore = defineStore('domains', {
  state: (): DomainsState => ({
    items: [],
    selectedDomainId: null,
    loading: false,
    error: null,
  }),
  getters: {
    selectedDomain(state): DomainRow | null {
      return state.items.find((item) => item.id === state.selectedDomainId) ?? null;
    },
  },
  actions: {
    async refresh() {
      this.loading = true;
      this.error = null;
      try {
        this.items = await $fetch<DomainRow[]>('/api/domains');
        if (!this.selectedDomainId && this.items.length > 0) {
          this.selectedDomainId = this.items[0]!.id;
        }
      } catch (error: any) {
        this.error = error.data?.message || 'Failed to fetch domains';
        throw error;
      } finally {
        this.loading = false;
      }
    },
    async addDomain(input: { clientName: string; domain: string; upstreamUrl: string }) {
      this.loading = true;
      this.error = null;
      try {
        await $fetch('/api/domains/register', { method: 'POST', body: input });
        await this.refresh();
      } catch (error: any) {
        this.error = error.data?.message || 'Failed to add domain';
        throw error;
      } finally {
        this.loading = false;
      }
    },
    async updateRules(domainId: string, rules: Rule[]) {
      this.loading = true;
      this.error = null;
      try {
        await $fetch(`/api/domains/${domainId}/rules`, { method: 'PUT', body: { rules } });
      } catch (error: any) {
        this.error = error.data?.message || 'Failed to update rules';
        throw error;
      } finally {
        this.loading = false;
      }
    },
    clearError() {
      this.error = null;
    },
  },
});
