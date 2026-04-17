import { Module } from '@nestjs/common';
import { AppController } from './app.controller';
import { DomainsModule } from './domains/domains.module';
import { DatabaseService } from './database/database.service';
import { AuthModule } from './auth/auth.module';
import { AnalyticsModule } from './analytics/analytics.module';
import { RulesModule } from './rules/rules.module';

@Module({
  imports: [DomainsModule, AuthModule, RulesModule, AnalyticsModule],
  controllers: [AppController],
  providers: [DatabaseService],
})
export class AppModule {}
