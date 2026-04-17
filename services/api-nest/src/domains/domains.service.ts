import { Injectable, Logger, NotFoundException } from '@nestjs/common';
import { randomUUID } from 'crypto';
import { DatabaseService } from '../database/database.service';
import { RuleInput, RulesService } from '../rules/rules.service';
import { DomainSyncService } from './domain-sync.service';
import { DnsVerificationService } from './dns-verification.service';

type RegisterResult = {
  domainId: string;
  status: 'pending_verification';
};

@Injectable()
export class DomainsService {
  private readonly logger = new Logger(DomainsService.name);

  constructor(
    private readonly db: DatabaseService,
    private readonly rulesService: RulesService,
    private readonly domainSyncService: DomainSyncService,
    private readonly dnsVerificationService: DnsVerificationService,
  ) {}

  async register(clientName: string, domain: string, upstreamUrl: string): Promise<RegisterResult> {
    const clientId = randomUUID();
    const domainId = randomUUID();

    await this.db.query('INSERT INTO clients (id, name) VALUES ($1, $2) ON CONFLICT DO NOTHING', [clientId, clientName]);
    
    const result = await this.db.query(
      `INSERT INTO domains (id, client_id, domain, upstream_url, status)
       VALUES ($1, $2, $3, $4, 'pending_verification')
       ON CONFLICT (domain) DO UPDATE 
       SET upstream_url = EXCLUDED.upstream_url, 
           status = 'pending_verification',
           verified = FALSE,
           proxy_connected = FALSE
       RETURNING id`,
      [domainId, clientId, domain, upstreamUrl],
    );

    const finalDomainId = result.rows[0].id;

    // Initiate verification automatically
    await this.dnsVerificationService.initiateDomainVerification(finalDomainId);
    await this.domainSyncService.publishDomainSync(domain);

    return { domainId: finalDomainId, status: 'pending_verification' };
  }

  async getById(id: string) {
    const result = await this.db.query(
      `SELECT id, domain, upstream_url, status, verified, proxy_connected, 
              verification_token, dns_records, verified_at, connected_at, 
              verification_attempts, last_verification_attempt, created_at
       FROM domains WHERE id = $1`,
      [id],
    );
    
    if (result.rowCount === 0) {
      throw new NotFoundException('Domain not found');
    }

    const domain = result.rows[0];
    
    return {
      id: domain.id,
      domain: domain.domain,
      upstreamUrl: domain.upstream_url,
      status: domain.status,
      verified: domain.verified,
      proxyConnected: domain.proxy_connected,
      verificationToken: domain.verification_token,
      dnsRecords: domain.dns_records,
      verifiedAt: domain.verified_at,
      connectedAt: domain.connected_at,
      verificationAttempts: domain.verification_attempts,
      lastVerificationAttempt: domain.last_verification_attempt,
      createdAt: domain.created_at,
    };
  }

  async status(id: string) {
    const result = await this.db.query<{ 
      id: string; 
      domain: string; 
      status: string; 
      verified: boolean;
      proxy_connected: boolean;
      verified_at: Date | null;
      connected_at: Date | null;
    }>(
      'SELECT id, domain, status, verified, proxy_connected, verified_at, connected_at FROM domains WHERE id = $1',
      [id],
    );
    if (result.rowCount === 0) {
      throw new NotFoundException('domain not found');
    }
    return {
      ...result.rows[0],
      proxyConnected: result.rows[0].proxy_connected,
      verifiedAt: result.rows[0].verified_at,
      connectedAt: result.rows[0].connected_at,
    };
  }

  async initiateVerification(id: string) {
    const verification = await this.dnsVerificationService.initiateDomainVerification(id);
    return {
      domainId: id,
      ...verification,
    };
  }

  async verifyDns(id: string) {
    const result = await this.dnsVerificationService.verifyDomain(id);
    
    if (result.verified) {
      const domain = await this.db.query<{ domain: string }>('SELECT domain FROM domains WHERE id = $1', [id]);
      if (domain.rowCount && domain.rowCount > 0) {
        await this.domainSyncService.publishDomainSync(domain.rows[0].domain);
      }
    }

    return {
      domainId: id,
      ...result,
    };
  }

  async checkConnection(id: string) {
    const result = await this.dnsVerificationService.checkProxyConnection(id);
    
    if (result.connected) {
      const domain = await this.db.query<{ domain: string }>('SELECT domain FROM domains WHERE id = $1', [id]);
      if (domain.rowCount && domain.rowCount > 0) {
        await this.domainSyncService.publishDomainSync(domain.rows[0].domain);
      }
    }

    return {
      domainId: id,
      ...result,
    };
  }

  async getVerificationLogs(id: string) {
    const logs = await this.dnsVerificationService.getVerificationLogs(id);
    return { domainId: id, logs };
  }

  async countDomains(): Promise<number> {
    const result = await this.db.query<{ count: string }>('SELECT COUNT(*)::text AS count FROM domains');
    return Number(result.rows[0]?.count ?? 0);
  }

  async list() {
    const result = await this.db.query<{ 
      id: string; 
      domain: string; 
      upstream_url: string; 
      status: string; 
      verified: boolean;
      proxy_connected: boolean;
      created_at: Date;
    }>(
      'SELECT id, domain, upstream_url, status, verified, proxy_connected, created_at FROM domains ORDER BY created_at DESC',
    );
    return result.rows.map((row) => ({
      id: row.id,
      domain: row.domain,
      upstreamUrl: row.upstream_url,
      status: row.status,
      verified: row.verified,
      proxyConnected: row.proxy_connected,
      createdAt: row.created_at,
    }));
  }

  async updateRules(domainId: string, rules: RuleInput[]) {
    const result = await this.rulesService.replaceDomainRules(domainId, rules);
    const domain = await this.db.query<{ domain: string }>('SELECT domain FROM domains WHERE id = $1', [domainId]);
    if ((domain.rowCount ?? 0) > 0) {
      await this.domainSyncService.publishDomainSync(domain.rows[0].domain);
    } else {
      this.logger.warn(`domain sync skipped after rules update: domain not found id=${domainId}`);
    }
    return result;
  }
}
