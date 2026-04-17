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

    if (rules.length > 0) {
      const values: string[] = [];
      const params: Array<string> = [];
      for (let i = 0; i < rules.length; i++) {
        const rule = rules[i];
        const base = i * 4;
        values.push(`($${base + 1}, $${base + 2}, $${base + 3}, $${base + 4})`);
        params.push(randomUUID(), domainId, rule.pathPrefix || '/', rule.action);
      }
      await this.db.query(`INSERT INTO proxy_rules (id, domain_id, path_prefix, action) VALUES ${values.join(', ')}`, params);
    }

    return { domainId, rulesCount: rules.length };
  }
}
