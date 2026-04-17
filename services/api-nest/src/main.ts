import 'reflect-metadata';
import { readFileSync } from 'fs';
import { NestFactory } from '@nestjs/core';
import { AppModule } from './app.module';

async function bootstrap() {
  const jwtSecret = readFromEnvOrFile('JWT_SECRET');
  const normalized = jwtSecret?.toLowerCase().replace(/[_-]/g, '');
  if (!jwtSecret || normalized === 'changeme') {
    throw new Error('JWT_SECRET must be configured before starting api-nest');
  }
  if (jwtSecret.length < 32) {
    throw new Error('JWT_SECRET must be at least 32 characters');
  }
  const originSecret = readFromEnvOrFile('ORIGIN_SECRET');
  const mode = (process.env.ENV ?? process.env.MODE ?? '').trim().toLowerCase();
  if ((mode === 'prod' || mode === 'production') && originSecret && originSecret === jwtSecret) {
    throw new Error('ORIGIN_SECRET must be different from JWT_SECRET in production');
  }
  process.env.JWT_SECRET = jwtSecret;
  const app = await NestFactory.create(AppModule);
  await app.listen(Number(process.env.API_PORT ?? 3000));
}

bootstrap();

function readFromEnvOrFile(key: string): string | undefined {
  const direct = process.env[key];
  if (direct && direct.trim() !== '') {
    return direct.trim();
  }
  const filePath = process.env[`${key}_FILE`];
  if (!filePath) {
    return undefined;
  }
  return readFileSync(filePath, 'utf8').trim();
}
