import { Injectable, Logger } from '@nestjs/common';
import { DatabaseService } from '../database/database.service';
import * as dns from 'dns/promises';
import { randomUUID } from 'crypto';

export interface DnsVerificationResult {
  success: boolean;
  verified: boolean;
  message: string;
  records?: string[];
  error?: string;
}

@Injectable()
export class DnsVerificationService {
  private readonly logger = new Logger(DnsVerificationService.name);
  private readonly MAX_VERIFICATION_ATTEMPTS = 10;
  private readonly VERIFICATION_PREFIX = 'sp-verify-';
  private readonly TXT_RECORD_NAME = '_shieldproxy';
  private readonly DNS_MAX_RETRIES = Number(process.env.DNS_VERIFICATION_RETRIES ?? 3);
  private readonly DNS_TIMEOUT_MS = Number(process.env.DNS_VERIFICATION_TIMEOUT_MS ?? 5000);
  private readonly DNS_RETRY_DELAY_MS = Number(process.env.DNS_VERIFICATION_RETRY_DELAY_MS ?? 1000);
  private readonly proxyCnameTarget = this.normalizeDnsName(
    process.env.PROXY_CNAME_TARGET ?? 'proxy.shieldproxy.com',
  );

  constructor(private readonly db: DatabaseService) {}

  generateVerificationToken(): string {
    return randomUUID();
  }

  getTxtRecordName(domain: string): string {
    return `${this.TXT_RECORD_NAME}.${domain}`;
  }

  async initiateDomainVerification(domainId: string): Promise<{
    verificationToken: string;
    txtRecordName: string;
    txtRecordValue: string;
  }> {
    const token = this.generateVerificationToken();
    
    const result = await this.db.query(
      `UPDATE domains 
       SET verification_token = $1, 
           status = 'pending_verification',
           verification_attempts = 0
       WHERE id = $2
       RETURNING domain`,
      [token, domainId]
    );

    if (result.rows.length === 0) {
      throw new Error('Domain not found');
    }

    const domain = result.rows[0].domain;
    const txtRecordName = this.getTxtRecordName(domain);

    await this.logVerification(domainId, 'initiated', 'success', {
      token,
      txtRecordName,
    });

    return {
      verificationToken: token,
      txtRecordName,
      txtRecordValue: this.buildTxtVerificationValue(token),
    };
  }

  async verifyDomain(domainId: string): Promise<DnsVerificationResult> {
    const domainResult = await this.db.query(
      `SELECT id, domain, verification_token, verified, verification_attempts, status
       FROM domains 
       WHERE id = $1`,
      [domainId]
    );

    if (domainResult.rows.length === 0) {
      return {
        success: false,
        verified: false,
        message: 'Domain not found',
        error: 'DOMAIN_NOT_FOUND',
      };
    }

    const domainData = domainResult.rows[0];

    if (domainData.verified) {
      return {
        success: true,
        verified: true,
        message: 'Domain already verified',
      };
    }

    if (domainData.verification_attempts >= this.MAX_VERIFICATION_ATTEMPTS) {
      await this.logVerification(domainId, 'dns_txt', 'failed', {
        reason: 'max_attempts_reached',
      });
      return {
        success: false,
        verified: false,
        message: 'Maximum verification attempts reached',
        error: 'MAX_ATTEMPTS_REACHED',
      };
    }

    // Increment verification attempts
    await this.db.query(
      `UPDATE domains 
       SET verification_attempts = verification_attempts + 1,
           last_verification_attempt = NOW()
       WHERE id = $1`,
      [domainId]
    );

    // Perform DNS verification
    const verificationResult = await this.checkDnsTxtRecord(
      domainData.domain,
      domainData.verification_token
    );

    if (verificationResult.verified) {
      await this.db.query(
        `UPDATE domains 
         SET verified = TRUE,
             verified_at = NOW(),
             status = 'verified',
             dns_records = $1
         WHERE id = $2`,
        [JSON.stringify({ txt: verificationResult.records }), domainId]
      );

      await this.logVerification(domainId, 'dns_txt', 'success', {
        records: verificationResult.records,
      });
    } else {
      await this.logVerification(domainId, 'dns_txt', 'failed', {
        reason: verificationResult.message,
        records: verificationResult.records,
      });
    }

    return verificationResult;
  }

