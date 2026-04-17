import { readBody } from 'h3';
import { backendRequest } from '../../../utils/backend';

export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id');
  if (!id) {
    throw createError({ statusCode: 400, statusMessage: 'domain id is required' });
  }
  const body = await readBody<{ rules: Array<{ pathPrefix: string; action: 'proxy' | 'block' | 'challenge' }> }>(event);
  return await backendRequest(event, `/domains/${id}/rules`, { method: 'PUT', body });
});
