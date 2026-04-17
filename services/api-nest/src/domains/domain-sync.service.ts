import { Injectable, OnModuleDestroy } from '@nestjs/common';
import { createClient, RedisClientType } from 'redis';

@Injectable()
export class DomainSyncService implements OnModuleDestroy {
  private readonly prefix = (process.env.REDIS_KEY_PREFIX ?? 'shield').trim() || 'shield';
  private readonly channel = `${this.prefix}:domain:sync`;
  private readonly client: RedisClientType | null;
  private readonly connectPromise: Promise<void> | null;

  constructor() {
    const addr = (process.env.REDIS_ADDR ?? '').trim();
    if (!addr) {
      this.client = null;
      this.connectPromise = null;
      return;
    }
    const url = addr.includes('://') ? addr : `redis://${addr}`;
    this.client = createClient({ url });
    this.connectPromise = this.client.connect().then(() => undefined).catch(() => undefined);
  }

  async publishDomainSync(host: string) {
    const normalizedHost = host.trim().toLowerCase();
    if (!normalizedHost || !this.client) {
      return;
    }
    await this.connectPromise;
    await this.client.publish(this.channel, normalizedHost).catch(() => undefined);
  }

  async onModuleDestroy() {
    if (!this.client) {
      return;
    }
    await this.client.quit().catch(() => undefined);
  }
}
