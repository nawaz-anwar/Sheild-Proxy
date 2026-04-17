export const useAuth = () => {
  const store = useAuthStore();
  
  return {
    isAuthenticated: computed(() => store.isAuthenticated),
    user: computed(() => store.user),
    loading: computed(() => store.loading),
    error: computed(() => store.error),
    login: store.login,
    register: store.register,
    logout: store.logout,
    clearError: store.clearError,
  };
};
