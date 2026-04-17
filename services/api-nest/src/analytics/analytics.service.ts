import { Injectable } from '@nestjs/common';
import { DatabaseService } from '../database/database.service';

@Injectable()
export class AnalyticsService {
  constructor(private readonly db: DatabaseService) {}

  async overview(domainId?: string) {
    const params: string[] = [];
    let where = '';
    if (domainId) {
      params.push(domainId);
      where = 'WHERE domain_id = $1';
    }

    const total = await this.db.query<{ count: string }>(`SELECT COUNT(*)::text AS count FROM request_logs ${where}`, params);
    const blocked = await this.db.query<{ count: string }>(
      `SELECT COUNT(*)::text AS count FROM request_logs ${where ? `${where} AND` : 'WHERE'} status_code >= 400`,
      params,
    );

    return {
      domainId: domainId ?? null,
      totalRequests: Number(total.rows[0]?.count ?? 0),
      blockedOrErroredRequests: Number(blocked.rows[0]?.count ?? 0),
    };
  }
}
