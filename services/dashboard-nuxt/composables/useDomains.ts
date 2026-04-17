export const useDomains = () => {
  const store = useDomainsStore();
  
  const addDomain = async (data: { clientName: string; domain: string; upstreamUrl: string }) => {
    await store.addDomain(data);
  };
  
  const refreshDomains = async () => {
    await store.refresh();
  };
  
  const updateRules = async (domainId: string, rules: any[]) => {
    await store.updateRules(domainId, rules);
  };
  
  return {
    domains: computed(() => store.items),
    selectedDomain: computed(() => store.selectedDomain),
    selectedDomainId: computed({
      get: () => store.selectedDomainId,
      set: (value) => { store.selectedDomainId = value; },
    }),
    loading: computed(() => store.loading),
    addDomain,
    refreshDomains,
    updateRules,
  };
};
