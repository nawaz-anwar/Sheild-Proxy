import { Controller, Get, Header } from '@nestjs/common';
import { DomainsService } from './domains/domains.service';

@Controller()
export class AppController {
  constructor(private readonly domainsService: DomainsService) {}

  @Get('/healthz')
  health(): { status: string; phase: string } {
    return { status: 'ok', phase: 'mvp-core-proxy' };
  }

  @Get('/metrics')
  @Header('Content-Type', 'text/plain; version=0.0.4')
  async metrics(): Promise<string> {
    const count = await this.domainsService.countDomains();
    return `shield_api_domains_total ${count}\n`;
  }
}
