import { readBody } from 'h3';
import { backendRequest } from '../../utils/backend';

export default defineEventHandler(async (event) => {
  const body = await readBody<{ name: string; email: string; password: string }>(event);
  return await backendRequest(event, '/auth/register', { method: 'POST', body });
});