  private async checkDnsTxtRecord(
    domain: string,
    expectedToken: string | null
  ): Promise<DnsVerificationResult> {
    const txtRecordName = this.getTxtRecordName(domain);
    const expectedTxtValue = this.buildTxtVerificationValue(expectedToken);

    try {
      this.logger.log(`Checking TXT record for ${txtRecordName}`);
      
      const records = await this.resolveWithRetry<readonly string[][]>(
        `TXT ${txtRecordName}`,
        () => dns.resolveTxt(txtRecordName),
      );
      const flatRecords = records.map((r) => r.join(''));

      this.logger.log(`Found TXT records: ${JSON.stringify(flatRecords)}`);

      if (!expectedTxtValue) {
        return {
          success: false,
          verified: false,
          message: 'Verification token is not available. Re-initiate verification and try again.',
          records: flatRecords,
          error: 'TOKEN_MISSING',
        };
      }

      const hasValidToken = flatRecords.some((record) => this.normalizeTxtValue(record) === expectedTxtValue);

      if (hasValidToken) {
        return {
          success: true,
          verified: true,
          message: 'Domain verified successfully',
          records: flatRecords,
        };
      }

      return {
        success: false,
        verified: false,
        message: 'Verification token not found in DNS records',
        records: flatRecords,
        error: 'TOKEN_MISMATCH',
      };
    } catch (error: any) {
      this.logger.warn(`DNS lookup failed for ${txtRecordName}: ${error.message}`);

      if (error.code === 'ENOTFOUND' || error.code === 'ENODATA') {
        return {
          success: false,
          verified: false,
          message: 'DNS record not found. Please ensure the TXT record is added and DNS has propagated.',
          error: 'DNS_NOT_FOUND',
        };
      }

      return {
        success: false,
        verified: false,
        message: `DNS verification failed: ${error.message}`,
        error: 'DNS_ERROR',
      };
    }
  }

  async checkProxyConnection(domainId: string): Promise<{
    connected: boolean;
    method?: 'cname' | 'a_record';
    records?: any;
    message: string;
  }> {
    const domainResult = await this.db.query(
      `SELECT domain, verified FROM domains WHERE id = $1`,
      [domainId]
    );

    if (domainResult.rows.length === 0) {
      return { connected: false, message: 'Domain not found' };
    }

    const { domain, verified } = domainResult.rows[0];

    if (!verified) {
      return { connected: false, message: 'Domain must be verified first' };
    }

    try {
      const cnameHostsToCheck = [domain, `www.${domain}`];
      for (const cnameHost of cnameHostsToCheck) {
        try {
          const cnameRecords = await this.resolveWithRetry<string[]>(
            `CNAME ${cnameHost}`,
            () => dns.resolveCname(cnameHost),
          );
          const hasProxyCname = cnameRecords.some(
            (record) => this.normalizeDnsName(record) === this.proxyCnameTarget,
          );

          if (hasProxyCname) {
            await this.updateProxyConnection(domainId, true, 'cname', {
              host: cnameHost,
              values: cnameRecords,
              target: this.proxyCnameTarget,
            });
            return {
              connected: true,
              method: 'cname',
              records: { host: cnameHost, values: cnameRecords },
              message: `Domain connected via CNAME (${cnameHost} -> ${this.proxyCnameTarget})`,
            };
          }
        } catch (cnameError: any) {
          if (cnameError.code !== 'ENODATA' && cnameError.code !== 'ENOTFOUND') {
            throw cnameError;
          }
        }
      }

      // Check A record
      const aRecords = await this.resolveWithRetry<string[]>(
        `A ${domain}`,
        () => dns.resolve4(domain),
      );
      const proxyServerIPs = this.getConfiguredProxyIPs();
      const hasProxyARecord = aRecords.some((record) => proxyServerIPs.has(record));

      if (hasProxyARecord) {
        await this.updateProxyConnection(domainId, true, 'a_record', {
          host: domain,
          values: aRecords,
          acceptedProxyIps: Array.from(proxyServerIPs),
        });
        return {
          connected: true,
          method: 'a_record',
          records: aRecords,
          message: 'Domain connected via A record',
        };
      }

      return {
        connected: false,
        message:
          proxyServerIPs.size > 0
            ? `Domain is not pointing to ShieldProxy (expected CNAME target "${this.proxyCnameTarget}" or A record IP in [${Array.from(proxyServerIPs).join(', ')}])`
            : `Domain is not pointing to ShieldProxy (expected CNAME target "${this.proxyCnameTarget}")`,
      };
    } catch (error: any) {
      this.logger.error(`Proxy connection check failed: ${error.message}`);
      return {
        connected: false,
        message: `Connection check failed: ${error.message}`,
      };
    }
  }

