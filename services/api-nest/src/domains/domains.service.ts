import { Injectable, NotFoundException } from '@nestjs/common';
import { randomUUID } from 'crypto';
import { resolveTxt } from 'dns/promises';
import { DatabaseService } from '../database/database.service';
import { RuleInput, RulesService } from '../rules/rules.service';
import { DomainSyncService } from './domain-sync.service';

type RegisterResult = {
  domainId: string;
  dnsToken: string;
  status: 'pending_verification';
};

@Injectable()
export class DomainsService {
  constructor(
    private readonly db: DatabaseService,
    private readonly rulesService: RulesService,
    private readonly domainSyncService: DomainSyncService,
  ) {}

  async register(clientName: string, domain: string, upstreamUrl: string): Promise<RegisterResult> {
    const clientId = randomUUID();
    const domainId = randomUUID();
    const dnsToken = randomUUID();

    await this.db.query('INSERT INTO clients (id, name) VALUES ($1, $2) ON CONFLICT DO NOTHING', [clientId, clientName]);
    await this.db.query(
      `INSERT INTO domains (id, client_id, domain, upstream_url, dns_token, status)
       VALUES ($1, $2, $3, $4, $5, 'pending_verification')
       ON CONFLICT (domain) DO UPDATE SET upstream_url = EXCLUDED.upstream_url, dns_token = EXCLUDED.dns_token, status = 'pending_verification'`,
      [domainId, clientId, domain, upstreamUrl, dnsToken],
    );
    await this.domainSyncService.publishDomainSync(domain);

    return { domainId, dnsToken, status: 'pending_verification' };
  }

  async status(id: string) {
    const result = await this.db.query<{ id: string; domain: string; status: string; verified_at: Date | null }>(
      'SELECT id, domain, status, verified_at FROM domains WHERE id = $1',
      [id],
    );
    if (result.rowCount === 0) {
      throw new NotFoundException('domain not found');
    }
    return result.rows[0];
  }

  async verifyDns(id: string) {
    const result = await this.db.query<{ domain: string; dns_token: string }>('SELECT domain, dns_token FROM domains WHERE id = $1', [id]);
    if (result.rowCount === 0) {
      throw new NotFoundException('domain not found');
    }

    const { domain, dns_token: token } = result.rows[0];
    const record = `${process.env.DOMAIN_VERIFICATION_PREFIX ?? '_shield-verify'}.${domain}`;
    const txt = await resolveTxt(record).catch(() => [] as string[][]);
    const matches = txt.some((parts) => parts.join('') === token);

    if (matches) {
      await this.db.query("UPDATE domains SET status = 'active', verified_at = NOW() WHERE id = $1", [id]);
      await this.domainSyncService.publishDomainSync(domain);
      return { id, status: 'active', verified: true };
    }

    return { id, status: 'pending_verification', verified: false };
  }

  async countDomains(): Promise<number> {
    const result = await this.db.query<{ count: string }>('SELECT COUNT(*)::text AS count FROM domains');
    return Number(result.rows[0]?.count ?? 0);
  }

  async list() {
    const result = await this.db.query<{ id: string; domain: string; upstream_url: string; status: string; created_at: Date }>(
      'SELECT id, domain, upstream_url, status, created_at FROM domains ORDER BY created_at DESC',
    );
    return result.rows.map((row) => ({
      id: row.id,
      domain: row.domain,
      upstreamUrl: row.upstream_url,
      status: row.status,
      createdAt: row.created_at,
    }));
  }

  async updateRules(domainId: string, rules: RuleInput[]) {
    const result = await this.rulesService.replaceDomainRules(domainId, rules);
    const domain = await this.db.query<{ domain: string }>('SELECT domain FROM domains WHERE id = $1', [domainId]);
    if ((domain.rowCount ?? 0) > 0) {
      await this.domainSyncService.publishDomainSync(domain.rows[0].domain);
    }
    return result;
  }
}
