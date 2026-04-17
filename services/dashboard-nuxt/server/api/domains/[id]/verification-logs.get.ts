import { backendRequest } from '../../../utils/backend';

export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id');
  return await backendRequest(event, `/domains/${id}/verification-logs`, { method: 'GET' });
});
