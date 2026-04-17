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

  constructor(private readonly db: DatabaseService) {}

  generateVerificationToken(): string {
    return `${this.VERIFICATION_PREFIX}${randomUUID()}`;
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
      txtRecordValue: token,
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
    expectedToken: string
  ): Promise<DnsVerificationResult> {
    const txtRecordName = this.getTxtRecordName(domain);

    try {
      this.logger.log(`Checking TXT record for ${txtRecordName}`);
      
      const records = await dns.resolveTxt(txtRecordName);
      const flatRecords = records.map((r) => r.join(''));

      this.logger.log(`Found TXT records: ${JSON.stringify(flatRecords)}`);

      const hasValidToken = flatRecords.some((record) => record === expectedToken);

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
      // Check CNAME record
      try {
        const cnameRecords = await dns.resolveCname(domain);
        const hasProxyCname = cnameRecords.some((record) =>
          record.includes('proxy.shieldproxy.com')
        );

        if (hasProxyCname) {
          await this.updateProxyConnection(domainId, true, 'cname', cnameRecords);
          return {
            connected: true,
            method: 'cname',
            records: cnameRecords,
            message: 'Domain connected via CNAME',
          };
        }
      } catch (cnameError: any) {
        // CNAME not found, try A record
        if (cnameError.code === 'ENODATA' || cnameError.code === 'ENOTFOUND') {
          // Continue to A record check
        } else {
          throw cnameError;
        }
      }

      // Check A record
      const aRecords = await dns.resolve4(domain);
      // TODO: Replace with actual proxy server IP
      const PROXY_SERVER_IP = process.env.PROXY_SERVER_IP || '0.0.0.0';
      
      const hasProxyARecord = aRecords.some((record) => record === PROXY_SERVER_IP);

      if (hasProxyARecord) {
        await this.updateProxyConnection(domainId, true, 'a_record', aRecords);
        return {
          connected: true,
          method: 'a_record',
          records: aRecords,
          message: 'Domain connected via A record',
        };
      }

      return {
        connected: false,
        message: 'Domain not pointing to ShieldProxy servers',
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
}
