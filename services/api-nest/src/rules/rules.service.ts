import { Injectable, NotFoundException } from '@nestjs/common';
import { randomUUID } from 'crypto';
import { DatabaseService } from '../database/database.service';

export type RuleInput = {
  pathPrefix: string;
  action: 'proxy' | 'block' | 'challenge';
};

@Injectable()
export class RulesService {
  constructor(private readonly db: DatabaseService) {}

  async replaceDomainRules(domainId: string, rules: RuleInput[]) {
    const domain = await this.db.query<{ id: string }>('SELECT id FROM domains WHERE id = $1', [domainId]);
    if (domain.rowCount === 0) {
      throw new NotFoundException('domain not found');
    }

    await this.db.query('DELETE FROM proxy_rules WHERE domain_id = $1', [domainId]);

    for (const rule of rules) {
      await this.db.query(
        'INSERT INTO proxy_rules (id, domain_id, path_prefix, action) VALUES ($1, $2, $3, $4)',
        [randomUUID(), domainId, rule.pathPrefix || '/', rule.action],
      );
    }

    return { domainId, rulesCount: rules.length };
  }
}
