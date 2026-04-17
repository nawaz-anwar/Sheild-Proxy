import test from 'node:test';
import assert from 'node:assert/strict';
import { AnalyticsService } from '../src/analytics/analytics.service';

type QueryResult<T> = { rows: T[]; rowCount: number };

class FakeDb {
  async query<T>(text: string): Promise<QueryResult<T>> {
    if (text.includes('FROM request_logs') && text.includes('AS total') && text.includes('AS blocked') && !text.includes('WITH buckets')) {
      return { rows: [{ total: '10', blocked: '2' } as T], rowCount: 1 };
    }
    if (text.includes('WITH buckets')) {
      return {
        rows: [
          { bucket: '2026-04-17T00:00:00Z', total: '3', blocked: '1' } as T,
          { bucket: '2026-04-17T01:00:00Z', total: '0', blocked: '0' } as T,
        ],
        rowCount: 2,
      };
    }
    if (text.includes('COALESCE(NULLIF(remote_ip')) {
      return {
        rows: [
          { ip: '203.0.113.10', requests: '7', blocked: '1' } as T,
          { ip: '198.51.100.5', requests: '2', blocked: '1' } as T,
        ],
        rowCount: 2,
      };
    }
    return { rows: [], rowCount: 0 };
  }
}

test('analytics overview returns block rate', async () => {
  const service = new AnalyticsService(new FakeDb() as never);
  const result = await service.overview();
  assert.equal(result.totalRequests, 10);
  assert.equal(result.blockedRequests, 2);
  assert.equal(result.blockRate, 0.2);
});

test('analytics time-series enforces hour bounds', async () => {
  const service = new AnalyticsService(new FakeDb() as never);
  const result = await service.timeSeries(undefined, 9999);
  assert.equal(result.hours, 168);
  assert.equal(result.points.length, 2);
});

test('analytics top-ips enforces limit bounds', async () => {
  const service = new AnalyticsService(new FakeDb() as never);
  const result = await service.topIPs(undefined, 9999);
  assert.equal(result.limit, 100);
  assert.equal(result.items[0]?.ip, '203.0.113.10');
});
