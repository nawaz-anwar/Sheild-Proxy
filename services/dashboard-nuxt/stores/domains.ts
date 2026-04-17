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
};

export const useDomainsStore = defineStore('domains', {
  state: (): DomainsState => ({
    items: [],
    selectedDomainId: null,
    loading: false,
  }),
  getters: {
    selectedDomain(state): DomainRow | null {
      return state.items.find((item) => item.id === state.selectedDomainId) ?? null;
    },
  },
  actions: {
    async refresh() {
      this.loading = true;
      try {
        this.items = await $fetch<DomainRow[]>('/api/domains');
        if (!this.selectedDomainId && this.items.length > 0) {
          this.selectedDomainId = this.items[0]!.id;
        }
      } finally {
        this.loading = false;
      }
    },
    async addDomain(input: { clientName: string; domain: string; upstreamUrl: string }) {
      await $fetch('/api/domains/register', { method: 'POST', body: input });
      await this.refresh();
    },
    async updateRules(domainId: string, rules: Rule[]) {
      await $fetch(`/api/domains/${domainId}/rules`, { method: 'PUT', body: { rules } });
    },
  },
});
