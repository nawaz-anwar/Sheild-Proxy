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
