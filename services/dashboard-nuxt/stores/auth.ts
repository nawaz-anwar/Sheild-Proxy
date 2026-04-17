import { defineStore } from 'pinia';

type AuthState = {
  clientId: string | null;
  token: string | null;
  email: string | null;
  loading: boolean;
  error: string | null;
};

export const useAuthStore = defineStore('auth', {
  state: (): AuthState => ({
    clientId: null,
    token: null,
    email: null,
    loading: false,
    error: null,
  }),
  
  getters: {
    isAuthenticated: (state) => !!state.token,
    user: (state) => state.email ? { email: state.email, clientId: state.clientId } : null,
  },
  
  actions: {
    async register(input: { name: string; email: string; password: string }) {
      this.loading = true;
      this.error = null;
      try {
        const result = await $fetch<{ clientId: string; token: string }>('/api/auth/register', { 
          method: 'POST', 
          body: input 
        });
        this.clientId = result.clientId;
        this.token = result.token;
        this.email = input.email;
        return result;
      } catch (error: any) {
        this.error = error.data?.message || 'Registration failed';
        throw error;
      } finally {
        this.loading = false;
      }
    },
    
    async login(input: { email: string; password: string }) {
      this.loading = true;
      this.error = null;
      try {
        const result = await $fetch<{ clientId: string; token: string }>('/api/auth/login', { 
          method: 'POST', 
          body: input 
        });
        this.clientId = result.clientId;
        this.token = result.token;
        this.email = input.email;
        return result;
      } catch (error: any) {
        this.error = error.data?.message || 'Login failed';
        throw error;
      } finally {
        this.loading = false;
      }
    },
    
    logout() {
      this.clientId = null;
      this.token = null;
      this.email = null;
      this.error = null;
      navigateTo('/login');
    },
    
    clearError() {
      this.error = null;
    },
  },
  
  persist: {
    storage: persistedState.localStorage,
  },
});
