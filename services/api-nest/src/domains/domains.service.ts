import { Injectable, NotFoundException } from '@nestjs/common';
import { randomUUID } from 'crypto';
import { resolveTxt } from 'dns/promises';
import { DatabaseService } from '../database/database.service';

type RegisterResult = {
  domainId: string;
  dnsToken: string;
  status: 'pending_verification';
};

@Injectable()
export class DomainsService {
  constructor(private readonly db: DatabaseService) {}

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
      return { id, status: 'active', verified: true };
    }

    return { id, status: 'pending_verification', verified: false };
  }

  async countDomains(): Promise<number> {
    const result = await this.db.query<{ count: string }>('SELECT COUNT(*)::text AS count FROM domains');
    return Number(result.rows[0]?.count ?? 0);
  }
}
