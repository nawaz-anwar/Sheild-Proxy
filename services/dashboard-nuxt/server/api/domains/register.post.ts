import { readBody } from 'h3';
import { backendRequest } from '../../utils/backend';

export default defineEventHandler(async (event) => {
  const body = await readBody<{ clientName: string; domain: string; upstreamUrl: string }>(event);
  return await backendRequest(event, '/domains/register', { method: 'POST', body });
});
