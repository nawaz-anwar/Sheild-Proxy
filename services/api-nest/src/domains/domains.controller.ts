import { Body, Controller, Get, Param, Post, Put } from '@nestjs/common';
import { DomainsService } from './domains.service';

@Controller('/domains')
export class DomainsController {
  constructor(private readonly domainsService: DomainsService) {}

  @Get()
  list() {
    return this.domainsService.list();
  }

  @Get('/:id')
  getById(@Param('id') id: string) {
    return this.domainsService.getById(id);
  }

  @Post('/register')
  register(@Body() body: { clientName: string; domain: string; upstreamUrl: string }) {
    return this.domainsService.register(body.clientName, body.domain, body.upstreamUrl);
  }

  @Get('/:id/status')
  status(@Param('id') id: string) {
    return this.domainsService.status(id);
  }

  @Post('/:id/initiate-verification')
  initiateVerification(@Param('id') id: string) {
    return this.domainsService.initiateVerification(id);
  }

  @Post('/:id/verify-dns')
  verifyDns(@Param('id') id: string) {
    return this.domainsService.verifyDns(id);
  }

  @Post('/:id/check-connection')
  checkConnection(@Param('id') id: string) {
    return this.domainsService.checkConnection(id);
  }

  @Get('/:id/verification-logs')
  getVerificationLogs(@Param('id') id: string) {
    return this.domainsService.getVerificationLogs(id);
  }

  @Put('/:id/rules')
  updateRules(
    @Param('id') id: string,
    @Body() body: { rules: Array<{ pathPrefix: string; action: 'proxy' | 'block' | 'challenge' }> },
  ) {
    return this.domainsService.updateRules(id, body.rules ?? []);
  }
}
