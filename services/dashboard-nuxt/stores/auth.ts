import { defineStore } from 'pinia';

type AuthState = {
  clientId: string | null;
  token: string | null;
  email: string | null;
  loading: boolean;
};

export const useAuthStore = defineStore('auth', {
  state: (): AuthState => ({
    clientId: null,
    token: null,
    email: null,
    loading: false,
  }),
  actions: {
    async register(input: { name: string; email: string; password: string }) {
      this.loading = true;
      try {
        const result = await $fetch<{ clientId: string; token: string }>('/api/auth/register', { method: 'POST', body: input });
        this.clientId = result.clientId;
        this.token = result.token;
        this.email = input.email;
        return result;
      } finally {
        this.loading = false;
      }
    },
    async login(input: { email: string; password: string }) {
      this.loading = true;
      try {
        const result = await $fetch<{ clientId: string; token: string }>('/api/auth/login', { method: 'POST', body: input });
        this.clientId = result.clientId;
        this.token = result.token;
        this.email = input.email;
        return result;
      } finally {
        this.loading = false;
      }
    },
    logout() {
      this.clientId = null;
      this.token = null;
      this.email = null;
    },
  },
});
