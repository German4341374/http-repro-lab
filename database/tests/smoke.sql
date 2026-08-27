DO $$
DECLARE missing integer;
BEGIN
  SELECT count(*) INTO missing FROM (VALUES
    ('analysis_sessions'), ('sources'), ('requests'), ('responses'), ('findings'),
    ('sensitive_values'), ('reproductions'), ('comparisons'), ('repro_packs'),
    ('generated_clients'), ('audit_events')) AS expected(name)
  WHERE to_regclass('public.' || expected.name) IS NULL;
  IF missing <> 0 THEN RAISE EXCEPTION 'missing % required tables', missing; END IF;
END $$;

