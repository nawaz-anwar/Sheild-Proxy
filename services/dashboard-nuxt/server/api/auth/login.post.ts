import { readBody } from 'h3';
import { backendRequest } from '../../utils/backend';

export default defineEventHandler(async (event) => {
  const body = await readBody<{ email: string; password: string }>(event);
  return await backendRequest(event, '/auth/login', { method: 'POST', body });
});
