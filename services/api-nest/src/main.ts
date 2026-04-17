import 'reflect-metadata';
import { readFileSync } from 'fs';
import { NestFactory } from '@nestjs/core';
import { AppModule } from './app.module';

async function bootstrap() {
  const jwtSecret = readFromEnvOrFile('JWT_SECRET');
  if (!jwtSecret || jwtSecret === 'change-me') {
    throw new Error('JWT_SECRET must be configured before starting api-nest');
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
