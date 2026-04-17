import { Module } from '@nestjs/common';
import { AppController } from './app.controller';
import { DomainsModule } from './domains/domains.module';
import { DatabaseService } from './database/database.service';

@Module({
  imports: [DomainsModule],
  controllers: [AppController],
  providers: [DatabaseService],
})
export class AppModule {}
