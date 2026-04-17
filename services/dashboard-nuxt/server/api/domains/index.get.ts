import { backendRequest } from '../../utils/backend';

export default defineEventHandler(async (event) => {
  return await backendRequest(event, '/domains', { method: 'GET' });
});
