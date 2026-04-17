import { Injectable } from '@nestjs/common';
import { DatabaseService } from '../database/database.service';

type OverviewRow = { total: string; blocked: string };
type TimeSeriesRow = { bucket: string; total: string; blocked: string };
type TopIPRow = { ip: string; requests: string; blocked: string };

@Injectable()
export class AnalyticsService {
  constructor(private readonly db: DatabaseService) {}

  async overview(domainId?: string) {
    const { where, params } = this.whereDomain(domainId);
    const result = await this.db.query<OverviewRow>(
      `SELECT
         COUNT(*)::text AS total,
         COUNT(*) FILTER (WHERE status_code >= 400)::text AS blocked
       FROM request_logs ${where}`,
      params,
    );
    const row = result.rows[0] ?? { total: '0', blocked: '0' };
    const totalRequests = Number(row.total);
    const blockedRequests = Number(row.blocked);

    return {
      domainId: domainId ?? null,
      totalRequests,
      blockedRequests,
      blockRate: totalRequests > 0 ? blockedRequests / totalRequests : 0,
    };
  }

  async timeSeries(domainId?: string, hours = 24) {
    const boundedHours = Number.isFinite(hours) ? Math.min(Math.max(Math.floor(hours), 1), 168) : 24;
    const { where, params } = this.whereDomain(domainId, 3);

    const rows = await this.db.query<TimeSeriesRow>(
      `WITH buckets AS (
          SELECT generate_series(
            date_trunc('hour', NOW() - make_interval(hours => $1::int - 1)),
            date_trunc('hour', NOW()),
            interval '1 hour'
          ) AS bucket
        ),
        stats AS (
          SELECT
            date_trunc('hour', created_at) AS bucket,
            COUNT(*)::text AS total,
            COUNT(*) FILTER (WHERE status_code >= 400)::text AS blocked
          FROM request_logs
          ${where}
          AND created_at >= NOW() - make_interval(hours => $2::int)
          GROUP BY 1
        )
        SELECT
          to_char(b.bucket AT TIME ZONE 'UTC', 'YYYY-MM-DD\"T\"HH24:MI:SS\"Z\"') AS bucket,
          COALESCE(s.total, '0') AS total,
          COALESCE(s.blocked, '0') AS blocked
        FROM buckets b
        LEFT JOIN stats s ON s.bucket = b.bucket
        ORDER BY b.bucket ASC`,
      [boundedHours, boundedHours, ...params],
    );

    return {
      domainId: domainId ?? null,
      hours: boundedHours,
      points: rows.rows.map((row) => ({
        bucket: row.bucket,
        totalRequests: Number(row.total),
        blockedRequests: Number(row.blocked),
      })),
    };
  }

  async topIPs(domainId?: string, limit = 10) {
    const boundedLimit = Number.isFinite(limit) ? Math.min(Math.max(Math.floor(limit), 1), 100) : 10;
    const { where, params } = this.whereDomain(domainId);
    const rows = await this.db.query<TopIPRow>(
      `SELECT
         COALESCE(NULLIF(remote_ip, ''), 'unknown') AS ip,
         COUNT(*)::text AS requests,
         COUNT(*) FILTER (WHERE status_code >= 400)::text AS blocked
       FROM request_logs
       ${where}
       GROUP BY 1
       ORDER BY COUNT(*) DESC, ip ASC
       LIMIT $${params.length + 1}`,
      [...params, boundedLimit],
    );
    return {
      domainId: domainId ?? null,
      limit: boundedLimit,
      items: rows.rows.map((row) => ({
        ip: row.ip,
        requests: Number(row.requests),
        blocked: Number(row.blocked),
      })),
    };
  }

  private whereDomain(domainId?: string, offset = 1): { where: string; params: string[] } {
    if (!domainId) {
      return { where: 'WHERE TRUE', params: [] };
    }
    return { where: `WHERE domain_id = $${offset}`, params: [domainId] };
  }
}