  private async updateProxyConnection(
    domainId: string,
    connected: boolean,
    method: string,
    records: any
  ): Promise<void> {
    await this.db.query(
      `UPDATE domains 
       SET proxy_connected = $1,
           connected_at = NOW(),
           status = 'active',
           dns_records = jsonb_set(
             COALESCE(dns_records, '{}'::jsonb),
             '{connection}',
             $2::jsonb
           )
       WHERE id = $3`,
      [connected, JSON.stringify({ method, records }), domainId]
    );

    await this.logVerification(domainId, 'proxy_connection', 'success', {
      method,
      records,
    });
  }

  private async logVerification(
    domainId: string,
    verificationType: string,
    status: string,
    details: any
  ): Promise<void> {
    await this.db.query(
      `INSERT INTO domain_verification_logs (domain_id, verification_type, status, details)
       VALUES ($1, $2, $3, $4)`,
      [domainId, verificationType, status, JSON.stringify(details)]
    );
  }

  async getVerificationLogs(domainId: string, limit: number = 10) {
    const result = await this.db.query(
      `SELECT * FROM domain_verification_logs 
       WHERE domain_id = $1 
       ORDER BY created_at DESC 
       LIMIT $2`,
      [domainId, limit]
    );
    return result.rows;
  }

  private buildTxtVerificationValue(token: string | null | undefined): string {
    const raw = (token ?? '').trim();
    if (!raw) {
      return '';
    }
    if (raw.startsWith(this.VERIFICATION_PREFIX)) {
      return raw;
    }
    return `${this.VERIFICATION_PREFIX}${raw}`;
  }

  private normalizeTxtValue(value: string): string {
    return value.trim().replace(/^"+|"+$/g, '');
  }

  private normalizeDnsName(value: string): string {
    return value.trim().toLowerCase().replace(/\.$/, '');
  }

  private getConfiguredProxyIPs(): Set<string> {
    const configured = (process.env.PROXY_SERVER_IPS ?? process.env.PROXY_SERVER_IP ?? '')
      .split(',')
      .map((part) => part.trim())
      .filter((part) => part.length > 0);
    return new Set(configured);
  }

  private async resolveWithRetry<T>(
    label: string,
    resolver: () => Promise<T>,
  ): Promise<T> {
    let lastError: any;
    const maxAttempts = Math.max(1, this.DNS_MAX_RETRIES);
    for (let attempt = 1; attempt <= maxAttempts; attempt++) {
      try {
        return await this.withTimeout(resolver(), this.DNS_TIMEOUT_MS);
      } catch (error: any) {
        lastError = error;
        const retryable = this.isRetryableDnsError(error);
        if (!retryable || attempt === maxAttempts) {
          break;
        }
        this.logger.warn(
          `DNS lookup retry ${attempt}/${maxAttempts} for ${label} failed: ${error.message}`,
        );
        await this.delay(this.DNS_RETRY_DELAY_MS);
      }
    }
    throw lastError;
  }

  private isRetryableDnsError(error: any): boolean {
    const code = String(error?.code ?? '').toUpperCase();
    return [
      'ETIMEOUT',
      'ENOTFOUND',
      'ENODATA',
      'ESERVFAIL',
      'EAI_AGAIN',
      'ECONNREFUSED',
      'EHOSTUNREACH',
      'ENETUNREACH',
    ].includes(code);
  }

  private withTimeout<T>(promise: Promise<T>, timeoutMs: number): Promise<T> {
    return new Promise<T>((resolve, reject) => {
      const timer = setTimeout(() => {
        const timeoutError = new Error(`DNS lookup timeout after ${timeoutMs}ms`) as NodeJS.ErrnoException;
        timeoutError.code = 'ETIMEOUT';
        reject(timeoutError);
      }, timeoutMs);
      promise
        .then((result) => {
          clearTimeout(timer);
          resolve(result);
        })
        .catch((error) => {
          clearTimeout(timer);
          reject(error);
        });
    });
  }

  private async delay(ms: number): Promise<void> {
    await new Promise<void>((resolve) => setTimeout(resolve, ms));
  }
}
