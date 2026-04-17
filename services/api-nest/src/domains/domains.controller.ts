import { Body, Controller, Get, Param, Post } from '@nestjs/common';
import { DomainsService } from './domains.service';

@Controller('/domains')
export class DomainsController {
  constructor(private readonly domainsService: DomainsService) {}

  @Post('/register')
  register(@Body() body: { clientName: string; domain: string; upstreamUrl: string }) {
    return this.domainsService.register(body.clientName, body.domain, body.upstreamUrl);
  }

  @Get('/:id/status')
  status(@Param('id') id: string) {
    return this.domainsService.status(id);
  }

  @Post('/:id/verify-dns')
  verifyDns(@Param('id') id: string) {
    return this.domainsService.verifyDns(id);
  }
}
