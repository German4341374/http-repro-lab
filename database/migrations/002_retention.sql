BEGIN;
CREATE OR REPLACE FUNCTION delete_expired_sessions(retain_for interval DEFAULT interval '30 days')
RETURNS bigint LANGUAGE plpgsql AS $$
DECLARE deleted_count bigint;
BEGIN
  DELETE FROM analysis_sessions WHERE created_at < now() - retain_for;
  GET DIAGNOSTICS deleted_count = ROW_COUNT;
  RETURN deleted_count;
END;
$$;
COMMIT;

