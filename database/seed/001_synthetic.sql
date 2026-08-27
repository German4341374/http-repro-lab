INSERT INTO analysis_sessions (source_sha256, status, completed_at, metadata)
VALUES (repeat('a', 64), 'complete', now(), '{"fixture":"synthetic"}')
ON CONFLICT DO NOTHING;

