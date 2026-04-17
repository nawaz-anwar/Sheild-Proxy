import { Injectable, OnModuleDestroy } from '@nestjs/common';
import { readFileSync } from 'fs';
import { Pool, QueryResultRow } from 'pg';

@Injectable()
export class DatabaseService implements OnModuleDestroy {
  private readonly pool: Pool;

  constructor() {
    const databaseDsn = this.readFromEnvOrFile('DATABASE_DSN');
    const postgresPassword = this.readFromEnvOrFile('POSTGRES_PASSWORD') ?? 'shield';

    this.pool = new Pool({
      ...(databaseDsn ? { connectionString: databaseDsn } : {}),
      host: process.env.POSTGRES_HOST ?? 'localhost',
      port: Number(process.env.POSTGRES_PORT ?? 5432),
      database: process.env.POSTGRES_DB ?? 'shield_proxy',
      user: process.env.POSTGRES_USER ?? 'shield',
      password: postgresPassword,
    });
  }

  query<T extends QueryResultRow = QueryResultRow>(text: string, params: unknown[] = []) {
    return this.pool.query<T>(text, params);
  }

  async onModuleDestroy() {
    await this.pool.end();
  }

  private readFromEnvOrFile(key: string): string | undefined {
    const direct = process.env[key];
    if (direct && direct.trim() !== '') {
      return direct.trim();
    }
    const filePath = process.env[`${key}_FILE`];
    if (!filePath) {
      return undefined;
    }
    return readFileSync(filePath, 'utf8').trim();
  }
}
