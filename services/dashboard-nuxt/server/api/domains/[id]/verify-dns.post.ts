import { backendRequest } from '../../../utils/backend';

export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id');
  if (!id) {
    throw createError({ statusCode: 400, statusMessage: 'domain id is required' });
  }
  return await backendRequest(event, `/domains/${id}/verify-dns`, { method: 'POST' });
});
