-- Drop triggers
DROP TRIGGER IF EXISTS prevent_self_connection;
DROP TRIGGER IF EXISTS update_connections_updated_at;

-- Drop indexes
DROP INDEX IF EXISTS idx_connections_from_to;
DROP INDEX IF EXISTS idx_connections_updated_at;
DROP INDEX IF EXISTS idx_connections_created_at;
DROP INDEX IF EXISTS idx_connections_strength;
DROP INDEX IF EXISTS idx_connections_type;
DROP INDEX IF EXISTS idx_connections_to_note_id;
DROP INDEX IF EXISTS idx_connections_from_note_id;
DROP INDEX IF EXISTS idx_connections_unique;

-- Drop connections table
DROP TABLE IF EXISTS connections;