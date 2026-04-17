import { Injectable, UnauthorizedException } from '@nestjs/common';
import { createHash, createHmac, randomUUID, timingSafeEqual } from 'crypto';
import { DatabaseService } from '../database/database.service';

type AuthTokenPayload = {
  sub: string;
  email: string;
  iat: number;
  exp: number;
  iss: string;
  aud: string;
};

@Injectable()
export class AuthService {
  constructor(private readonly db: DatabaseService) {}

  async register(name: string, email: string, password: string) {
    const id = randomUUID();
    const passwordHash = this.hash(password);
    const result = await this.db.query<{ id: string; email: string }>(
      `INSERT INTO clients (id, name, email, password_hash)
       VALUES ($1, $2, $3, $4)
       ON CONFLICT (email) DO UPDATE SET name = EXCLUDED.name, password_hash = EXCLUDED.password_hash
       RETURNING id, email`,
      [id, name, email.toLowerCase(), passwordHash],
    );
    const client = result.rows[0];
    const token = this.signToken({ sub: client.id, email: client.email });
    return { clientId: client.id, token };
  }

  async login(email: string, password: string) {
    const result = await this.db.query<{ id: string; email: string; password_hash: string | null }>(
      'SELECT id, email, password_hash FROM clients WHERE email = $1',
      [email.toLowerCase()],
    );
    if (result.rowCount === 0) {
      throw new UnauthorizedException('invalid credentials');
    }
    const row = result.rows[0];
    const expected = this.hash(password);
    if (
      !row.password_hash ||
      row.password_hash.length !== expected.length ||
      !timingSafeEqual(Buffer.from(row.password_hash), Buffer.from(expected))
    ) {
      throw new UnauthorizedException('invalid credentials');
    }
    const token = this.signToken({ sub: row.id, email: row.email });
    return { clientId: row.id, token };
  }

  verifyToken(token: string): AuthTokenPayload {
    const [header, payload, signature] = token.split('.');
    if (!header || !payload || !signature) {
      throw new UnauthorizedException('invalid token format');
    }
    const expected = this.sign(`${header}.${payload}`);
    if (signature.length !== expected.length || !timingSafeEqual(Buffer.from(signature), Buffer.from(expected))) {
      throw new UnauthorizedException('invalid token signature');
    }
    const data = JSON.parse(Buffer.from(payload, 'base64url').toString('utf8')) as AuthTokenPayload;
    if (Date.now() >= data.exp * 1000) {
      throw new UnauthorizedException('token expired');
    }
    return data;
  }

  private hash(value: string): string {
    return createHash('sha256').update(value).digest('hex');
  }

  private signToken(input: { sub: string; email: string }): string {
    const issuer = process.env.JWT_ISSUER ?? 'shield-api';
    const audience = process.env.JWT_AUDIENCE ?? 'shield-dashboard';
    const ttlSeconds = Number(process.env.JWT_TTL_SECONDS ?? 3600);
    const now = Math.floor(Date.now() / 1000);
    const payload: AuthTokenPayload = {
      sub: input.sub,
      email: input.email,
      iat: now,
      exp: now + ttlSeconds,
      iss: issuer,
      aud: audience,
    };
    const header = { alg: 'HS256', typ: 'JWT' };
    const encodedHeader = Buffer.from(JSON.stringify(header)).toString('base64url');
    const encodedPayload = Buffer.from(JSON.stringify(payload)).toString('base64url');
    const unsigned = `${encodedHeader}.${encodedPayload}`;
    const signature = this.sign(unsigned);
    return `${unsigned}.${signature}`;
  }

  private sign(unsignedToken: string): string {
    const secret = process.env.JWT_SECRET ?? 'change-me';
    return createHmac('sha256', secret).update(unsignedToken).digest('base64url');
  }
}
