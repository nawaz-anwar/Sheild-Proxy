import { Module } from '@nestjs/common';
import { DomainsController } from './domains.controller';
import { DomainsService } from './domains.service';
import { DatabaseService } from '../database/database.service';
import { RulesModule } from '../rules/rules.module';
import { DomainSyncService } from './domain-sync.service';

@Module({
  imports: [RulesModule],
  controllers: [DomainsController],
  providers: [DomainsService, DatabaseService, DomainSyncService],
  exports: [DomainsService],
})
export class DomainsModule {}
