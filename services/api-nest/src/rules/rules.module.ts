import { Module } from '@nestjs/common';
import { RulesService } from './rules.service';
import { DatabaseService } from '../database/database.service';

@Module({
  providers: [RulesService, DatabaseService],
  exports: [RulesService],
})
export class RulesModule {}
