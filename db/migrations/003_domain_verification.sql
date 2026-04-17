-- Add domain verification and connection fields
ALTER TABLE domains 
ADD COLUMN IF NOT EXISTS verification_token VARCHAR(255),
ADD COLUMN IF NOT EXISTS verified BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS proxy_connected BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS verification_method VARCHAR(50) DEFAULT 'dns_txt',
ADD COLUMN IF NOT EXISTS dns_records JSONB DEFAULT '{}',
ADD COLUMN IF NOT EXISTS last_verification_attempt TIMESTAMP,
ADD COLUMN IF NOT EXISTS verification_attempts INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS verified_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS connected_at TIMESTAMP;

-- Update status enum to include new states
-- status can be: pending_verification, verified, active, failed, suspended

-- Create index for faster lookups
CREATE INDEX IF NOT EXISTS idx_domains_verification_token ON domains(verification_token);
CREATE INDEX IF NOT EXISTS idx_domains_verified ON domains(verified);
CREATE INDEX IF NOT EXISTS idx_domains_proxy_connected ON domains(proxy_connected);
CREATE INDEX IF NOT EXISTS idx_domains_status ON domains(status);

-- Create domain verification logs table
CREATE TABLE IF NOT EXISTS domain_verification_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  domain_id UUID NOT NULL REFERENCES domains(id) ON DELETE CASCADE,
  verification_type VARCHAR(50) NOT NULL,
  status VARCHAR(50) NOT NULL,
  details JSONB,
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_verification_logs_domain_id ON domain_verification_logs(domain_id);
CREATE INDEX IF NOT EXISTS idx_verification_logs_created_at ON domain_verification_logs(created_at);
