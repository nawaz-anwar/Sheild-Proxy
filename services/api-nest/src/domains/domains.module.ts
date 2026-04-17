import { Module } from '@nestjs/common';
import { DomainsController } from './domains.controller';
import { DomainsService } from './domains.service';
import { DatabaseService } from '../database/database.service';
import { RulesModule } from '../rules/rules.module';

@Module({
  imports: [RulesModule],
  controllers: [DomainsController],
  providers: [DomainsService, DatabaseService],
  exports: [DomainsService],
})
export class DomainsModule {}
