DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_available_extensions WHERE name = 'timescaledb') THEN
    CREATE EXTENSION IF NOT EXISTS timescaledb;

    -- NOTE: migrate_data => TRUE rewrites existing request_logs rows.
    -- For datasets >= 1,000,000 rows, run this during a maintenance window or use a
    -- dual-write/backfill rollout procedure before cutover to avoid extended lock time.
    PERFORM create_hypertable('request_logs', 'created_at', if_not_exists => TRUE, migrate_data => TRUE);

    ALTER TABLE request_logs
      SET (
        timescaledb.compress,
        timescaledb.compress_segmentby = 'domain_id,host',
        timescaledb.compress_orderby = 'created_at DESC'
      );

    PERFORM add_compression_policy('request_logs', INTERVAL '7 days', if_not_exists => TRUE);
    PERFORM add_retention_policy('request_logs', INTERVAL '90 days', if_not_exists => TRUE);
  ELSE
    RAISE NOTICE 'timescaledb extension is unavailable; skipping hypertable/compression/retention setup';
  END IF;
END $$;
