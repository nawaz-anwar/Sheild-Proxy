import { H3Event } from 'h3';

export function backendBase(event: H3Event): string {
  const config = useRuntimeConfig(event);
  return String(config.backendApiBase ?? 'http://api-nest:3000');
}

export async function backendRequest<T>(event: H3Event, path: string, init: Parameters<typeof $fetch<T>>[1] = {}): Promise<T> {
  const baseURL = backendBase(event);
  return await $fetch<T>(path, {
    baseURL,
    headers: {
      ...(init.headers ?? {}),
    },
    ...init,
  });
}
