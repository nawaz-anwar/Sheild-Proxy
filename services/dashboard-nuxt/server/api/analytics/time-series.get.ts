import { getQuery } from 'h3';
import { backendRequest } from '../../utils/backend';

export default defineEventHandler(async (event) => {
  const query = getQuery(event);
  return await backendRequest(event, '/analytics/time-series', { method: 'GET', query });
});
