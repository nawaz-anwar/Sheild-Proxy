import 'reflect-metadata';
import { NestFactory } from '@nestjs/core';
import { AppModule } from './app.module';

async function bootstrap() {
  if (!process.env.JWT_SECRET || process.env.JWT_SECRET === 'change-me') {
    throw new Error('JWT_SECRET must be configured before starting api-nest');
  }
  const app = await NestFactory.create(AppModule);
  await app.listen(Number(process.env.API_PORT ?? 3000));
}

bootstrap();
